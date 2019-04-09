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

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/validation"

	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
)

// ValidateExecution accepts a execution and performs validation against it. If lint is specified as
// true, will skip some validations which is permissible during linting but not submission
func ValidateExecution(execution *genev1alpha1.Execution) error {
	if msgs := validation.IsDNS1123Label(execution.Name); len(msgs) > 0 {
		return fmt.Errorf("name is not valid %v", msgs)
	}

	if execution.Spec.Parallelism != nil && *execution.Spec.Parallelism < 0 {
		return fmt.Errorf("parallelism must be greater than or equal to 0")
	}
	if len(execution.Spec.Tasks) == 0 {
		return fmt.Errorf("tasks of execution must not be empty")
	}
	if err := validateTasks(execution.Spec.Tasks); err != nil {
		return err
	}
	if ok := validateNoCycle(execution); !ok {
		return fmt.Errorf("dependents of execution exist cycle")
	}
	return nil
}

func validateNoCycle(execution *genev1alpha1.Execution) bool {
	graph := newGraph(execution)
	return graph.IsDAG()
}

func validateTasks(tasks []genev1alpha1.Task) error {
	taskNames := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		if _, exist := taskNames[task.Name]; exist {
			return fmt.Errorf("task name %s duplicated", task.Name)
		}
		if err := validateTask(task, tasks); err != nil {
			return err
		}
	}
	return nil
}

func validateTask(task genev1alpha1.Task, tasks []genev1alpha1.Task) error {
	if msgs := validation.IsDNS1123Label(task.Name); len(msgs) > 0 {
		return fmt.Errorf("task name %s is not valid %v", task.Name, msgs)
	}
	if len(task.Image) == 0 {
		return fmt.Errorf("task image must not be empty")
	}
	if len(task.CommandSet) == 0 && task.CommandsIter == nil {
		return fmt.Errorf("task commandSet must not be empty")
	}
	if task.Parallelism != nil && *task.Parallelism < 0 {
		return fmt.Errorf("task parallelism must be greater than or equal to 0")
	}
	if task.BackoffLimit != nil && *task.BackoffLimit < 0 {
		return fmt.Errorf("task backoffLimit must be greater than or equal to 0")
	}
	if task.ActiveDeadlineSeconds != nil && *task.ActiveDeadlineSeconds < 0 {
		return fmt.Errorf("task activeDeadlineSeconds must be greater than or equal to 0")
	}
	if task.Type != genev1alpha1.JobTaskType && task.Type != genev1alpha1.SparkTaskType {
		return fmt.Errorf("wrong task type: %s", task.Type)
	}
	if len(task.Dependents) != 0 {
		if err := validateDependents(task.Name, task.Dependents, tasks); err != nil {
			return err
		}
	}
	if task.GenericCondition != nil {
		if err := validateGenericCondition(task.Name, task.GenericCondition, tasks); err != nil {
			return err
		}
	}
	if task.Condition != nil {
		if err := validateCondition(task.Name, task.Condition, tasks); err != nil {
			return err
		}
	}
	if task.CommandsIter != nil {
		if err := validateCommandsIter(task.Name, task.CommandsIter, tasks); err != nil {
			return err
		}
	}

	return nil
}

func validateDependents(taskName string, dependents []genev1alpha1.Dependent, tasks []genev1alpha1.Task) error {
	for _, dependent := range dependents {
		if dependent.Type != genev1alpha1.DependTypeWhole && dependent.Type != genev1alpha1.DependTypeIterate {
			return fmt.Errorf("wrong dependent type of task %s: %s", taskName, dependent.Type)
		}
		if len(dependent.Target) == 0 {
			return fmt.Errorf("%s: dependent target must not be empty", taskName)
		}
		if !taskExist(tasks, dependent.Target) {
			return fmt.Errorf("%s: dependent target %s not exist", taskName, dependent.Target)
		}
	}

	return nil
}

func taskExist(tasks []genev1alpha1.Task, name string) bool {
	for _, task := range tasks {
		if task.Name == name {
			return true
		}
	}
	return false
}

func getTaskByName(tasks []genev1alpha1.Task, name string) (bool, genev1alpha1.Task) {
	var tt genev1alpha1.Task
	for _, task := range tasks {
		if task.Name == name {
			return true, task
		}
	}
	return false, tt
}

func validateGenericDependency(taskName string, dependJobName string, tasks []genev1alpha1.Task) error {
	var dependTask genev1alpha1.Task
	var flag bool
	flag, dependTask = getTaskByName(tasks, dependJobName)

	if !flag {
		return fmt.Errorf(" the dependecy job is missing, but the real one is %s", dependJobName)
	} else {
		// depend job should have single command only because it should be single k8s- job related to that job
		if (len(dependTask.CommandSet) > 1) ||
			((dependTask.CommandsIter != nil) && len(dependTask.CommandsIter.VarsIter) > 1) {

			return fmt.Errorf("the dependecy job has more than one command  dependTask :%v", dependTask)
		}
	}

	var currentTask genev1alpha1.Task
	flag, currentTask = getTaskByName(tasks, taskName)

	if !flag {
		return fmt.Errorf("the current task is missing, but the real one is %s", taskName)
	}

	if len(currentTask.Dependents) != 1 {
		return fmt.Errorf("the current task has more dependecies %v", currentTask.Dependents)
	}

	for i := 0; i < len(currentTask.Dependents); i++ {
		if (currentTask.Dependents[i].Target == dependJobName) &&
			(currentTask.Dependents[i].Type) == "whole" {
			return nil
		}
	}

	return fmt.Errorf("the task dependency type is wrong ")
}

func validateGenericCondition(taskName string, gCondition *genev1alpha1.GenericCondition, tasks []genev1alpha1.Task) error {

	err := validateGenericDependency(taskName, gCondition.DependJobName, tasks)
	if err != nil {
		return err
	}

	for i := range gCondition.MatchRules {
		prefix := fmt.Sprintf("executionspec.%s.genericcondition.matchrules[%d]", taskName, i)
		switch gCondition.MatchRules[i].Operator {
		case genev1alpha1.MatchOperatorOpIn, genev1alpha1.MatchOperatorOpNotIn,
			genev1alpha1.MatchOperatorOpEqual, genev1alpha1.MatchOperatorOpDoubleEqual,
			genev1alpha1.MatchOperatorOpNotEqual:
			if len(gCondition.MatchRules[i].Values) == 0 {
				return fmt.Errorf("%s.values must be specified when `operator` is 'In' or 'NotIn'", prefix)
			}
		case genev1alpha1.MatchOperatorOpExists, genev1alpha1.MatchOperatorOpDoesNotExist:
			if len(gCondition.MatchRules[i].Values) > 0 {
				return fmt.Errorf("%s.values must not be specified when `operator` is 'Exists' or 'DoesNotExist' ", prefix)
			}
		case genev1alpha1.MatchOperatorOpGt, genev1alpha1.MatchOperatorOpLt:
			if len(gCondition.MatchRules[i].Values) != 1 {
				return fmt.Errorf("%s.values must be specified single value when `operator` is 'Lt' or 'Gt' ", prefix)
			}
		default:
			return fmt.Errorf("%s.values not a valid Match operator ", prefix)

		}
		if len(gCondition.MatchRules[i].Key) == 0 {
			return fmt.Errorf("%s.key should not be empty or shoud not be variant ", prefix)
		}
	}
	return nil
}

func validateCommandsIter(taskName string, commandsIter *genev1alpha1.CommandsIter, tasks []genev1alpha1.Task) error {
	glog.V(2).Infof("In validateCommandsIter commandsIter %v", commandsIter)
	var v []interface{}
	if commandsIter.Command == "" {
		return fmt.Errorf("%s task command must not be empty in commands_iter if commands_iter is not nil", taskName)
	}
	for _, vs := range commandsIter.VarsIter {

		switch vs.(type) {
		case []interface{}:
			v = vs.([]interface{})

		default:
			return fmt.Errorf("The commandsIter  format is wrong in task :%s", taskName)
		}

		if func_name, ok := v[0].(string); ok && func_name == "get_result" {
			if len(v) != 3 {
				return fmt.Errorf("In commands_iter  get_result format is wrong in task :%s", taskName)
			}
			var dependJobName string
			if dependJobName, ok = v[1].(string); !ok {
				return fmt.Errorf("In commands_iter  get_result doesn't have the depend job parameter in task :%s", taskName)
			}
			if err := validateGenericDependency(taskName, dependJobName, tasks); err != nil {
				return err
			}
			if _, ok := v[2].(string); !ok {
				return fmt.Errorf("In commands_iter  get_result doesn't have the exp parameter in task :%s", taskName)
			}
		}
	}
	return nil

}

func validateCondition(taskName string, condition *genev1alpha1.Condition, tasks []genev1alpha1.Task) error {

	var v []interface{}
	glog.V(2).Infof("In validateCondition condition %v", condition)
	switch condition.Condition.(type) {
	case []interface{}:
		v = condition.Condition.([]interface{})

	default:
		return fmt.Errorf("The condition  format is wrong in task :%s", taskName)
	}

	if _, ok := v[0].(bool); ok {
		return nil
	}

	if func_name, ok := v[0].(string); ok && func_name == "check_result" {

		if len(v) != 3 {
			return fmt.Errorf("In condition  check_result format is wrong in task :%s", taskName)
		}
		var dependJobName string
		if dependJobName, ok = v[1].(string); !ok {
			return fmt.Errorf("In condition  check_result doesn't have the depend job parameter in task :%s", taskName)
		}
		if err := validateGenericDependency(taskName, dependJobName, tasks); err != nil {
			return err
		}
		if _, ok := v[2].(string); !ok {
			return fmt.Errorf("In condition  check_result doesn't have the exp parameter in task :%s", taskName)
		}
	} else {
		return fmt.Errorf("%s task has other than check_result in condition ", taskName)
	}
	return nil
}
