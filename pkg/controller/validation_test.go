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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
)

type ModifyExecution func(exec *genev1alpha1.Execution)

func validateExecution() *genev1alpha1.Execution {
	validateExecution := genev1alpha1.Execution{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Execution",
			APIVersion: genev1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "simple-example",
			Namespace: "exec-system",
		},
		Spec: genev1alpha1.ExecutionSpec{
			Tasks: []genev1alpha1.Task{
				{
					Name:       "a",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo A >> /tmp/hostvolume/"},
					Image:      "hello-word",
					Volumes: map[string]genev1alpha1.Volume{
						"volumea": {
							MountPath: "/tmp/hostvolume",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: "test-host-path",
							},
						},
					},
				},
				{
					Name:       "b",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo B >> /tmp/hostvolume/"},
					Image:      "hello-word",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "a",
							Type:   genev1alpha1.DependTypeWhole,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumeb": {
							MountPath: "/tmp/hostvolume",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: "test-host-path",
							},
						},
					},
				},
				{
					Name:       "c",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo C >> /tmp/hostvolume/"},
					Image:      "hello-word",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "a",
							Type:   genev1alpha1.DependTypeWhole,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumec": {
							MountPath: "/tmp/hostvolume",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: "test-host-path",
							},
						},
					},
				},
				{
					Name:       "d",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo hello D"},
					Image:      "hello-word",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "b",
							Type:   genev1alpha1.DependTypeWhole,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumed": {
							MountPath: "/tmp/hostvolume",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: "test-host-path",
							},
						},
					},
				},
			},
		},
	}
	return &validateExecution
}

func TestValidateExecution(t *testing.T) {
	testCases := []struct {
		Name       string
		ModifyFunc ModifyExecution
		ExpectErr  bool
	}{
		{
			Name:       "validate execution",
			ModifyFunc: nil,
			ExpectErr:  false,
		},
		{
			Name: "dependents of execution exist cycle",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:       "a",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo A >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Dependents: []genev1alpha1.Dependent{
							{
								Target: "c",
								Type:   genev1alpha1.DependTypeWhole,
							},
						},
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
					{
						Name:       "b",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo B >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Dependents: []genev1alpha1.Dependent{
							{
								Target: "a",
								Type:   genev1alpha1.DependTypeWhole,
							},
						},
						Volumes: map[string]genev1alpha1.Volume{
							"volumeb": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
					{
						Name:       "c",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo C >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Dependents: []genev1alpha1.Dependent{
							{
								Target: "b",
								Type:   genev1alpha1.DependTypeWhole,
							},
						},
						Volumes: map[string]genev1alpha1.Volume{
							"volumec": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "tasks of execution must not be empty",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = nil
			},
			ExpectErr: true,
		},
		{
			Name: "parallelism must be greater than or equal to 0",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Parallelism = NewInt64(-10)
			},
			ExpectErr: true,
		},
		{
			Name: "task name must not be empty",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo A >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "task image must not be empty",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:       "a",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo A >> /tmp/hostvolume/"},
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "task commands must not be empty",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:  "a",
						Type:  genev1alpha1.JobTaskType,
						Image: "hello-word",
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "task parallelism must be greater than or equal to 0",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:        "a",
						Type:        genev1alpha1.JobTaskType,
						CommandSet:  []string{"echo A >> /tmp/hostvolume/"},
						Image:       "hello-word",
						Parallelism: NewInt64(-5),
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "task backoffLimit must be greater than or equal to 0",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:         "a",
						Type:         genev1alpha1.JobTaskType,
						CommandSet:   []string{"echo A >> /tmp/hostvolume/"},
						Image:        "hello-word",
						BackoffLimit: NewInt32(-5),
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "task activeDeadlineSeconds must be greater than or equal to 0",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:       "a",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo A >> /tmp/hostvolume/"},
						Image:      "hello-word",
						ActiveDeadlineSeconds: NewInt64(-5),
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "dependent target must not be empty",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:       "a",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo A >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Dependents: []genev1alpha1.Dependent{
							{
								Type: genev1alpha1.DependTypeWhole,
							},
						},
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "dependent target not exist",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:       "a",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo A >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Dependents: []genev1alpha1.Dependent{
							{
								Target: "c",
								Type:   genev1alpha1.DependTypeWhole,
							},
						},
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
		{
			Name: "wrong dependent type of task",
			ModifyFunc: func(exec *genev1alpha1.Execution) {
				exec.Spec.Tasks = []genev1alpha1.Task{
					{
						Name:       "a",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo A >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Dependents: []genev1alpha1.Dependent{
							{
								Target: "c",
								Type:   genev1alpha1.DependType("test"),
							},
						},
						Volumes: map[string]genev1alpha1.Volume{
							"volumea": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
					{
						Name:       "c",
						Type:       genev1alpha1.JobTaskType,
						CommandSet: []string{"echo C >> /tmp/hostvolume/"},
						Image:      "hello-word",
						Volumes: map[string]genev1alpha1.Volume{
							"volumec": {
								MountPath: "/tmp/hostvolume",
								MountFrom: genev1alpha1.VolumeSource{
									Pvc: "test-host-path",
								},
							},
						},
					},
				}
			},
			ExpectErr: true,
		},
	}

	for _, testCase := range testCases {
		exec := validateExecution()
		if testCase.ModifyFunc != nil {
			testCase.ModifyFunc(exec)
		}
		err := ValidateExecution(exec)
		if testCase.ExpectErr == true && err == nil {
			t.Errorf("%s: Expect error, but got nil", testCase.Name)
		}
		if testCase.ExpectErr == false && err != nil {
			t.Errorf("%s: Expect no error, but got error %v", testCase.Name, err)
		}
	}

}

func NewInt64(n int) *int64 {
	num := int64(n)
	return &num
}

func NewInt32(n int) *int32 {
	num := int32(n)
	return &num
}
