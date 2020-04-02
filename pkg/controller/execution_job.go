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
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	genelisters "kubegene.io/kubegene/pkg/client/listers/gene/v1alpha1"
	"kubegene.io/kubegene/pkg/common"
	"kubegene.io/kubegene/pkg/graph"
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

	klog.Infof("Starting execution job controller")
	defer klog.Infof("Shutting down execution job controller")

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
		klog.V(2).Infof("Parallel running jobs exceeded the execution limit, retry after 10s")
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
		klog.V(2).Infof("graph of execution %s does not exist", event.Key)
		return nil
	}

	switch event.Type {
	case NewAdded:
		klog.V(2).Infof("execution %v start running.", event.Key)

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
		klog.V(2).Infof("job %v has run successfully.", event.Name)

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
				klog.V(2).Infof("all dependent of job %v has run successfully, start running.", child.Data.Job.Name)

				if child.Data.DynamicJob != nil {
					// flag setting true to handle if Condition is nil
					flag := true
					var err error

					if child.Data.DynamicJob.GenericCondition != nil {
						klog.V(2).Infof(" conditional based job GenericCondition:%v", child.Data.DynamicJob.GenericCondition)
						flag, err = e.evalGenericConditionResult(vertex.Data.Job, child, graph, event.Key)
						if err != nil {
							return fmt.Errorf("evalGenericConditionResult failed : %v", err)
						}
						if !flag {
							// if the condition or check_result validation is false
							// then we don't create the k8s job and make the job is Finished true
							// so that other jobs will continue or execution will complete
							klog.V(2).Infof(" The final condition is false")
							child.Data.Finished = true
							continue
						}
					}

					if child.Data.DynamicJob.Condition != nil {
						klog.V(2).Infof(" conditional based job condition:%v", child.Data.DynamicJob.Condition)
						flag, err = e.evalConditionResult(vertex.Data.Job, child, graph, event.Key)
						if err != nil {
							return fmt.Errorf("evalConditionResult failed : %v", err)
						}
						if !flag {
							// if the condition or check_result validation is false
							// then we don't create the k8s job and make the job is Finished true
							// so that other jobs will continue or execution will complete
							klog.V(2).Infof(" The final condition is false")
							child.Data.Finished = true
							continue
						}
					}

					if len(child.Data.DynamicJob.CommandSet) > 0 && child.Data.DynamicJob.CommandsIter == nil {

						// construct the dynamic jobs based on conditional branch result
						err := e.createDynamicJobsBasedOnConditionalChk(child, graph, event.Key)
						if err != nil {
							return fmt.Errorf("createDynamicJobsBasedOnConditionalChk failed : %v", err)
						}
						continue
					}

					if child.Data.DynamicJob.CommandsIter != nil {

						// get the result of the dependent job
						result, err := e.getJobResult(vertex.Data.Job)
						if err != nil {
							return fmt.Errorf("getJobResult failed : %v", err)
						}
						// construct the dynamic job based on get_result
						err = e.createDynamicJob(child, result, graph, event.Key)
						if err != nil {
							return fmt.Errorf("createDynamicJob failed : %v", err)
						}
						continue
					}

				}
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

func (e *ExecutionJobController) evalGenericConditionResult(dependJob *batch.Job, vertex *graph.Vertex, graph *graph.Graph, key string) (bool, error) {

	klog.V(6).Infof("In evalGenericConditionResult GenericCondition:%v", vertex.Data.DynamicJob.GenericCondition)

	genericCond := vertex.Data.DynamicJob.GenericCondition

	// get the result of the dependent job
	result, err := e.getJobResult(dependJob)
	if err != nil {
		return false, fmt.Errorf("getJobResult failed in evalGenericConditionResult: %v", err)
	}

	var keyvalues map[string]string

	keyvalues = make(map[string]string, 0)
	strs := strings.Split(result, ",")

	klog.V(6).Infof("In evalGenericConditionResult strs %v", strs)

	for i := range strs {

		kv := strings.Split(strs[i], ":")
		klog.V(2).Infof("In evalGenericConditionResult kv %v", kv)
		keyvalues[kv[0]] = kv[1]
	}
	klog.V(2).Infof("In evalGenericConditionResult keyvalues %v", keyvalues)

	// check for the matching key values from the result to generic condition
	for i := range genericCond.MatchRules {
		if util.RuleSatisfied(genericCond.MatchRules[i], keyvalues) {
			klog.V(2).Infof("successfully matched for keyvalues: %v for rules:%v", keyvalues, genericCond.MatchRules)
			return true, nil
		}
	}
	klog.V(2).Infof("In evalGenericConditionResult Rules are not matched")
	return false, fmt.Errorf("Rules are not matched")
}

func (e *ExecutionJobController) evalConditionResult(dependJob *batch.Job, vertex *graph.Vertex, graph *graph.Graph, key string) (bool, error) {

	klog.V(6).Infof("In evalConditionResult condition:%v", vertex.Data.DynamicJob.Condition.Condition)

	v := vertex.Data.DynamicJob.Condition.Condition.([]interface{})

	switch v[0].(type) {

	case bool:
		return v[0].(bool), nil
	case string:
		if item, ok := v[0].(string); ok && item == "check_result" {

			parentJobName := v[1].(string)
			exp := v[2].(string)
			klog.V(6).Infof("In evalConditionResult jobName: %s exp:%s", parentJobName, exp)
			// get the result of the dependent job
			result, err := e.getJobResult(dependJob)
			if err != nil {
				return false, fmt.Errorf("getJobResult failed in evalConditionResult: %v", err)
			}

			if exp == result {
				return true, nil
			} else {
				return false, fmt.Errorf("In evalConditionResult job result is %v but expected value is  %v", result, exp)
			}
		} else {
			return false, fmt.Errorf("In evalConditionResult Invalid condition %v", vertex.Data.DynamicJob.Condition.Condition)
		}
	}

	return false, fmt.Errorf("In evalConditionResult Invalid condition %v", vertex.Data.DynamicJob.Condition.Condition)
}

func evalJobResult(jobResult string, vars []interface{}) ([]common.Var, error) {
	result := make([]common.Var, 0, len(vars))
	klog.V(6).Infof("In evalJobResult vars:%v", vars)

	for _, vs := range vars {
		v := vs.([]interface{})

		if item, ok := v[0].(string); ok && item == "get_result" {

			parentJobName := v[1].(string)
			sep := v[2].(string)

			if sep == "" {
				temp := []interface{}{jobResult}
				result = append(result, temp)
			} else {
				var strslice []interface{}
				for _, str := range strings.Split(jobResult, sep) {
					if str != "" {
						strslice = append(strslice, str)
					}
				}
				klog.Infof("In evalJobResult strslice:%v", strslice)

				result = append(result, strslice)
			}
			klog.V(6).Infof("In evalJobResult jobName: %s sep:%s", parentJobName, sep)
		} else {
			//except get_result other parameters need to be appended
			result = append(result, v)
		}
	}
	klog.V(6).Infof("In evalJobResult result: %v", result)
	return result, nil
}

func ConvertVars(vars []interface{}) []common.Var {
	result := make([]common.Var, 0, len(vars))

	for _, v := range vars {
		res := make([]interface{}, 0)
		res = append(res, v)
		result = append(result, res)
	}
	return result
}

func (e *ExecutionJobController) createDynamicJobsBasedOnConditionalChk(vertex *graph.Vertex, graph *graph.Graph, key string) error {
	klog.V(2).Infof(" The condition based job which has normal commands")
	task := vertex.Data.DynamicJob

	// update this task in the execution struct then call the Update()
	namespace, name, _ := cache.SplitMetaNamespaceKey(key)
	execution, err := e.executionLister.Executions(namespace).Get(name)
	if err != nil {
		klog.Errorf("Get execution %s error: %v", key, err)
		return err
	}
	//set the dynamic job Count of this vertex
	vertex.SetDynamicJobCnt(len(task.CommandSet))

	// create all the jobs here
	jobNamePrefix := execution.Name + Separator + task.Name + Separator
	for index, command := range task.CommandSet {
		jobName := jobNamePrefix + strconv.Itoa(index)
		// make up k8s job resource
		job := newJob(jobName, command, execution, task)

		if err := e.createJob(job); err != nil {
			klog.Errorf("createJob failed error: %v", err)
			key := util.KeyOf(job)
			return fmt.Errorf("create job %s error: %v", key, err)
		}
	}
	//set the dynamic job Count of this vertex
	// -1 because already one vertex  related to dynamic job is considered as one job
	graph.AddDynamicJobCnt(len(task.CommandSet) - 1)

	return nil
}

func (e *ExecutionJobController) createDynamicJob(vertex *graph.Vertex, jobResult string, graph *graph.Graph, key string) error {

	task := vertex.Data.DynamicJob

	klog.V(2).Infof("vertex.Data.DynamicJob.CommandsIter.VarsIter : %#v", vertex.Data.DynamicJob.CommandsIter.VarsIter)
	varsIter, err := evalJobResult(jobResult, vertex.Data.DynamicJob.CommandsIter.VarsIter)
	if err != nil {
		klog.V(2).Infof("Error in evalJobResult execution job name %q , . Error: %v", task.Name, err)
		return err
	}
	klog.V(2).Infof("evalJobResult output varsIter: %v", varsIter)

	// convert varsIter to var
	iterVars := common.VarIter2Vars(varsIter)

	klog.V(2).Infof(" In createDynamicJob iterVars: %v", iterVars)

	command := vertex.Data.DynamicJob.CommandsIter.Command

	klog.V(2).Infof(" In createDynamicJob command: %v", command)

	// generate all commands.
	iterCommands := common.Iter2Array(command, iterVars)

	task.CommandSet = iterCommands

	klog.V(2).Infof("final commandset task.CommandSet %v ", task.CommandSet)

	// update this task in the execution struct then call the Update()
	namespace, name, _ := cache.SplitMetaNamespaceKey(key)
	execution, err := e.executionLister.Executions(namespace).Get(name)
	if err != nil {
		klog.Errorf("Get execution %s error: %v", key, err)
		return err
	}
	//set the dynamic job Count of this vertex
	vertex.SetDynamicJobCnt(len(task.CommandSet))

	// create all the jobs here
	jobNamePrefix := execution.Name + Separator + task.Name + Separator
	for index, command := range task.CommandSet {
		jobName := jobNamePrefix + strconv.Itoa(index)
		// make up k8s job resource
		job := newJob(jobName, command, execution, task)

		if err := e.createJob(job); err != nil {
			klog.Errorf("createJob failed error: %v", err)
			key := util.KeyOf(job)
			return fmt.Errorf("create job %s error: %v", key, err)
		}
	}
	//set the dynamic job Count of this vertex
	// -1 because already one vertex  related to dynamic job is considered as one job
	graph.AddDynamicJobCnt(len(task.CommandSet) - 1)

	return nil
}

func (e *ExecutionJobController) createJob(job *batch.Job) error {
	_, err := e.jobLister.Jobs(job.Namespace).Get(job.Name)
	// job has been already created
	if !errors.IsNotFound(err) {
		return nil
	}

	_, err = e.kubeClient.BatchV1().Jobs(job.Namespace).Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

func (e *ExecutionJobController) getJobResult(job *batch.Job) (string, error) {
	result := ""
	job, err := e.jobLister.Jobs(job.Namespace).Get(job.Name)

	if err != nil {
		klog.V(2).Infof("In getJobResult func get job failed: %v", err)
		return result, err
	}

	sel, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		klog.V(2).Infof("In getJobResult func LabelSelectorAsSelector failed: %v", err)
		return result, err
	}
	var opts metav1.ListOptions

	opts.LabelSelector = sel.String()
	podList, err := e.kubeClient.CoreV1().Pods(job.Namespace).List(context.TODO(), opts)
	if err != nil {
		klog.V(2).Infof("In getJobResult func get pods list failed: %v", err)
		return result, err
	}

	if len(podList.Items) != 1 {
		klog.V(2).Infof("In getJobResult func  pods list has more than one pod ")
		err := fmt.Errorf("Received  podList has more than one pod")
		return result, err
	}

	//size limit 1k bytes extra 100 bytes added
	var sizeLimit int64
	sizeLimit = 1024 + 100

	opt := v1.PodLogOptions{LimitBytes: &sizeLimit, SinceTime: &metav1.Time{}}

	res, err := e.kubeClient.CoreV1().Pods(job.Namespace).GetLogs(podList.Items[0].Name,
		&opt).Stream(context.TODO())
	if err != nil {
		klog.V(2).Infof("In getJobResult with opt func get logs failed: %v", err)
		return result, err
	}
	bytes, err := ioutil.ReadAll(res)
	if err != nil {
		klog.V(2).Infof("In getJobResult func ioutil.ReadAll failed: %v", err)
		return result, err
	}
	result = string(bytes)
	result = strings.TrimSuffix(result, "\n")
	klog.V(2).Infof("the succful getJobResult is: %s", result)
	return result, err
}

func (e *ExecutionJobController) handleErr(err error, event Event) {
	if err == nil {
		e.queue.Forget(event)
		return
	}

	if e.queue.NumRequeues(event) < maxRetries {
		klog.V(2).Infof("Error syncing execution jobs %q , retrying. Error: %v", event.Key, err)
		e.queue.AddRateLimited(event)
		return
	}

	klog.Warningf("Dropping event %v out of the queue: %v", event, err)
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
		klog.V(2).Infof("Execution %v has been deleted", key)
		return false
	}
	if err != nil {
		klog.Errorf("Get execution %s error: %v", key, err)
		return false
	}
	if execution.Spec.Parallelism != nil {
		jobs, err := e.getActiveJobsForExecution(job.Namespace, labels.Set(job.Labels).AsSelector())
		if err != nil {
			klog.Errorf("Get active jobs for execution %s error: %v", key, err)
			return false
		}

		if len(jobs) >= int(*execution.Spec.Parallelism) {
			return false
		}
	}

	return true
}
