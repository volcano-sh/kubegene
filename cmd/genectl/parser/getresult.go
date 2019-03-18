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

package parser

import (
	"fmt"
	"regexp"
	"strings"
)

const IsGetResultFuncRegexFmt = `^get_result\(\s*([^,]+)\s*(,\s*("([^,]+)"|'([^,]+)'|\$\{[^,]+\}))?\)$`

var getResultRegExp = regexp.MustCompile(IsGetResultFuncRegexFmt)
var inputsVarRegExp = regexp.MustCompile("\\$\\{[^,]+\\}")

func decodeNonPrintChar(sep string) string {

	sep = strings.Replace(sep, "\\n", "\n", -1)
	sep = strings.Replace(sep, "\\t", "\t", -1)
	sep = strings.Replace(sep, "\\r", "\r", -1)

	return sep
}

// IsGetResultFunc checks a string whether is a get_result function.
// get_result function in the workflows must follow the format:
// 		get_result(jobName, sep)
// The jobName of the Job
// The separator used to split the string:
// 	Can be a single character or a string;
// 	Use double quotes to indicate, such as "\n";
// 	Variables such as ${input} can be used.
//
// get_result function example
//
// ---- get_result(job-a, "\n")
// ---- get_result(job-target)
// ---- get_result(job-target, ${input})
func IsGetResultFunc(str string) bool {
	if matched := getResultRegExp.MatchString(str); matched {
		return true
	}
	return false
}

// GetgetResultFuncParam extract parameter from range function.
// for example:
//
// 		Str ---> get_result(job-a,"\n")
// 		result ---> job-a,  "\n"
//
// 		Str ---> get_result(job-target,sep )
// 		result ---> job-target,sep
func GetgetResultFuncParam(str string) (jobName string, sep string) {
	submatch := getResultRegExp.FindStringSubmatch(str)

	for i := 0; i < len(submatch); i++ {
		fmt.Println("submatch ", i, submatch[i])
	}

	jobName = strings.Replace(submatch[1], " ", "", -1)

	fmt.Println("jobName ", jobName)

	if matched := inputsVarRegExp.MatchString(str); matched {
		return
	}

	if submatch[4] != "" {
		sep = submatch[4]
	} else {
		sep = submatch[5]
	}

	fmt.Println("Before sep ", sep)
	sep = decodeNonPrintChar(sep)
	fmt.Println("After sep ", sep)
	return
}

func isJobExists(jobName string, workflow *Workflow) bool {
	_, ok := workflow.Jobs[jobName]
	return ok
}

func validatedependecy(prefix string, jobName string, dependjobName string, workflow *Workflow) error {

	dependJob, ok := workflow.Jobs[dependjobName]
	if !ok {
		err := fmt.Errorf("%s: the get_result function dependecy job is missing, but the real one is %s", prefix, dependjobName)
		return err
	}

	//depend job should have single command only because it should be single k8s- job related to that job

	if (len(dependJob.Commands) > 1) || (len(dependJob.CommandsIter.Vars) > 1) || (len(dependJob.CommandsIter.VarsIter) > 1) {
		err := fmt.Errorf("the get_result function dependecy job has more than one command %s dependjobName :%s", prefix, dependjobName)
		return err
	}

	currentJob, ok := workflow.Jobs[jobName]
	if !ok {
		err := fmt.Errorf("%s: the get_result function dependecy job is missing, but the real one is %s", prefix, dependjobName)
		return err
	}
	for i := 0; i < len(currentJob.Depends); i++ {
		if (currentJob.Depends[i].Target == dependjobName) &&
			(currentJob.Depends[i].Type) == "whole" {
			return nil
		}
	}

	err := fmt.Errorf("%s: the get_result function dependecy job type is wrong %s", prefix, dependjobName)

	return err
}

// validategetResultFunc validate parameter of get_result function is valid.
func validategetResultFunc(prefix, str string, inputs map[string]Input, jobName string, workflow *Workflow) ErrorList {
	allErr := ErrorList{}
	dependjobName, _ := GetgetResultFuncParam(str)
	if isJobExists(jobName, workflow) {
		err := fmt.Errorf("%s: the get_result function dependecy job is missing, but the real one is %s", prefix, dependjobName)
		allErr = append(allErr, err)
	}
	// validate the dependency
	err := validatedependecy(prefix, jobName, dependjobName, workflow)
	if err != nil {
		allErr = append(allErr, err)
	}
	return allErr
}
func InstantiategetResultFunc(prefix, str string, data map[string]string) (Var, map[string]bool, error) {
	dependsResult := map[string]bool{}
	jobName, sep := GetgetResultFuncParam(str)

	getresult := []interface{}{"get_result", jobName, sep}

	dependsResult[jobName] = true
	return getresult, dependsResult, nil
}
