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

	"kubegene.io/kubegene/pkg/common"
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
	return getResultRegExp.MatchString(str)
}

// getResultFuncParam extract parameter from get_result function.
// for example:
//
// 		Str ---> get_result(job-a,"\n")
// 		result ---> job-a,  "\n"
//
// 		Str ---> get_result(job-target,sep )
// 		result ---> job-target,sep
func getResultFuncParam(str string) (jobName string, sep string) {

	submatch := getResultRegExp.FindStringSubmatch(str)

	jobName = strings.Replace(submatch[1], " ", "", -1)

	if submatch[4] != "" {
		sep = submatch[4]
	} else {
		sep = submatch[5]
	}
	sep = decodeNonPrintChar(sep)
	return
}

func isJobExists(jobName string, workflow *Workflow) bool {
	_, ok := workflow.Jobs[jobName]
	return ok
}

func validateDependency(prefix string, jobName string, dependJobName string, workflow *Workflow) error {

	dependJob, ok := workflow.Jobs[dependJobName]
	if !ok {
		err := fmt.Errorf("%s: the get_result function dependecy job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	// depend job should have single command only because it should be single k8s- job related to that job

	if (len(dependJob.Commands) > 1) || (len(dependJob.CommandsIter.Vars) > 1) || (len(dependJob.CommandsIter.VarsIter) > 1) {
		err := fmt.Errorf("the get_result function dependecy job has more than one command %s dependjobName :%s", prefix, dependJobName)
		return err
	}

	currentJob, ok := workflow.Jobs[jobName]
	if !ok {
		err := fmt.Errorf("%s: the get_result function  job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	if len(currentJob.Depends) != 1 {
		err := fmt.Errorf("%s: the get_result  job has more dependecies %v", prefix, currentJob.Depends)
		return err
	}

	for i := 0; i < len(currentJob.Depends); i++ {
		if (currentJob.Depends[i].Target == dependJobName) &&
			(currentJob.Depends[i].Type) == "whole" {
			return nil
		}
	}

	err := fmt.Errorf("%s: the get_result function dependecy job type is wrong %s", prefix, dependJobName)

	return err
}

// validateGetResultFunc validate parameter of get_result function is valid.
func validateGetResultFunc(prefix, str string, inputs map[string]Input, jobName string, workflow *Workflow) ErrorList {
	allErr := ErrorList{}
	dependJobName, sep := getResultFuncParam(str)

	if !isJobExists(jobName, workflow) {
		err := fmt.Errorf("%s: the get_result function dependecy job is missing, but the real one is %s", prefix, dependJobName)
		allErr = append(allErr, err)
	}

	if sep != "" {
		if IsVariant(sep) {
			if err := ValidateVariant(prefix, sep, []string{StringType}, inputs); err != nil {
				allErr = append(allErr, err)
			}
		}
	}
	// validate the dependency
	err := validateDependency(prefix, jobName, dependJobName, workflow)
	if err != nil {
		allErr = append(allErr, err)
	}
	return allErr
}

func InstantiateGetResultFunc(prefix, str string, data map[string]string) common.Var {

	jobName, sep := getResultFuncParam(str)
	// replace variant for sep
	sep = common.ReplaceVariant(sep, data)
	getresult := []interface{}{"get_result", jobName, sep}

	return getresult
}
