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

package parser

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubegene.io/kubegene/cmd/genectl/util"
)

const DefaultParallelism = 5

func GetExecutionNamespace(inputs map[string]Input) string {
	namespace := GetStringValue("namespace", inputs)
	if len(namespace) == 0 {
		namespace = metav1.NamespaceDefault
	}
	return namespace
}

func GetExecutionName(inputs map[string]Input) string {
	name := GetStringValue("executionName", inputs)
	if len(name) == 0 {
		name = util.GenerateExecName("workflows")
	}
	return name
}

func GetAffinityFromInputs(inputs map[string]Input) (*apiv1.Affinity, error) {
	// interface -> json -> apiv1.Affinity
	if input, ok := inputs["affinity"]; ok {
		bytes, err := json.Marshal(input.Value)
		if err != nil {
			return nil, fmt.Errorf("marshal affinity error: %v", err)
		}

		var affinity apiv1.Affinity
		err = json.Unmarshal(bytes, &affinity)
		if err != nil {
			return nil, fmt.Errorf("unmarshal affinity error: %v", err)
		}

		return &affinity, nil
	}

	return nil, nil
}

func GetTolerationsFromInputs(inputs map[string]Input) ([]apiv1.Toleration, error) {
	// interface -> json -> []apiv1.Toleration
	if input, ok := inputs["tolerations"]; ok {
		bytes, err := json.Marshal(input.Value)
		if err != nil {
			return nil, fmt.Errorf("marshal tolerations error: %v", err)
		}

		var tolerations []apiv1.Toleration
		err = json.Unmarshal(bytes, &tolerations)
		if err != nil {
			return nil, fmt.Errorf("unmarshal tolerations error: %v", err)
		}

		return tolerations, nil
	}

	return nil, nil
}

func GetNodeSelectorFromInputs(inputs map[string]Input) (map[string]string, error) {
	// interface -> json -> map[string]string
	if input, ok := inputs["nodeSelector"]; ok {
		bytes, err := json.Marshal(input.Value)
		if err != nil {
			return nil, fmt.Errorf("marshal nodeSelector error: %v", err)
		}

		var nodeSelector map[string]string
		err = json.Unmarshal(bytes, &nodeSelector)
		if err != nil {
			return nil, fmt.Errorf("unmarshal nodeSelector error: %v", err)
		}

		return nodeSelector, nil
	}
	return nil, nil
}

func GetBackoffLimit(inputs map[string]Input) (*int32, error) {
	if input, ok := inputs["backoffLimit"]; ok {
		value := input.Value
		backoffLimit, ok := value.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid backoffLimit")
		}

		backoffLimitInt32 := int32(backoffLimit)
		return &backoffLimitInt32, nil
	}

	return nil, nil
}

func GetParallelism(inputs map[string]Input) *int64 {
	if input, ok := inputs["parallelism"]; ok {
		value := input.Value
		parallelism, ok := value.(float64)
		if !ok {
			glog.Infof("invalid parallelism, using the default parallelism")
			return NewInt64(DefaultParallelism)
		}

		parallelismInt64 := int64(parallelism)
		return &parallelismInt64
	}

	return NewInt64(DefaultParallelism)
}

func GetActiveDeadlineSeconds(inputs map[string]Input) (*int64, error) {
	if input, ok := inputs["activeDeadlineSeconds"]; ok {
		value := input.Value
		activeDeadlineSeconds, ok := value.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid activeDeadlineSeconds")
		}

		activeDeadlineSecondsInt64 := int64(activeDeadlineSeconds)
		return &activeDeadlineSecondsInt64, nil
	}

	return nil, nil
}

func NewInt64(n int64) *int64 {
	return &n
}
