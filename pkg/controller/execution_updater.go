/*
Copyright 2019 The Kubegene Authors.

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
	"encoding/json"
	"fmt"

	"github.com/evanphx/json-patch"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	geneclientset "kubegene.io/kubegene/pkg/client/clientset/versioned/typed/gene/v1alpha1"
	"kubegene.io/kubegene/pkg/util"
)

// ExecutionUpdater is an interface used to update the ExecutionSpec associated with a Execution.
type ExecutionUpdater interface {
	UpdateExecution(modified *genev1alpha1.Execution, original *genev1alpha1.Execution) error
}

// NewExecutionUpdater returns a ExecutionUpdater that updates the Spec of a Execution,
// using the supplied client and setLister.
func NewExecutionUpdater(client geneclientset.ExecutionsGetter) ExecutionUpdater {
	return &executionUpdater{client}
}

type executionUpdater struct {
	execClient geneclientset.ExecutionsGetter
}

func (esu *executionUpdater) UpdateExecution(modified *genev1alpha1.Execution, original *genev1alpha1.Execution) error {
	patchBytes, err := preparePatchBytesForExecution(modified, original)
	if err != nil {
		return err
	}

	for i := 0; i < specUpdateRetries; i++ {
		var current *genev1alpha1.Execution
		current, err = esu.execClient.Executions(modified.Namespace).Get(modified.Name, metav1.GetOptions{})
		if err != nil {
			glog.Error("getting the execution is failed. Error: %v", err)
			break
		}

		var curBytes []byte
		curBytes, err = json.Marshal(current)
		if err != nil {
			glog.Error("after getting the execution json.Marshal failed. Error: %v", err)
			break
		}

		var bytes []byte
		bytes, err = jsonpatch.MergePatch(curBytes, patchBytes)
		if err != nil {
			glog.Error("after getting the execution jsonpatch.MergePatch failed. Error: %v", err)
			break
		}

		var updated genev1alpha1.Execution
		err = json.Unmarshal(bytes, &updated)
		if err != nil {
			glog.Error("after getting the execution json.Unmarshal failed. Error: %v", err)
			break
		}

		_, err = esu.execClient.Executions(modified.Namespace).Update(&updated)
		if err == nil {
			break
		}
	}

	return err
}

func preparePatchBytesForExecution(modifiedExec *genev1alpha1.Execution, originExec *genev1alpha1.Execution) ([]byte, error) {
	origin, err := json.Marshal(originExec)
	if err != nil {
		glog.Error("In preparePatchBytesForExecution func  json.Marshal failed for %s Error: %v", util.KeyOf(originExec), err)
		return nil, fmt.Errorf("unable to marshal execution %s", util.KeyOf(originExec))
	}

	modified, err := json.Marshal(modifiedExec)
	if err != nil {
		glog.Error("In preparePatchBytesForExecution func  json.Marshal failed for %s Error: %v", util.KeyOf(modifiedExec), err)
		return nil, fmt.Errorf("unable to marshal execution %s", util.KeyOf(modifiedExec))
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origin, modified)
	if err != nil {
		glog.Error("In preparePatchBytesForExecution func  jsonpatch.CreateMergePatch failed for %s Error: %v", util.KeyOf(modifiedExec), err)
		return nil, fmt.Errorf("failed to CreateMergePatch for execution %q: %v", util.KeyOf(modifiedExec), err)
	}

	return patchBytes, nil
}
