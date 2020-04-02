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
	"encoding/json"
	"fmt"
	"time"

	"github.com/evanphx/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"

	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	geneclientset "kubegene.io/kubegene/pkg/client/clientset/versioned/typed/gene/v1alpha1"
	"kubegene.io/kubegene/pkg/util"
)

// DefaultRetry is a default retry backoff settings when retrying API calls
var DefaultRetry = wait.Backoff{
	Steps:    UpdateRetries,         // Exit with error after this many steps
	Duration: 10 * time.Millisecond, // the base duration
	Factor:   1.0,                   // Duration is multiplied by factor each iteration
	Jitter:   0.1,                   // The amount of jitter applied each iteration
}

// ExecutionStatusUpdater is an interface used to update the ExecutionStatus associated with a Execution.
type ExecutionUpdater interface {
	// UpdateStatefulSetStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateExecutionStatus(modified *genev1alpha1.Execution, original *genev1alpha1.Execution) error
	UpdateExecution(modified *genev1alpha1.Execution, original *genev1alpha1.Execution) error
}

// NewExecutionStatusUpdater returns a ExecutionStatusUpdater that updates the Status of a Execution,
// using the supplied client and setLister.
func NewExecutionStatusUpdater(client geneclientset.ExecutionsGetter) ExecutionUpdater {
	return &executionUpdater{client}
}

type executionUpdater struct {
	execClient geneclientset.ExecutionsGetter
}

func (esu *executionUpdater) UpdateExecutionStatus(modified *genev1alpha1.Execution, original *genev1alpha1.Execution) error {
	patchBytes, err := preparePatchBytesForExecutionStatus(modified, original)
	if err != nil {
		return err
	}

	var current *genev1alpha1.Execution
	current, err = esu.execClient.Executions(modified.Namespace).Get(context.TODO(), modified.Name, metav1.GetOptions{})
	if err != nil {
		klog.V(2).Infof("getting the execution is failed. Error: %v", err)
		return err
	}

	var curBytes []byte
	curBytes, err = json.Marshal(current)
	if err != nil {
		klog.V(2).Infof("after getting the execution json.Marshal failed. Error: %v", err)
		return err
	}

	var bytes []byte
	bytes, err = jsonpatch.MergePatch(curBytes, patchBytes)
	if err != nil {
		klog.V(2).Infof("after getting the execution jsonpatch.MergePatch failed. Error: %v", err)
		return err
	}

	var updated genev1alpha1.Execution
	err = json.Unmarshal(bytes, &updated)
	if err != nil {
		klog.V(2).Infof("after getting the execution json.Unmarshal failed. Error: %v", err)
		return err
	}

	err = wait.ExponentialBackoff(DefaultRetry, func() (bool, error) {
		_, err = esu.execClient.Executions(modified.Namespace).UpdateStatus(context.TODO(), &updated, metav1.UpdateOptions{})
		if err != nil {
			klog.V(2).Infof("Failed to update execution status '%s': %v", current.Name, err)
			return false, nil
		}
		return true, nil
	})

	return err
}
func (esu *executionUpdater) UpdateExecution(modified *genev1alpha1.Execution, original *genev1alpha1.Execution) error {
	patchBytes, err := preparePatchBytesForExecutionStatus(modified, original)
	if err != nil {
		return err
	}

	var current *genev1alpha1.Execution
	current, err = esu.execClient.Executions(modified.Namespace).Get(context.TODO(), modified.Name, metav1.GetOptions{})
	if err != nil {
		klog.V(2).Infof("getting the execution is failed %v", err)
		return err
	}

	var curBytes []byte
	curBytes, err = json.Marshal(current)
	if err != nil {
		klog.V(2).Infof("after getting the execution json.Marshal failed %v", err)
		return err
	}

	var bytes []byte
	bytes, err = jsonpatch.MergePatch(curBytes, patchBytes)
	if err != nil {
		klog.V(2).Infof("after getting the execution jsonpatch.MergePatch failed. Error: %v", err)
		return err
	}

	var updated genev1alpha1.Execution
	err = json.Unmarshal(bytes, &updated)
	if err != nil {
		klog.V(2).Infof("after getting the execution json.Unmarshal failed. Error: %v", err)
		return err
	}

	err = wait.ExponentialBackoff(DefaultRetry, func() (bool, error) {
		_, err = esu.execClient.Executions(modified.Namespace).Update(context.TODO(), &updated, metav1.UpdateOptions{})
		if err != nil {
			klog.V(2).Infof("Failed to update execution '%s': %v", current.Name, err)
			return false, nil
		}
		return true, nil
	})

	return err
}
func preparePatchBytesForExecutionStatus(modifiedExec *genev1alpha1.Execution, originExec *genev1alpha1.Execution) ([]byte, error) {
	origin, err := json.Marshal(originExec)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal execution %s", util.KeyOf(originExec))
	}

	modified, err := json.Marshal(modifiedExec)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal execution %s", util.KeyOf(modifiedExec))
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origin, modified)
	if err != nil {
		return nil, fmt.Errorf("failed to CreateMergePatch for execution %q: %v", util.KeyOf(modifiedExec), err)
	}

	return patchBytes, nil
}
