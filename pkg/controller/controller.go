/*
Copyright 2018 The Kubegene Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	clientset "k8s.io/client-go/kubernetes"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	geneclientset "kubegene.io/kubegene/pkg/client/clientset/versioned/typed/gene/v1alpha1"
	geneinformers "kubegene.io/kubegene/pkg/client/informers/externalversions/gene/v1alpha1"
	genelisters "kubegene.io/kubegene/pkg/client/listers/gene/v1alpha1"
	"kubegene.io/kubegene/pkg/util"
	"kubegene.io/kubegene/pkg/version"
)

// controllerKind contains the schema.GroupVersionKind for this controller type.
var execKind = genev1alpha1.SchemeGroupVersion.WithKind("Execution")

// ControllerParameters contains arguments for creation of a new ExecutionController.
type ControllerParameters struct {
	EventRecorder     record.EventRecorder
	KubeClient        clientset.Interface
	ExecutionClient   geneclientset.ExecutionsGetter
	JobInformer       batchinformers.JobInformer
	ExecutionInformer geneinformers.ExecutionInformer
}

type ExecutionController struct {
	// eventRecorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	eventRecorder record.EventRecorder

	// kubeclientset is a standard kubernetes clientset
	kubeClient clientset.Interface
	// execClient is a client for execution.kubegene.io group
	execClient geneclientset.ExecutionsGetter

	execLister genelisters.ExecutionLister
	execSynced cache.InformerSynced

	jobLister batchv1listers.JobLister
	jobSynced cache.InformerSynced

	execQueue  workqueue.RateLimitingInterface
	jobQueue   workqueue.RateLimitingInterface
	eventQueue workqueue.RateLimitingInterface

	syncExecHandler func(execKey string) error
	syncJobHandler  func(jobKey string) (bool, error)

	execGraphBuilder *GraphBuilder

	execJobController *ExecutionJobController

	execStatusUpdater ExecutionStatusUpdater
	execUpdater       ExecutionUpdater
}

func NewExecutionController(p *ControllerParameters) *ExecutionController {
	controller := &ExecutionController{
		eventRecorder: p.EventRecorder,
		kubeClient:    p.KubeClient,
		execClient:    p.ExecutionClient,
		execLister:    p.ExecutionInformer.Lister(),
		execSynced:    p.ExecutionInformer.Informer().HasSynced,
		jobLister:     p.JobInformer.Lister(),
		jobSynced:     p.JobInformer.Informer().HasSynced,
		execQueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "execution"),
		jobQueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "execution-job"),
		eventQueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "job-event"),
	}

	p.ExecutionInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { controller.enqueueObj(controller.execQueue, obj) },
			DeleteFunc: func(obj interface{}) { controller.enqueueObj(controller.execQueue, obj) },
		},
	)

	p.JobInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.addJob,
			UpdateFunc: controller.updateJob,
			DeleteFunc: controller.deleteJob,
		},
	)

	controller.syncJobHandler = controller.syncJob
	controller.syncExecHandler = controller.syncExecution
	controller.execGraphBuilder = NewGraphBuilder()
	controller.execUpdater = NewExecutionUpdater(p.ExecutionClient)
	controller.execJobController = NewExecutionJobController(p.KubeClient, controller.jobLister, controller.execLister,
		controller.eventQueue, controller.execGraphBuilder, controller.execUpdater)
	controller.execStatusUpdater = NewExecutionStatusUpdater(p.ExecutionClient)

	return controller
}

// Run the main goroutine responsible for watching and syncing executions.
func (c *ExecutionController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.execQueue.ShutDown()
	defer c.jobQueue.ShutDown()

	glog.Infof("Starting execution controller with version %s", version.GetVersion())
	defer glog.Infof("Shutting down execution controller")

	if !cache.WaitForCacheSync(stopCh, c.execSynced, c.jobSynced) {
		glog.Errorf("Cannot sync caches")
		return
	}

	// start asynchronous go routine processing event queue.
	go c.execJobController.Run(workers, stopCh)

	for i := 0; i < workers; i++ {
		go wait.Until(c.execWorker, time.Second, stopCh)
		go wait.Until(c.jobWorker, time.Second, stopCh)
	}

	<-stopCh
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *ExecutionController) execWorker() {
	for c.processNextExecItem() {
	}
}

func (c *ExecutionController) processNextExecItem() bool {
	key, quit := c.execQueue.Get()
	if quit {
		return false
	}
	defer c.execQueue.Done(key)

	glog.Infof("execution %v enter sync", key)
	err := c.syncExecHandler(key.(string))
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error syncing execution: %v", err))
	}

	return true
}

// jobWorker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *ExecutionController) jobWorker() {
	for c.processNextJobItem() {
	}
}

func (c *ExecutionController) processNextJobItem() bool {
	key, quit := c.jobQueue.Get()
	if quit {
		return false
	}
	defer c.jobQueue.Done(key)

	forget, err := c.syncJobHandler(key.(string))
	if err == nil {
		if forget {
			c.jobQueue.Forget(key)
		}
		return true
	}

	utilruntime.HandleError(fmt.Errorf("error syncing job: %v", err))
	c.jobQueue.AddRateLimited(key)

	return true
}

func (c *ExecutionController) syncJob(key string) (bool, error) {
	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing job %q (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return false, err
	}
	if len(ns) == 0 || len(name) == 0 {
		return false, fmt.Errorf("invalid job key %q: either namespace or name is missing", key)
	}

	sharedJob, err := c.jobLister.Jobs(ns).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.V(4).Infof("Job has been deleted: %v", key)
			return true, nil
		}
		return false, err
	}
	job := sharedJob.DeepCopy()

	controllerRef := metav1.GetControllerOf(job)
	// This should never happen because we do the check when we queue the job.
	// Unless someone remove the OwnerReference on the job.
	if controllerRef == nil {
		glog.Infof("can not find ownerReference for job %s", key)
		return true, nil
	}

	sharedExec := c.resolveControllerRef(job.Namespace, controllerRef)
	if sharedExec == nil {
		// Ignore jobs unrelated to execution.
		glog.Infof("job %s does not belongs to execution", key)
		return true, nil
	}
	exec := sharedExec.DeepCopy()

	// The execution has been marked as completed, just return.
	if util.IsExecutionCompleted(exec) {
		return true, nil
	}

	graph := c.execGraphBuilder.GetGraph(util.KeyOf(exec))
	if graph == nil {
		// The execution has been running but the graph has been deleted.
		util.MarkExecutionError(exec, fmt.Errorf("graph of execution %s do not exist", util.KeyOf(exec)))

		// Ask api server to update etcd data.
		// if update error, we will retry it next sync.
		if err = c.execStatusUpdater.UpdateExecutionStatus(exec, sharedExec); err != nil {
			glog.V(3).Infof("update execution %s status error: %#v", key, err)
			return false, err
		}

		glog.Infof("graph of execution %s do not exist", util.KeyOf(exec))
		return true, nil
	}

	// find the vertex in the graph.
	vertex := graph.FindVertexByName(job.Name)
	if vertex == nil {
		util.MarkExecutionError(exec, fmt.Errorf(missVertexMessage))
		// Ask api server to update etcd data.
		// if update error, we will retry it next sync.
		if err = c.execStatusUpdater.UpdateExecutionStatus(exec, sharedExec); err != nil {
			glog.V(3).Infof("update execution %s status error: %#v", key, err)
			return false, err
		}
		return true, nil
	}

	// get job condition
	jobConditionType, message := util.GetJobCondition(job)

	// in case missing the running event, following status set will cause panic
	if util.GetVertexStatus(exec, job.Name) == nil {
		vertexStatus := util.InitializeVertexStatus(vertex.Data.Job.Name, genev1alpha1.VertexRunning, vertexRunningMessage, vertex.Children)
		if exec.Status.Vertices == nil {
			exec.Status.Vertices = make(map[string]genev1alpha1.VertexStatus)
		}
		exec.Status.Vertices[vertexStatus.ID] = vertexStatus
	}

	switch jobConditionType {
	case batch.JobFailed:
		// Job is failed, mark the vertex as failed.
		util.MarkVertexFailed(exec, job.Name, message)

		// Job is failed, the execution is marked failed and will not retry.
		util.MarkExecutionFailed(exec, message)

		// Ask api server to update etcd data.
		if err = c.execStatusUpdater.UpdateExecutionStatus(exec, sharedExec); err != nil {
			glog.V(3).Infof("update execution %s status error: %#v", key, err)
			return false, err
		}

		return true, nil

	case batch.JobComplete:
		// the vertex has been finished.
		vertex.Data.Finished = true
		// Mark the vertex as success.
		if len(message) == 0 {
			message = "success"
		}
		util.MarkVertexSuccess(exec, job.Name, message)
		// The number of successful vertex plus 1.
		graph.PlusNumOfSuccess()
		if graph.NumOfSuccess == graph.VertexCount {
			// All of the vertex has been successful, then mark the execution as successful.
			util.MarkExecutionSuccess(exec, executionSuccessMessage)
		}

		// Ask api server to update etcd data.
		if err = c.execStatusUpdater.UpdateExecutionStatus(exec, sharedExec); err != nil {
			glog.V(3).Infof("update execution %s status error: %#v", key, err)
			return false, err
		}

		if graph.NumOfSuccess != graph.VertexCount {
			// add execution to event queue to trigger running.
			event := Event{Type: JobsAfter, Name: job.Name, Key: util.KeyOf(exec)}
			c.eventQueue.Add(event)
		}

		return true, nil

	default:
		glog.V(4).Infof("Job is running and has not update its condition: %v", key)

		// usually a add event can approach here and mark the vertex as running.
		if util.GetVertexStatus(exec, job.Name) == nil {
			vertexStatus := util.InitializeVertexStatus(vertex.Data.Job.Name, genev1alpha1.VertexRunning, vertexRunningMessage, vertex.Children)
			if exec.Status.Vertices == nil {
				exec.Status.Vertices = make(map[string]genev1alpha1.VertexStatus)
			}
			exec.Status.Vertices[vertexStatus.ID] = vertexStatus
		}
		// If the phase of execution is unset, and we have watch job that belongs to
		// this execution, then mark the phase of execution running.
		if len(exec.Status.Phase) == 0 {
			util.MarkExecutionRunning(exec, executionRunningMessage)
		}
		// Ask api server to update etcd data.
		if err = c.execStatusUpdater.UpdateExecutionStatus(exec, sharedExec); err != nil {
			glog.V(3).Infof("update execution %s status error: %#v", util.KeyOf(exec), err)
			return false, err
		}

		return true, nil
	}
}

// syncExecution will sync the execution with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *ExecutionController) syncExecution(key string) error {
	startTime := time.Now()
	glog.V(3).Infof("Started syncing execution %q (%v)", key, startTime)
	defer func() {
		glog.V(3).Infof("Finished syncing execution %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	execution, err := c.execLister.Executions(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("execution %v has been deleted", key)
		c.execGraphBuilder.DeleteGraph(key)
		return nil
	}
	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	exec := execution.DeepCopy()

	if err := ValidateExecution(exec); err != nil {
		util.MarkExecutionError(exec, err)
		c.execStatusUpdater.UpdateExecutionStatus(exec, execution)
		return err
	}

	graph := c.execGraphBuilder.GetGraph(key)
	if graph == nil {
		glog.V(2).Infof("generate graph for execution %v", key)
		c.execGraphBuilder.AddGraph(exec)
	}

	// add execution to event queue to trigger running
	event := Event{Type: NewAdded, Key: util.KeyOf(exec)}
	c.eventQueue.Add(event)

	return nil
}

// enqueueObj adds execution or job to given work queue.
func (c *ExecutionController) enqueueObj(queue workqueue.Interface, obj interface{}) {
	// Beware of "xxx deleted" events
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}

	objName := util.KeyOf(obj)

	glog.V(5).Infof("enqueued %q for sync", objName)
	queue.Add(objName)
}

func (c *ExecutionController) addJob(obj interface{}) {
	job := obj.(*batch.Job)
	if job.DeletionTimestamp != nil {
		// on a restart of the controller controller, it's possible a new pod shows up in a state that
		// is already pending deletion. Prevent the pod from being a creation observation.
		c.deleteJob(job)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(job); controllerRef != nil {
		exec := c.resolveControllerRef(job.Namespace, controllerRef)
		if exec == nil {
			return
		}

		c.enqueueObj(c.jobQueue, job)
		return
	}

}

func (c *ExecutionController) deleteJob(obj interface{}) {
	job, ok := obj.(*batch.Job)

	// When a delete is dropped, the relist will notice a job in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the job
	// changed labels the new exec will not be woken up till the periodic resync.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %+v", obj))
			return
		}
		job, ok = tombstone.Obj.(*batch.Job)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a job %+v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(job)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	exec := c.resolveControllerRef(job.Namespace, controllerRef)
	if exec == nil {
		return
	}

	c.enqueueObj(c.jobQueue, job)
}

func (c *ExecutionController) updateJob(old, cur interface{}) {
	curJob := cur.(*batch.Job)
	oldJob := old.(*batch.Job)
	if curJob.ResourceVersion == oldJob.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same job will always have different RVs.
		return
	}
	if curJob.DeletionTimestamp != nil {
		c.deleteJob(curJob)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(curJob); controllerRef != nil {
		exec := c.resolveControllerRef(curJob.Namespace, controllerRef)
		if exec == nil {
			return
		}

		c.enqueueObj(c.jobQueue, curJob)
		return
	}
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *ExecutionController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *genev1alpha1.Execution {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != execKind.Kind {
		return nil
	}
	exec, err := c.execLister.Executions(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if exec.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return exec
}
