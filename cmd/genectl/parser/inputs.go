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
	"fmt"

	"kubegene.io/kubegene/pkg/common"
)

func SetDefaultInputs(workflow *Workflow) {
	inputs := make(map[string]Input)
	for key, input := range workflow.Inputs {
		tmpInput := input
		if len(tmpInput.Type) == 0 {
			tmpInput.Type = StringType
		}
		inputs[key] = tmpInput
	}
	workflow.Inputs = inputs
}

func ValidateInputs(inputs map[string]Input) ErrorList {
	errorList := ErrorList{}
	for key, input := range inputs {
		if input.Type == "" {
			errorList = append(errorList, fmt.Errorf("inputs.%s.type does not exist", key))
			continue
		}
		if !IsValidType(input.Type, InputTypeList) {
			err := fmt.Errorf("inputs.%s.type [%s] is invalid type. Valid type: %v", key, input.Type, InputTypeList)
			errorList = append(errorList, err)
		}
		if input.Default != nil {
			if !IsValidInputValue(input.Default, input.Type) {
				err := fmt.Errorf("inputs.type error: %s.type is %s, but the value is %v", key, input.Type, input.Default)
				errorList = append(errorList, err)
			}
		}
	}
	return errorList
}

func MergeInputs(inputsDefult map[string]Input, inputs map[string]interface{}) (map[string]Input, error) {
	mergedInputs := make(map[string]Input, len(inputsDefult))
	for key, input := range inputsDefult {
		// If input value is empty, then we will use default.
		if input.Value == nil {
			input.Value = input.Default
		}

		value, ok := inputs[key]
		if ok {
			if !IsValidInputValue(value, input.Type) {
				return nil, fmt.Errorf("type error: inputs.%s.type is %s, but the given input value is %v", key, input.Type, value)
			}
			input.Value = value
		}

		if input.Value == nil {
			return nil, fmt.Errorf("inputs.[%s].value is empty", key)
		}
		mergedInputs[key] = input
	}

	for key, value := range inputs {
		_, ok := inputsDefult[key]
		if !ok {
			inputType := GetInputType(value)
			mergedInputs[key] = Input{
				Value: value,
				Type:  inputType,
			}
		}
	}

	return mergedInputs, nil
}

func Inputs2ReplaceData(inputs map[string]Input) map[string]string {
	data := make(map[string]string)
	for key, val := range inputs {
		data[key] = common.ToString(val.Value)
	}
	return data
}

func GetStringValue(key string, inputs map[string]Input) string {
	if value, ok := inputs[key]; ok {
		return common.ToString(value.Value)
	}
	return ""
}
