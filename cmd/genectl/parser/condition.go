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

const IsCheckResultFuncRegexFmt = `^check_result\(\s*([^,]+)\s*(,\s*("([^,]+)"|'([^,]+)'|\$\{[^,]+\}))?\)$`

var checkResultRegExp = regexp.MustCompile(IsCheckResultFuncRegexFmt)
var inputsChkVarRegExp = regexp.MustCompile("\\$\\{[^,]+\\}")

// IsCheckResultFunc checks a string whether is a check_result function.
// check_result function in the workflows must follow the format:
// 		check_result(jobName, exp)
// The jobName of the Job
// The exp used to determine if the output of jobName is equal or not:
// 	Can be a single character or a string;
// 	Use double quotes to indicate, such as "\n";
// 	Variables such as ${input} can be used.
//
// check_result function example
//
// ---- check_result(job-a, "\n")
// ---- check_result(job-target, ${input})
func IsCheckResultFunc(str string) bool {
	return checkResultRegExp.MatchString(str)
}

// checkResultFuncParam extract parameter from check_result function.
// for example:
//
// 		Str ---> check_result(job-a,"\n")
// 		result ---> job-a,  "\n"
//
// 		Str ---> check_result(job-target,exp )
// 		result ---> job-target,exp
func checkResultFuncParam(str string) (jobName string, exp string) {

	submatch := checkResultRegExp.FindStringSubmatch(str)

	jobName = strings.Replace(submatch[1], " ", "", -1)

	if submatch[4] != "" {
		exp = submatch[4]
	} else {
		exp = submatch[5]
	}
	exp = decodeNonPrintChar(exp)
	return
}

func validateCheckResultDependency(prefix string, jobName string, dependJobName string, workflow *Workflow) error {

	dependJob, ok := workflow.Jobs[dependJobName]
	if !ok {
		err := fmt.Errorf("%s: the check_result function dependecy job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	// depend job should have single command only because it should be single k8s- job related to that job

	if (len(dependJob.Commands) > 1) || (len(dependJob.CommandsIter.Vars) > 1) || (len(dependJob.CommandsIter.VarsIter) > 1) {
		err := fmt.Errorf("the check_result function dependecy job has more than one command %s dependjobName :%s", prefix, dependJobName)
		return err
	}

	currentJob, ok := workflow.Jobs[jobName]
	if !ok {
		err := fmt.Errorf("%s: the check_result function  job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	if len(currentJob.Depends) != 1 {
		err := fmt.Errorf("%s: the check_result  job has more dependecies %v", prefix, currentJob.Depends)
		return err
	}

	for i := 0; i < len(currentJob.Depends); i++ {
		if (currentJob.Depends[i].Target == dependJobName) &&
			(currentJob.Depends[i].Type) == "whole" {
			return nil
		}
	}

	err := fmt.Errorf("%s: the check_result function dependecy job type is wrong %s", prefix, dependJobName)

	return err
}

// validateCheckResultFunc validate parameter of check_result function is valid.
func validateCheckResultFunc(prefix, str string, inputs map[string]Input, jobName string, workflow *Workflow) ErrorList {
	allErr := ErrorList{}
	dependJobName, exp := checkResultFuncParam(str)

	if !isJobExists(jobName, workflow) {
		err := fmt.Errorf("%s: the check_result function job is missing, but the real one is %s", prefix, dependJobName)
		allErr = append(allErr, err)
	}

	if IsVariant(exp) {
		if err := ValidateVariant(prefix, exp, []string{StringType}, inputs); err != nil {
			allErr = append(allErr, err)
		}
	}
	// validate the dependency
	err := validateCheckResultDependency(prefix, jobName, dependJobName, workflow)
	if err != nil {
		allErr = append(allErr, err)
	}
	return allErr
}

// validateStringCondition validate parameter of condition which is string is valid.
func validateStringCondition(prefix string, condition string, inputs map[string]Input, jobName string, workflow *Workflow) ErrorList {
	allErr := ErrorList{}
	if IsVariant(condition) {
		if err := ValidateVariant(prefix, condition, []string{StringType}, inputs); err != nil {
			allErr = append(allErr, err)
		}
	} else if IsCheckResultFunc(condition) {
		allErr = validateCheckResultFunc(prefix, condition, inputs, jobName, workflow)
	} else {
		err := fmt.Errorf("In validateStringCondition Invalid condition string %v", condition)
		allErr = append(allErr, err)
	}
	return allErr
}

// validateCondition validate parameter of condition is valid.
func validateCondition(jobName string, condition interface{}, inputs map[string]Input, workflow *Workflow) ErrorList {
	allErr := ErrorList{}

	if condition == nil {
		return allErr
	}

	switch condition.(type) {

	case bool:
		return allErr
	case string:
		prefix := fmt.Sprintf("workflow.%s.condition", jobName)
		return validateStringCondition(prefix, condition.(string), inputs, jobName, workflow)
	default:
		err := fmt.Errorf("In validateCondition Invalid condition parameter %v", condition)
		allErr = append(allErr, err)
	}

	return allErr
}

func InstantiateCondition(prefix string, condition interface{}, data map[string]string) (common.Var, error) {

	if condition == nil {
		return nil, nil
	}

	switch condition.(type) {
	case bool:
		return []interface{}{condition}, nil
	case string:
		str := condition.(string)
		if IsVariant(str) {
			output := common.ReplaceVariant(str, data)
			if output == "true" {
				return []interface{}{true}, nil

			} else if output == "false" {
				return []interface{}{false}, nil
			} else {
				err := fmt.Errorf("Invalid data in the condition %v", condition)
				return nil, err
			}
		} else {
			ret := InstantiateCheckResultFunc(prefix, condition.(string), data)
			return ret, nil
		}
	default:
		err := fmt.Errorf("Invalid data in the condition %v", condition)
		return nil, err

	}
}

func InstantiateCheckResultFunc(prefix, str string, data map[string]string) common.Var {

	jobName, exp := checkResultFuncParam(str)
	// replace variant for sep
	exp = common.ReplaceVariant(exp, data)
	chkResult := []interface{}{"check_result", jobName, exp}

	return chkResult
}
