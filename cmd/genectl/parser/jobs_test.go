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
	"strconv"
	"testing"
)

func TestValidateCPU(t *testing.T) {
	testCases := []struct {
		Name      string
		CPU       string
		ExpectErr bool
	}{
		{
			Name:      "valid CPU case1",
			CPU:       "8C",
			ExpectErr: false,
		},
		{
			Name:      "valid CPU case2",
			CPU:       "8c",
			ExpectErr: false,
		},
		{
			Name:      "valid CPU case3",
			CPU:       "18c",
			ExpectErr: false,
		},
		{
			Name:      "valid CPU case4",
			CPU:       "18.5c",
			ExpectErr: false,
		},
		{
			Name:      "valid CPU case5",
			CPU:       "18.5",
			ExpectErr: false,
		},
		{
			Name:      "invalid CPU case1",
			CPU:       "8G",
			ExpectErr: true,
		},
		{
			Name:      "invalid CPU case2",
			CPU:       "${cpu}G",
			ExpectErr: true,
		},
	}

	for i, testCase := range testCases {
		err := ValidateCPU("test"+strconv.Itoa(i), testCase.CPU)
		if testCase.ExpectErr == true && err == nil {
			t.Errorf("%s: Expect error, but got nil", testCase.Name)
		}
		if testCase.ExpectErr == false && err != nil {
			t.Errorf("%s: Expect no error, but got error %v", testCase.Name, err)
		}
	}
}

func TestValidateMemory(t *testing.T) {
	testCases := []struct {
		Name      string
		Memory    string
		ExpectErr bool
	}{
		{
			Name:      "valid Memory case1",
			Memory:    "8G",
			ExpectErr: false,
		},
		{
			Name:      "valid Memory case2",
			Memory:    "8g",
			ExpectErr: false,
		},
		{
			Name:      "valid Memory case3",
			Memory:    "18g",
			ExpectErr: false,
		},
		{
			Name:      "valid Memory case4",
			Memory:    "18.5g",
			ExpectErr: false,
		},
		{
			Name:      "valid Memory case5",
			Memory:    "18.5",
			ExpectErr: false,
		},
		{
			Name:      "invalid Memory case1",
			Memory:    "8C",
			ExpectErr: true,
		},
		{
			Name:      "invalid Memory case2",
			Memory:    "${mem}C",
			ExpectErr: true,
		},
	}

	for i, testCase := range testCases {
		err := ValidateMemory("test"+strconv.Itoa(i), testCase.Memory)
		if testCase.ExpectErr == true && err == nil {
			t.Errorf("%s: Expect error, but got nil", testCase.Name)
		}
		if testCase.ExpectErr == false && err != nil {
			t.Errorf("%s: Expect no error, but got error %v", testCase.Name, err)
		}
	}
}

func TestValidateResources(t *testing.T) {
	testCases := []struct {
		Name         string
		Res          Resources
		ExpectErrNum int
	}{
		{
			Name:         "valid Res case1",
			Res:          Resources{Cpu: "8C", Memory: "8G"},
			ExpectErrNum: 0,
		},
		{
			Name:         "valid Res case2",
			Res:          Resources{Memory: "8G"},
			ExpectErrNum: 0,
		},
		{
			Name:         "valid Res case3",
			Res:          Resources{Cpu: "8c"},
			ExpectErrNum: 0,
		},
		{
			Name:         "invalid Res case1",
			Res:          Resources{Cpu: "8Cc", Memory: "8G"},
			ExpectErrNum: 1,
		},
		{
			Name:         "invalid Res case2",
			Res:          Resources{Cpu: "8Cc", Memory: "8Gg"},
			ExpectErrNum: 2,
		},
	}

	for i, testCase := range testCases {
		errList := ValidateResources("test"+strconv.Itoa(i), testCase.Res)
		if testCase.ExpectErrNum != len(errList) {
			t.Errorf("%s: Expect error number %d, but got %d", testCase.Name, testCase.ExpectErrNum, len(errList))
		}
	}
}

func TestValidateDepend(t *testing.T) {
	testCases := []struct {
		Name         string
		Depend       Depend
		Jobs         map[string]JobInfo
		ExpectErrNum int
	}{
		{
			Name:         "valid depend",
			Depend:       Depend{Target: "bar", Type: "whole"},
			Jobs:         map[string]JobInfo{"bar": {Image: "nginx"}},
			ExpectErrNum: 0,
		},
		{
			Name:         "depend target should not be a variant",
			Depend:       Depend{Target: "${bar}", Type: "whole"},
			Jobs:         map[string]JobInfo{"bar": {Image: "nginx"}},
			ExpectErrNum: 1,
		},
		{
			Name:         "depend type should not be a variant",
			Depend:       Depend{Target: "bar", Type: "${whole}"},
			Jobs:         map[string]JobInfo{"bar": {Image: "nginx"}},
			ExpectErrNum: 2,
		},
		{
			Name:         "depend target does not exist",
			Depend:       Depend{Target: "bar", Type: "whole"},
			Jobs:         map[string]JobInfo{"bar2": {Image: "nginx"}},
			ExpectErrNum: 1,
		},
		{
			Name:         "depend type should only be whole or iterator",
			Depend:       Depend{Target: "bar", Type: "xxxx"},
			Jobs:         map[string]JobInfo{"bar": {Image: "nginx"}},
			ExpectErrNum: 1,
		},
	}

	for i, testCase := range testCases {
		errList := ValidateDepend("test"+strconv.Itoa(i), testCase.Depend, testCase.Jobs)
		if testCase.ExpectErrNum != len(errList) {
			t.Errorf("%s: Expect error number %d, but got %d", testCase.Name, testCase.ExpectErrNum, len(errList))
		}
	}
}

func TestValidateDependsCircle(t *testing.T) {
	testCases := []struct {
		Jobs      map[string]JobInfo
		ExpectErr bool
	}{
		{
			Jobs: map[string]JobInfo{
				"job-gatk": {
					Tool:    "GATK:4.0.1",
					Depends: []Depend{{Target: "job-bwa", Type: "iterate"}},
				},
				"job-bwa": {
					Tool:    "bwa:0.71",
					Depends: []Depend{{Target: "job-zsplit", Type: "whole"}},
				},
				"job-zsplit": {
					Tool:    "zsplit:0.2",
					Depends: []Depend{{Target: "job-gatk", Type: "iterate"}},
				},
			},
			ExpectErr: true,
		},
		{
			Jobs: map[string]JobInfo{
				"job-gatk": {
					Tool:    "GATK:4.0.1",
					Depends: []Depend{{Target: "job-bwa", Type: "iterate"}},
				},
				"job-bwa": {
					Tool:    "bwa:0.71",
					Depends: []Depend{{Target: "job-zsplit", Type: "whole"}},
				},
				"job-zsplit": {
					Tool: "zsplit:0.2",
				},
			},
			ExpectErr: false,
		},
	}

	for i, testCase := range testCases {
		err := ValidateDependsCircle(testCase.Jobs)
		if testCase.ExpectErr == true && err == nil {
			t.Errorf("%d: Expect error, but got nil", i)
		}
		if testCase.ExpectErr == false && err != nil {
			t.Errorf("%d: Expect no error, but got error %v", i, err)
		}
	}
}

func TestValidateJobName(t *testing.T) {
	testCases := []struct {
		JobName   string
		ExpectErr bool
	}{
		{
			JobName:   "job-gatk",
			ExpectErr: false,
		},
		{
			JobName:   "job-abcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			ExpectErr: true,
		},
	}

	for i, testCase := range testCases {
		err := ValidateJobName(testCase.JobName)
		if testCase.ExpectErr == true && len(err) == 0 {
			t.Errorf("%d: Expect error, but got nil", i)
		}
		if testCase.ExpectErr == false && len(err) != 0 {
			t.Errorf("%d: Expect no error, but got error %v", i, err)
		}
	}
}

func TestValidateTool(t *testing.T) {
	testCases := []struct {
		ToolName  string
		ExpectErr bool
	}{
		{
			ToolName:  "gatk:4.2",
			ExpectErr: false,
		},
		{
			ExpectErr: true,
		},
		{
			ToolName:  "${gatk}",
			ExpectErr: true,
		},
	}

	for i, testCase := range testCases {
		err := ValidateTool("test", testCase.ToolName)
		if testCase.ExpectErr == true && len(err) == 0 {
			t.Errorf("%d: Expect error, but got nil", i)
		}
		if testCase.ExpectErr == false && len(err) != 0 {
			t.Errorf("%d: Expect no error, but got error %v", i, err)
		}
	}
}

func TestValidateCommands(t *testing.T) {
	testCases := []struct {
		Commands  []string
		Inputs    map[string]Input
		ExpectErr bool
	}{
		{
			Commands: []string{
				`./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample1.bam ${lost_input} -O ${obs-path}/sample1_output.g.vcf --spark-runner SPARK --spark-master yarn-client`,
				`./gatk HaplotypeCallerSpark -R ${obs-path}/ref.2bit -I ${obs-path}/sample2.bam -O ${obs-path}/sample2_output.g.vcf  --spark-runner SPARK --spark-master yarn-client`,
			},
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
	}

	for i, testCase := range testCases {
		err := ValidateCommands("test", testCase.Commands, testCase.Inputs)
		if testCase.ExpectErr == true && len(err) == 0 {
			t.Errorf("%d: Expect error, but got nil", i)
		}
		if testCase.ExpectErr == false && len(err) != 0 {
			t.Errorf("%d: Expect no error, but got error %v", i, err)
		}
	}
}

func TestValidateCommandsIter(t *testing.T) {
	testCases := []struct {
		Name         string
		CommandsIter string
		Inputs       map[string]Input
		ExpectErr    bool
	}{
		{
			Name: "valid commandsIter",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - - hello
      vars_iter:
      - - aaa`,
			Inputs:    makeInputs(),
			ExpectErr: false,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.commandsIter.command: variant [lost_input] undefine",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam ${lost_input}
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - - hello
      vars_iter:
      - - aaa`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.commandsIter.vars[0][0]: the variant [sample] is not defined in the inputs",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - - ${sample}
      vars_iter:
      - - aaa`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.commandsIter.vars[0][0]: the value type should not be array",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - - [a, b]
      vars_iter:
      - - aaa`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.commandsIter.vars[0][0]: the type of bwaArray can only be in [number string bool], but the real one is array",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - - ${bwaArray}
      vars_iter:
      - - aaa`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.vars or commandsIter is empty but command is not empty",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.vars or commandsIter is not empty but command is empty",
			CommandsIter: `
      vars:
      - - bbb
      vars_iter:
      - - aaa`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.commandsIter.vars[0]:the element of vars array should only be array variant or array, but the real one is bbb",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam ${2}
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - bbb
      vars_iter:
      - - aaa`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},

		// #########################################################################################################

		{
			Name: "workflow.job-gatk.commandsIter.vars[0]:the element of vars array should only be array variant or array, but the real one is 2",
			CommandsIter: `
      command: sh ${obs-path}/scripts/step3.gatkspark.sh I ${obs-path}/${1}.bam ${2}
        -O ${obs-path}/${1}_output.g.vcf  --spark-runner SPARK --spark-master yarn-client
      vars:
      - 2`,
			Inputs:    makeInputs(),
			ExpectErr: true,
		},
	}

	for i, testCase := range testCases {
		var commandsIter CommandsIter
		yaml.Unmarshal([]byte(testCase.CommandsIter), &commandsIter)
		err := ValidateCommandsIter("test", commandsIter, testCase.Inputs, nil)
		if testCase.ExpectErr == true && len(err) == 0 {
			t.Errorf("%d: Expect error, but got nil", i)
		}
		if testCase.ExpectErr == false && len(err) != 0 {
			t.Errorf("%d: Expect no error, but got error %v", i, err)
		}
	}
}
