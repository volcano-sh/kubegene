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
	execv1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	"regexp"
)

const CPURegexFmt = `^\d+(\.\d+)?[cC]?$`

func ValidateCPU(prefix, cpu string) error {
	if matched, _ := regexp.MatchString(CPURegexFmt, cpu); matched {
		return nil
	}
	return fmt.Errorf("%s.cpu %s is illegal", prefix, cpu)
}

const MemoryRegexFmt = `^\d+(\.\d+)?[gG]?$`

func ValidateMemory(prefix, memory string) error {
	if matched, _ := regexp.MatchString(MemoryRegexFmt, memory); matched {
		return nil
	}
	return fmt.Errorf("%s.memory %s is illegal", prefix, memory)
}

func ValidateResources(jobName string, res Resources) ErrorList {
	errors := ErrorList{}
	prefix := fmt.Sprintf("workflow.%s.resources", jobName)
	if len(res.Cpu) != 0 {
		if err := ValidateCPU(prefix, res.Cpu); err != nil {
			errors = append(errors, err)
		}
	}
	if len(res.Memory) != 0 {
		if err := ValidateMemory(prefix, res.Memory); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func ValidateDepend(prefix string, depend Depend, jobs map[string]JobInfo) ErrorList {
	errors := ErrorList{}
	if IsVariant(depend.Type) {
		errors = append(errors, fmt.Errorf("%s.type should not be variant", prefix))
	}
	if depend.Type != IterateDependType && depend.Type != WholeDependType {
		errors = append(errors, fmt.Errorf("%s.type should only be iterate or whole", prefix))
	}
	if IsVariant(depend.Target) {
		return append(errors, fmt.Errorf("%s.target should not be a variant", prefix))
	}
	if _, ok := jobs[depend.Target]; !ok {
		errors = append(errors, fmt.Errorf("%s.target [%s] not exist", prefix, depend.Target))
	}
	return errors
}

func ValidateDepends(jobName string, depends []Depend, jobs map[string]JobInfo) ErrorList {
	errors := ErrorList{}
	for i, depend := range depends {
		prefix := fmt.Sprintf("workflow.%s.depends[%d]", jobName, i)
		errors = append(errors, ValidateDepend(prefix, depend, jobs)...)
	}
	return errors
}
func validateGenericDependency(prefix string, jobName string, dependJobName string, workflow *Workflow) error {

	dependJob, ok := workflow.Jobs[dependJobName]
	if !ok {
		err := fmt.Errorf("%s: the dependecy job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	// depend job should have single command only because it should be single k8s- job related to that job

	if (len(dependJob.Commands) > 1) || (len(dependJob.CommandsIter.Vars) > 1) || (len(dependJob.CommandsIter.VarsIter) > 1) {
		err := fmt.Errorf("the dependecy job has more than one command %s dependjobName :%s", prefix, dependJobName)
		return err
	}

	currentJob, ok := workflow.Jobs[jobName]
	if !ok {
		err := fmt.Errorf("%s: the job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	if len(currentJob.Depends) != 1 {
		err := fmt.Errorf("%s: the job has more dependecies %v", prefix, currentJob.Depends)
		return err
	}

	for i := 0; i < len(currentJob.Depends); i++ {
		if (currentJob.Depends[i].Target == dependJobName) &&
			(currentJob.Depends[i].Type) == "whole" {
			return nil
		}
	}

	err := fmt.Errorf("%s: the dependecy job type is wrong %s", prefix, dependJobName)

	return err
}

func ValidateGenericCondition(jobName string, gCondition *GenericCondition, inputs map[string]Input, workflow *Workflow) ErrorList {
	allErrs := ErrorList{}

	if gCondition == nil {
		return allErrs
	}
	prefix := fmt.Sprintf("workflow.%s.genericcondition.dependjobname", jobName)
	err := validateGenericDependency(prefix, jobName, gCondition.DependJobName, workflow)
	if err != nil {
		allErrs = append(allErrs, err)
	}

	for i := range gCondition.MatchRules {
		prefix := fmt.Sprintf("workflow.%s.genericcondition.matchrules[%d]", jobName, i)
		switch gCondition.MatchRules[i].Operator {
		case MatchOperatorOpIn, MatchOperatorOpNotIn, MatchOperatorOpEqual,
			MatchOperatorOpDoubleEqual, MatchOperatorOpNotEqual:
			if len(gCondition.MatchRules[i].Values) == 0 {
				allErrs = append(allErrs, fmt.Errorf("%s.values must be specified when `operator` is 'In' or 'NotIn'", prefix))
			}
		case MatchOperatorOpExists, MatchOperatorOpDoesNotExist:
			if len(gCondition.MatchRules[i].Values) > 0 {
				allErrs = append(allErrs, fmt.Errorf("%s.values must not be specified when `operator` is 'Exists' or 'DoesNotExist' ", prefix))
			}
		case MatchOperatorOpGt, MatchOperatorOpLt:
			if len(gCondition.MatchRules[i].Values) != 1 {
				allErrs = append(allErrs, fmt.Errorf("%s.values must be specified single value when `operator` is 'Lt' or 'Gt' ", prefix))
			}
		default:
			allErrs = append(allErrs, fmt.Errorf("%s.values not a valid Match operator ", prefix))

		}
		if len(gCondition.MatchRules[i].Key) == 0 || IsVariant(gCondition.MatchRules[i].Key) {
			allErrs = append(allErrs, fmt.Errorf("%s.key should not be empty or shoud not be variant ", prefix))
		}
	}
	return allErrs
}

func ValidateTool(jobName string, toolName string) ErrorList {
	errors := ErrorList{}
	if len(toolName) == 0 {
		err := fmt.Errorf("workflow.%s.tool should not be empty", jobName)
		return append(errors, err)
	}

	if IsVariant(toolName) {
		err := fmt.Errorf("workflow.%s.tool should not be variant", jobName)
		return append(errors, err)
	}

	return errors
}

func ValidateCommands(jobName string, commands []string, inputs map[string]Input) ErrorList {
	allErr := ErrorList{}
	for i, command := range commands {
		prefix := fmt.Sprintf("workflow.%s.commands[%d]", jobName, i)
		_, errors := ValidateTemplate(command, prefix, "command", inputs)
		allErr = append(allErr, errors...)
	}
	return allErr
}

func ValidateCommandsIter(jobName string, commandsIter CommandsIter, inputs map[string]Input, workflow *Workflow) ErrorList {
	allError := ErrorList{}
	if len(commandsIter.Command) == 0 && IsCommandIterEmpty(commandsIter) {
		return allError
	}
	if len(commandsIter.Command) == 0 && !IsCommandIterEmpty(commandsIter) {
		err := fmt.Errorf("workflow.%s.vars or vars_iter is not empty but command is empty", jobName)
		return append(allError, err)
	}
	if len(commandsIter.Command) != 0 && IsCommandIterEmpty(commandsIter) {
		err := fmt.Errorf("workflow.%s.vars or vars_iter is empty but command is not empty", jobName)
		return append(allError, err)
	}

	prefix := fmt.Sprintf("workflow.%s.commands_iter.command", jobName)
	maxIndex, errors := ValidateTemplate(commandsIter.Command, prefix, "command", inputs)
	allError = append(allError, errors...)

	if maxIndex > len(commandsIter.VarsIter) {
		err := fmt.Errorf("workflow.%s.commands_iter.command: ${%d} is larger than commands_iter's rows", jobName, maxIndex)
		allError = append(allError, err)
	}

	prefix = fmt.Sprintf("workflow.%s.commands_iter.vars", jobName)
	allError = append(allError, ValidateVarsArray(prefix, commandsIter.Vars, inputs)...)

	prefix = fmt.Sprintf("workflow.%s.commands_iter.vars_iter", jobName)
	allError = append(allError, ValidateVarsIterArray(prefix, commandsIter.VarsIter, inputs, jobName, workflow)...)

	return allError
}

func printCircle(circle []string, start int) string {
	str := ""
	for i := start; i < len(circle); i++ {
		str = str + circle[i] + "->"
	}
	return str + circle[start]
}

func DFS(node string, graphNodes map[string]bool, graphEdges map[string][]string, nodeStack []string) error {
	graphNodes[node] = true
	nodeStack = append(nodeStack, node)
	for _, targetNode := range graphEdges[node] {
		if graphNodes[targetNode] == false {
			if err := DFS(targetNode, graphNodes, graphEdges, nodeStack); err != nil {
				return err
			}
		} else {
			index := sliceContain(nodeStack, targetNode)
			if index != -1 {
				return fmt.Errorf("detect circle from depend relationship. Circle:%s", printCircle(nodeStack, index))
			}
		}
	}
	return nil
}

func ValidateDependsCircle(jobs map[string]JobInfo) error {
	graphNodes := make(map[string]bool, len(jobs))
	graphEdge := make(map[string][]string)
	for jobName := range jobs {
		graphNodes[jobName] = false
	}
	for jobName, job := range jobs {
		for _, depend := range job.Depends {
			graphEdge[jobName] = append(graphEdge[jobName], depend.Target)
		}
	}
	for node := range graphNodes {
		if graphNodes[node] {
			continue
		}
		if err := DFS(node, graphNodes, graphEdge, nil); err != nil {
			return err
		}
	}
	return nil
}

func IsCommandIterEmpty(commandIter CommandsIter) bool {
	if len(commandIter.Vars) == 0 && len(commandIter.VarsIter) == 0 {
		return true
	}
	return false
}

const JobNameRegexFmt = "^[a-z]([-a-z0-9]*[a-z0-9])?$"

func ValidateJobName(jobName string) ErrorList {
	errors := ErrorList{}
	if len(jobName) > 40 {
		err := fmt.Errorf("workflow.%s: job Name is more than 40 characters", jobName)
		errors = append(errors, err)
	}
	if matched, _ := regexp.MatchString(JobNameRegexFmt, jobName); !matched {
		err := fmt.Errorf("workflow.%s: job Name is illegal, it must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character", jobName)
		errors = append(errors, err)
	}
	return errors
}

func TransDepend2ExecDepend(depends []Depend) []execv1alpha1.Dependent {
	execDepends := []execv1alpha1.Dependent{}
	for _, depend := range depends {
		var tmpDependent execv1alpha1.Dependent
		tmpDependent.Type = execv1alpha1.DependType(depend.Type)
		tmpDependent.Target = depend.Target
		execDepends = append(execDepends, tmpDependent)
	}
	return execDepends
}

func TransCommandIter2ExecCommandIter(commandsIter CommandsIter) *execv1alpha1.CommandsIter {
	var execCommandIter execv1alpha1.CommandsIter
	execCommandIter.Command = commandsIter.Command
	execCommandIter.VarsIter = make([]interface{}, 0)

	for _, var1 := range commandsIter.VarsIter {
		execCommandIter.VarsIter = append(execCommandIter.VarsIter, var1)
	}
	return &execCommandIter
}

func TransCond2ExecCond(condition interface{}) *execv1alpha1.Condition {
	execCond := execv1alpha1.Condition{}

	execCond.Condition = condition

	return &execCond
}

func TransGenericCond2ExecGenericCond(condition *GenericCondition) *execv1alpha1.GenericCondition {
	execCond := execv1alpha1.GenericCondition{}

	execCond.DependJobName = condition.DependJobName

	for i := range condition.MatchRules {
		var req execv1alpha1.MatchRule
		req.Key = condition.MatchRules[i].Key

		switch condition.MatchRules[i].Operator {

		case MatchOperatorOpIn:
			req.Operator = execv1alpha1.MatchOperatorOpIn
		case MatchOperatorOpNotIn:
			req.Operator = execv1alpha1.MatchOperatorOpNotIn
		case MatchOperatorOpExists:
			req.Operator = execv1alpha1.MatchOperatorOpExists
		case MatchOperatorOpDoesNotExist:
			req.Operator = execv1alpha1.MatchOperatorOpDoesNotExist
		case MatchOperatorOpGt:
			req.Operator = execv1alpha1.MatchOperatorOpGt
		case MatchOperatorOpLt:
			req.Operator = execv1alpha1.MatchOperatorOpLt
		case MatchOperatorOpEqual:
			req.Operator = execv1alpha1.MatchOperatorOpEqual
		case MatchOperatorOpNotEqual:
			req.Operator = execv1alpha1.MatchOperatorOpNotEqual
		case MatchOperatorOpDoubleEqual:
			req.Operator = execv1alpha1.MatchOperatorOpDoubleEqual
		}

		for j := range condition.MatchRules[i].Values {
			req.Values = append(req.Values, condition.MatchRules[i].Values[j])
		}
		execCond.MatchRules = append(execCond.MatchRules, req)
	}

	return &execCond
}
