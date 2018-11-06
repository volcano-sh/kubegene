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

	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
)

// ValidateExecution accepts a execution and performs validation against it. If lint is specified as
// true, will skip some validations which is permissible during linting but not submission
func ValidateExecution(execution *genev1alpha1.Execution) error {
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
	for _, task := range tasks {
		if err := validateTask(task, tasks); err != nil {
			return err
		}
	}
	return nil
}

func validateTask(task genev1alpha1.Task, tasks []genev1alpha1.Task) error {
	if len(task.Name) == 0 {
		return fmt.Errorf("task name must not be empty")
	}
	if len(task.Image) == 0 {
		return fmt.Errorf("task image must not be empty")
	}
	if len(task.CommandSet) == 0 {
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

	return nil
}

func validateDependents(taskName string, dependents []genev1alpha1.Dependent, tasks []genev1alpha1.Task) error {
	for _, dependent := range dependents {
		if dependent.Type != genev1alpha1.DependTypeWhole && dependent.Type != genev1alpha1.DependTypeIterate{
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
