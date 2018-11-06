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
	"io/ioutil"
	"testing"
)

var version = `version: genecontainer_0_1`

func makeTools() map[string]Tool {
	gatk := Tool{
		Name:        "GATK",
		Version:     "4.0.1",
		Image:       "1.0.0.21:/root/GATK:4.0.1",
		Cpu:         "1c",
		Memory:      "2G",
		Description: "GATK",
	}
	bwa := Tool{
		Name:    "bwa",
		Version: "0.71r",
		Image:   "1.0.0.21:/root/bwa:0.71r",
	}
	zsplit := Tool{
		Name:    "zsplit",
		Version: "0.2",
		Image:   "1.0.0.21:/root/zsplit:0.2",
	}

	tools := make(map[string]Tool)
	tools[gatk.Name+":"+gatk.Version] = gatk
	tools[bwa.Name+":"+bwa.Version] = bwa
	tools[zsplit.Name+":"+zsplit.Version] = zsplit

	return tools
}

func makeInputs() map[string]Input {
	inputs := map[string]Input{
		"npart": {
			Default:     2,
			Description: "split job n part",
			Type:        NumberType,
		},
		"obs-path": {
			Default:     "/root",
			Description: "obs path",
			Type:        StringType,
		},
		"bwaArray": {
			Default:     []int{1, 2, 3},
			Description: "array type",
			Type:        ArrayType,
		},
		"isOk": {
			Default:     true,
			Description: "bool type",
			Type:        BoolType,
		},
		"gcs-pvc": {
			Default:     "claim1",
			Description: "cliam name",
			Type:        StringType,
		},
	}

	return inputs
}

var inputs = `
inputs:
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
    description: "bool type"
`

var workflows = `
workflow:
  job-gatk:
    description: ""
    tool: GATK:4.0.1
    resources:
      memory: 8G
      cpu: 2c
    image: ""
    commands:
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam
      -O ${obs-path}/sample1_output.g.vcf --spark-runner SPARK --spark-master yarn-client
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample2.bam
      -O ${obs-path}/sample2_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - - hello
      vars_iter:
      - - aaa
`

var volumes = `
volumes:
  reference:
    mountPath: /root
    mountFrom:
      pvc: ${GCS_REF_PVC}
  sample-data:
    mountPath: /mnt
    mountFrom:
      pvc: ${GCS_DATA_PVC}
  temp-data:
    mountPath: /tmp
    mountFrom:
      pvc: ${GCS_SFS_PVC}
`

var outputs = `
outputs:
  gene-report-vcf:
    paths:
    - /saa/123/ddd.vcf
    pathsIter:
      path: ${obs-path}/${1}/merge.HaplotypeCaller.raw.vcf
      vars_iter:
      - - aaa
`

func TestValidateWorkflow(t *testing.T) {
	data, _ := ioutil.ReadFile("../../../example/gatk4-best-practices/gatk4-best-practices.yaml")
	workflow, err := UnmarshalWorkflow(data)
	if err != nil {
		t.Fatalf("unmarshal workflow err: %v", err)
	}

	errList := ValidateWorkflow(workflow)
	if len(errList) > 0 {
		t.Fatalf("expect no error, but got %v", errList)
	}
}

func TestInstantiateJobCommands(t *testing.T) {
	workflowStr := `
workflow:
  job-gatk:
    type: CCE.Job
    tool: GATK:4.0.1
    commands:
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam
`
	gcs := version + inputs + workflowStr + volumes

	workflow, _ := UnmarshalWorkflow([]byte(gcs))
	err := InstantiateWorkflow(workflow, nil, makeTools())
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	result, _ := json.Marshal(workflow.Jobs["job-gatk"].Commands)
	expectResult := `["./gatk HaplotypeCallerSpark -R /root/ref.2bit -I /root/sample1.bam"]`

	if string(result) != expectResult {
		t.Errorf("expect result %v, but got %v", expectResult, string(result))
	}
}

func TestInstantiateJobCommandsIter(t *testing.T) {
	testCases := []struct {
		Name         string
		WorkflowStr  string
		ExpectResult string
		ExpectErr    bool
	}{
		{
			Name: "valid case 1",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands:
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam ${2}
      vars_iter:
      - [aaa, bbb]
      - [ccc, ddd]`,
			ExpectResult: `["./gatk HaplotypeCallerSpark -R /root/ref.2bit -I /root/sample1.bam","sh /root/scripts/step3.gatkspark.sh I /root/aaa.bam ccc","sh /root/scripts/step3.gatkspark.sh I /root/aaa.bam ddd","sh /root/scripts/step3.gatkspark.sh I /root/bbb.bam ccc","sh /root/scripts/step3.gatkspark.sh I /root/bbb.bam ddd"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 2",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands:
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars:
      - - hello`,
			ExpectResult: `["./gatk HaplotypeCallerSpark -R /root/ref.2bit -I /root/sample1.bam","sh /root/scripts/step3.gatkspark.sh I /root/hello.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 3",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands:
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars:
      - - hello
      vars_iter:
      - - aaa`,
			ExpectResult: `["./gatk HaplotypeCallerSpark -R /root/ref.2bit -I /root/sample1.bam","sh /root/scripts/step3.gatkspark.sh I /root/hello.bam","sh /root/scripts/step3.gatkspark.sh I /root/aaa.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 4",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands:
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars:
      - - ${obs-path}`,
			ExpectResult: `["./gatk HaplotypeCallerSpark -R /root/ref.2bit -I /root/sample1.bam","sh /root/scripts/step3.gatkspark.sh I /root//root.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 5",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars_iter:
      - ${bwaArray}`,
			ExpectResult: `["sh /root/scripts/step3.gatkspark.sh I /root/1.bam","sh /root/scripts/step3.gatkspark.sh I /root/2.bam","sh /root/scripts/step3.gatkspark.sh I /root/3.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 6",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars_iter:
      - range(1,3)`,
			ExpectResult: `["sh /root/scripts/step3.gatkspark.sh I /root/1.bam","sh /root/scripts/step3.gatkspark.sh I /root/2.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 7",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars_iter:
      - range(0,${npart})`,
			ExpectResult: `["sh /root/scripts/step3.gatkspark.sh I /root/0.bam","sh /root/scripts/step3.gatkspark.sh I /root/1.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 8",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars_iter:
      - range(0,${npart}, 2)`,
			ExpectResult: `["sh /root/scripts/step3.gatkspark.sh I /root/0.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 9",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${item}  ${obs-path}/${1}.bam
      vars:
      - [2]
      - [1]`,
			ExpectResult: `["sh /root/scripts/step3.gatkspark.sh I 0  /root/2.bam","sh /root/scripts/step3.gatkspark.sh I 1  /root/1.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "valid case 10",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${1} ${obs-path}/${-1}.bam
      vars:
      - [2]`,
			ExpectResult: `["sh /root/scripts/step3.gatkspark.sh I 2 /root/${-1}.bam"]`,
			ExpectErr:    false,
		},

		// ####################################################################################################

		{
			Name: "workflow.commands_iter.job-gatk.vars_iter[0]:In range function, start should be smaller than end",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars_iter:
      - range(0,0)`,
			ExpectErr: true,
		},

		// ####################################################################################################

		{
			Name: "workflow.commands_iter.job-gatk.vars_iter[0]:In range function, step should be larger than 0",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars_iter:
      - range(0,2,-1)`,
			ExpectErr: true,
		},

		// ####################################################################################################

		{
			Name: "workflow.job-gatk.vars: the length of 0 line is 2, but the length of 1 line is 1",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
      vars:
      - [1, 2]
      - [1]`,
			ExpectErr: true,
		},

		// ####################################################################################################

		{
			Name: "workflow.job-gatk: the length of vars is 2, but the length of vars_iter is 1",
			WorkflowStr: `
workflow:
  job-gatk:
    description: ""
    type: CCE.Job
    tool: GATK:4.0.1
    image: ""
    commands:
    - ./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam
    commands_iter:
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam ${2}
      vars:
      - - hello
        - world
      vars_iter:
      - - aaa`,
			ExpectErr: true,
		},
	}

	for _, testCase := range testCases {
		gcs := version + inputs + testCase.WorkflowStr + volumes

		workflow, _ := UnmarshalWorkflow([]byte(gcs))
		err := InstantiateWorkflow(workflow, nil, makeTools())
		if testCase.ExpectErr == true && err == nil {
			t.Errorf("%s: Expect error, but got nil", testCase.Name)
		}
		if testCase.ExpectErr == false && err != nil {
			t.Errorf("%s: Expect no error, but got error %v", testCase.Name, err)
		}
		if err != nil {
			continue
		}

		result, _ := json.Marshal(workflow.Jobs["job-gatk"].Commands)
		if string(result) != testCase.ExpectResult {
			t.Errorf("%s: Expect result %v, but got %v", testCase.Name, testCase.ExpectResult, string(result))
		}
	}
}
