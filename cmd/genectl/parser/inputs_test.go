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
	"github.com/ghodss/yaml"
	"testing"
)

func TestValidateInputs(t *testing.T) {
	testCases := []struct {
		Name      string
		Inputs    string
		ExpectErr bool
	}{
		{
			Name: "valid input",
			Inputs: `
  npart:
    default: 2
    description: "split job n part"
    type: number
    label: global
  obs-path:
    default: /root
    description: "obs path"
    type: string
  bwaArray:
    default: [1,2,3]
    type: array
    description: "array type"
  isOk:
    default: true
    type: bool
    description: "bool type"`,
			ExpectErr: false,
		},

		// #########################################################################################################

		{
			Name: "inputs.type error: npart.type is bool, but the value is 2",
			Inputs: `
  npart:
    default: 2
    description: "split job n part"
    type: bool
    label: global
  obs-path:
    default: /root
    description: "obs path"
    type: string
  bwaArray:
    default: [1,2,3]
    type: array
    description: "array type"
  isOk:
    default: true
    type: bool
    description: "bool type"`,
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "inputs.npart.type does not exist",
			Inputs: `
  npart:
    default: 11
    description: "split job n part"
    label: global
  obs-path:
    default: /root
    description: "obs path"
    type: string
  bwaArray:
    default: [1,2,3]
    type: array
    description: "array type"
  isOk:
    default: true
    type: bool
    description: "bool type"`,
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "inputs.npart.type [xxx] error! Type can only be in [string number bool array]",
			Inputs: `
  npart:
    default: 11
    description: "split job n part"
    type: xxx
    label: global
  obs-path:
    default: /root
    description: "obs path"
    type: string
  bwaArray:
    default: [1,2,3]
    type: array
    description: "array type"
  isOk:
    default: true
    type: bool
    description: "bool type"`,
			ExpectErr: true,
		},
	}

	for _, testCase := range testCases {
		var inputs map[string]Input
		yaml.Unmarshal([]byte(testCase.Inputs), &inputs)
		err := ValidateInputs(inputs)
		if testCase.ExpectErr == true && len(err) == 0 {
			t.Errorf("%s: Expect error, but got nil", testCase.Name)
		}
		if testCase.ExpectErr == false && len(err) != 0 {
			t.Errorf("%s: Expect no error, but got error %v", testCase.Name, err)
		}
	}
}
