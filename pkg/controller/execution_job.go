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
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	genelisters "kubegene.io/kubegene/pkg/client/listers/gene/v1alpha1"
	"kubegene.io/kubegene/pkg/util"
)

type EventType string

const (
	// maxRetries is the number of times a execution will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
	// newly start execute execution jobs
	NewAdded EventType = "NewAdded"
	// start execute jobs that depend on a job
	JobsAfter EventType = "JobsAfter"
)

var ExceedParallelismError = fmt.Errorf("running jobs have reached the execution parallelism limit")

type Event struct {
	Type EventType
	// job name, if Type is `NewAdded`, it can be not specified.
	Name string
	// Execution key
	Key string
}

type ExecutionJobController struct {
	kubeClient       clientset.Interface
	jobLister        batchv1listers.JobLister
	executionLister  genelisters.ExecutionLister
	queue            workqueue.RateLimitingInterface
	execGraphBuilder *GraphBuilder
}

func NewExecutionJobController(
	kubeClient clientset.Interface,
	jobLister batchv1listers.JobLister,
	executionLister genelisters.ExecutionLister,
	eventQueue workqueue.RateLimitingInterface,
	execGraphBuilder *GraphBuilder,
) *ExecutionJobController {
	return &ExecutionJobController{
		queue:            eventQueue,
		kubeClient:       kubeClient,
		jobLister:        jobLister,
		executionLister:  executionLister,
		execGraphBuilder: execGraphBuilder,
	}
}

func (e *ExecutionJobController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer e.queue.ShutDown()

	glog.Infof("Starting execution job controller")
	defer glog.Infof("Shutting down execution job controller")

	for i := 0; i < workers; i++ {
		go wait.Until(e.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (e *ExecutionJobController) worker() {
	for e.processNextWorkItem() {

	}
}

func (e *ExecutionJobController) processNextWorkItem() bool {
	item, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(item)

	event := item.(Event)
	err := e.syncHandler(event)
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your key.  This will
		// reset things like failure counts for per-item rate limiting
		e.queue.Forget(item)
		return true
	}

	if err == ExceedParallelismError {
		glog.V(2).Infof("Parallel running jobs exceeded the execution limit, retry after 10s")
		// TODO: make this configurable
		e.queue.AddAfter(item, 10*time.Second)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", event, err))
	// since we failed, we should requeue the item to work on later.  This method will add a backoff
	// to avoid hotlooping on particular items (they're probably still not going to work right away)
	// and overall controller protection (everything I've done is broken, this controller needs to
	// calm down or it can starve other useful work) cases.
	e.queue.AddRateLimited(item)
	return true
}

func (e *ExecutionJobController) syncHandler(event Event) error {
	graph := e.execGraphBuilder.GetGraph(event.Key)
	if graph == nil {
		glog.V(2).Infof("graph of execution %s does not exist", event.Key)
		return nil
	}

	switch event.Type {
	case NewAdded:
		glog.V(2).Infof("execution %v start running.", event.Key)

		rootVertexs := graph.GetRootVertex()
		for _, rootVertex := range rootVertexs {
			if e.shouldStartJob(event.Key, rootVertex.Data.Job) {
				// root vertex, create job
				if err := e.createJob(rootVertex.Data.Job); err != nil {
					key := util.KeyOf(rootVertex.Data.Job)
					return fmt.Errorf("create job %s error: %v", key, err)
				}
			} else {
				return ExceedParallelismError
			}
		}
	case JobsAfter:
		glog.V(2).Infof("job %v has run successfully.", event.Name)

		vertex := graph.FindVertexByName(event.Name)
		for _, child := range vertex.Children {
			allDependentsFinished := true

			dependents := graph.FindDependentsByName(child.Data.Job.Name)
			for _, dependent := range dependents {
				childVertex := graph.FindVertex(dependent)
				if !childVertex.Data.Finished {
					allDependentsFinished = false
					break
				}
			}
			if allDependentsFinished {
				glog.V(2).Infof("all dependent of job %v has run successfully, start running.", child.Data.Job.Name)

				if e.shouldStartJob(event.Key, child.Data.Job) {
					if err := e.createJob(child.Data.Job); err != nil {
						key := util.KeyOf(child.Data.Job)
						return fmt.Errorf("create job %s error: %v", key, err)
					}
				} else {
					return ExceedParallelismError
				}
			}
		}
	}
	return nil
}

func (e *ExecutionJobController) createJob(job *batch.Job) error {
	_, err := e.jobLister.Jobs(job.Namespace).Get(job.Name)
	// job has been already created
	if !errors.IsNotFound(err) {
		return nil
	}

	_, err = e.kubeClient.BatchV1().Jobs(job.Namespace).Create(job)
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

func (e *ExecutionJobController) handleErr(err error, event Event) {
	if err == nil {
		e.queue.Forget(event)
		return
	}

	if e.queue.NumRequeues(event) < maxRetries {
		glog.V(2).Infof("Error syncing execution jobs %q , retrying. Error: %v", event.Key, err)
		e.queue.AddRateLimited(event)
		return
	}

	glog.Warningf("Dropping event %v out of the queue: %v", event, err)
	e.queue.Forget(event)
	utilruntime.HandleError(err)
}

// getActiveJobsForExecution returns the set of running jobs that this Execution should manage.
// Note that the returned Pods are pointers into the cache.
func (e *ExecutionJobController) getActiveJobsForExecution(namespace string, selector labels.Selector) ([]*batch.Job, error) {
	// List all pods to include those that don't match the selector anymore
	// but have a ControllerRef pointing to this controller.
	jobs, err := e.jobLister.Jobs(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	var result []*batch.Job
	for _, job := range jobs {
		if !util.IsJobFinished(job) {
			result = append(result, job)
		}
	}
	return result, nil
}

func (e *ExecutionJobController) shouldStartJob(key string, job *batch.Job) bool {
	namespace, name, _ := cache.SplitMetaNamespaceKey(key)
	execution, err := e.executionLister.Executions(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("Execution %v has been deleted", key)
		return false
	}
	if err != nil {
		glog.Errorf("Get execution %s error: %v", key, err)
		return false
	}
	if execution.Spec.Parallelism != nil {
		jobs, err := e.getActiveJobsForExecution(job.Namespace, labels.Set(job.Labels).AsSelector())
		if err != nil {
			glog.Errorf("Get active jobs for execution %s error: %v", key, err)
			return false
		}

		if len(jobs) >= int(*execution.Spec.Parallelism) {
			return false
		}
	}

	return true
}
