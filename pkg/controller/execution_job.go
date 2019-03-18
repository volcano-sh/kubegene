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
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"kubegene.io/kubegene/cmd/genectl/parser"
	genelisters "kubegene.io/kubegene/pkg/client/listers/gene/v1alpha1"
	"kubegene.io/kubegene/pkg/graph"
	"kubegene.io/kubegene/pkg/util"

	"strconv"
	"strings"
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
	execUpdater      ExecutionUpdater
}

func NewExecutionJobController(
	kubeClient clientset.Interface,
	jobLister batchv1listers.JobLister,
	executionLister genelisters.ExecutionLister,
	eventQueue workqueue.RateLimitingInterface,
	execGraphBuilder *GraphBuilder,
	execUpdater ExecutionUpdater,
) *ExecutionJobController {
	return &ExecutionJobController{
		queue:            eventQueue,
		kubeClient:       kubeClient,
		jobLister:        jobLister,
		executionLister:  executionLister,
		execGraphBuilder: execGraphBuilder,
		execUpdater:      execUpdater,
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

				/*start changes related to dynamicconcurrency / get_result */
				if child.Data.DynamicJob != nil {
					// get the result of the dependent job using getJobResult()
					result, err := e.getJobResult(vertex.Data.Job)
					if err != nil {
						glog.Infof("getJobResult failed %v", err)
						return fmt.Errorf("getJobResult failed : %v", err)
					}

					// construct the dynamic job based on result
					err1 := e.createDynamicJob(child, result, graph, event.Key)
					if err1 != nil {
						glog.Infof("createDynamicJob failed %v", err1)
						return fmt.Errorf("createDynamicJob failed : %v", err1)
					}
					// update the graph
					// update the execution e.execUpdater.UpdateExecution(exec, execution)
					return nil

				}
				/*end changes related to dynamic concurrency or get_result*/
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

func evalgetResult(output string, vars []interface{}) ([]parser.Var, string, error) {

	result := make([]parser.Var, 0, len(vars))
	var parentJobName string
	for _, vs := range vars {
		v := vs.([]interface{})

		if item, ok := v[0].(string); ok && item == "get_result" {

			parentJobName := v[1].(string)
			sep := v[2].(string)

			if sep == "" {
				temp := []interface{}{output}
				result = append(result, temp)
			} else {
				var strslice []interface{}
				for _, str := range strings.Split(output, sep) {
					if str != "" {
						strslice = append(strslice, str)
					}
				}
				temp := []interface{}{strslice}
				result = append(result, temp)
			}
			glog.Infof("In evalgetResult jobName: %s sep:%s", parentJobName, sep)
		}
	}
	glog.Infof("In evalgetResult result: %v", result)
	return result, parentJobName, nil
}
func ConvertVars(vars []interface{}) []parser.Var {
	result := make([]parser.Var, 0, len(vars))

	for _, v := range vars {
		res := make([]interface{}, 0)
		res = append(res, v)
		result = append(result, res)
	}
	return result
}

func (e *ExecutionJobController) createDynamicJob(child *graph.Vertex, jobresult string, graph1 *graph.Graph, key string) error {

	job := child.Data.Job
	controllerRef := metav1.GetControllerOf(job)

	if controllerRef == nil {
		err := fmt.Errorf("controllerRef is null")
		glog.Infof("can not find ownerReference for job %s", job.Name)
		return err
	}

	task := child.Data.DynamicJob.DeepCopy()
	varsIter, parentJobName, err := evalgetResult(jobresult, child.Data.DynamicJob.CommandsIter.VarsIter)
	if err != nil {
		glog.V(2).Infof("Error in evalgetResult execution job name %q , . Error: %v", job.Name, err)
		return err
	}
	glog.V(2).Infof("evalgetResult output parentJobName:%s , varsIter: %v", parentJobName, varsIter)

	newCommands := child.Data.DynamicJob.CommandSet

	vars := ConvertVars(child.Data.DynamicJob.CommandsIter.Vars)

	// convert varsIter to var
	iterVars := parser.VarIter2Vars(varsIter)

	// merge vars
	vars = append(vars, iterVars...)

	command := child.Data.DynamicJob.CommandsIter.Command

	// generate all commands.
	iterCommands := parser.Iter2Array(command, vars)

	// merge jobInfo.commands and jobInfo.iterCommands
	newCommands = append(newCommands, iterCommands...)

	//now need to update the task based on the new parameters

	task.CommandSet = newCommands

	glog.V(2).Infof("final commandset task.CommandSet %v ", task.CommandSet)

	//update this task in the execution struct then call the Update()
	namespace, name, _ := cache.SplitMetaNamespaceKey(key)
	execution, err := e.executionLister.Executions(namespace).Get(name)
	if err != nil {
		glog.Errorf("Get execution %s error: %v", key, err)
		return err
	}

	execcopy := execution.DeepCopy()

	for _, tt := range execcopy.Spec.Tasks {
		if tt.Name == task.Name {
			tt.CommandSet = task.CommandSet
		}
	}
	glog.V(2).Infof("Original Execution %v ", execution)

	glog.V(2).Infof("Updated Execution %v ", execcopy)

	err = e.execUpdater.UpdateExecution(execcopy, execution)
	if err != nil {
		glog.Errorf("execUpdater.UpdateExecution failed %s error: %v", err)
		return err
	}

	// TODO need to recreate/update the graph currently
	// single vertex is used for workflow job which can have multiple k8s jobs
	// other validations may not work like num vertex & job completed event in syncJob
	// once we reconstruct the graph then we may need not required create the k8s jobs here

	// temporary create all the jobs here
	jobNamePrefix := execution.Name + Separator + task.Name + Separator
	for index, command := range task.CommandSet {
		jobName := jobNamePrefix + strconv.Itoa(index)
		// make up k8s job resource
		job := newJob(jobName, command, execution, task)

		if err := e.createJob(job); err != nil {
			glog.Errorf("createJob failed error: %v", err)
			key := util.KeyOf(job)
			return fmt.Errorf("create job %s error: %v", key, err)
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

func (e *ExecutionJobController) getJobResult(job *batch.Job) (string, error) {
	result := ""
	job, err := e.jobLister.Jobs(job.Namespace).Get(job.Name)

	if err != nil {
		glog.V(2).Infof("In getJobResult func get job failed: %v", err)
		return result, err
	}

	sel, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		glog.V(2).Infof("In getJobResult func LabelSelectorAsSelector failed: %v", err)
		return result, err
	}
	var opts metav1.ListOptions

	opts.LabelSelector = sel.String()
	podList, err := e.kubeClient.CoreV1().Pods(job.Namespace).List(opts)
	if err != nil {
		glog.V(2).Infof("In getJobResult func get pods list failed: %v", err)
		return result, err
	}

	if len(podList.Items) != 1 {
		glog.V(2).Infof("In getJobResult func  pods list has more than one pod ")
		err := fmt.Errorf("Received  podList has more than one pod")
		return result, err
	}

	res, err := e.kubeClient.CoreV1().Pods(job.Namespace).GetLogs(podList.Items[0].Name, nil).Param("limitBytes", "1024").DoRaw()
	if err != nil {
		glog.V(2).Infof("In getJobResult func get logs failed: %v", err)
		return result, err
	}
	result = string(res)
	return result, err
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
