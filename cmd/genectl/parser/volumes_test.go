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

func TestValidateVolumes(t *testing.T) {
	testCases := []struct {
		Name      string
		Volumes   string
		Inputs    map[string]Input
		ExpectErr bool
	}{
		{
			Name: "valid volume",
			Volumes: `
  reference:
    mount_path: /root
    mount_from:
      pvc: ${gcs-pvc}`,
			Inputs:    makeInputs(),
			ExpectErr: false,
		},
		{
			Name: "volumes[reference].mount_from: volume only support pvc and mount_from.pvc should not be empty",
			Volumes: `
  reference:
    mount_path: /root
    mount_from:
      xxx: ${gcs-pvc}`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
		{
			Name: "volumes[reference].mount_from.pvc: the variant [xxx] is not defined in the inputs",
			Volumes: `
  reference:
    mount_path: /root
    mount_from:
      pvc: ${xxx}`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
		{
			Name: "volumes[reference].mount_from.pvc: the type of bwaArray can only be in [string], but the real one is array",
			Volumes: `
volumes:
  reference:
    mount_path: /root
    mount_from:
      pvc: ${bwaArray}`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
		{
			Name: "volumes[reference].mount_path: the type of bwaArray can only be in [string], but the real one is array",
			Volumes: `
  reference:
    mount_path: ${bwaArray}
    mount_from:
      pvc: ${gcs-pvc}`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
		{
			Name: "volumes[reference].mount_path: mount path should be empty",
			Volumes: `
  reference:
    mount_path:
    mount_from:
      pvc: ${gcs-pvc}`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
		{
			Name: "volumes[reference].mount_path: mount path should be an absolute path, but the real one is fdafdasf/fdaf",
			Volumes: `
  reference:
    mount_path: fdafdasf/fdaf
    mount_from:
      pvc: ${gcs-pvc}`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
	}

	for _, testCase := range testCases {
		var volumes map[string]Volume
		yaml.Unmarshal([]byte(testCase.Volumes), &volumes)
		err := ValidateVolumes(volumes, testCase.Inputs)
		if testCase.ExpectErr == true && len(err) == 0 {
			t.Errorf("%s: Expect error, but got nil", testCase.Name)
		}
		if testCase.ExpectErr == false && len(err) != 0 {
			t.Errorf("%s: Expect no error, but got error %v", testCase.Name, err)
		}
	}
}
