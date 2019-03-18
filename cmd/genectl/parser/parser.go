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
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	execv1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
)

func UnmarshalWorkflow(workflowData []byte) (*Workflow, error) {
	var workflow Workflow
	err := yaml.Unmarshal(workflowData, &workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflows: %v", err)
	}
	return &workflow, nil
}

// set default for workflows
func SetDefaultWorkflow(workflow *Workflow) {
	// set default for input
	SetDefaultInputs(workflow)
	// set default for other field
}

func ValidateWorkflow(workflow *Workflow) ErrorList {
	allErr := ErrorList{}
	if len(workflow.Jobs) <= 0 {
		errMsg := fmt.Sprint("No job defined in workflows")
		return append(allErr, fmt.Errorf(errMsg))
	}

	allErr = append(allErr, ValidateInputs(workflow.Inputs)...)

	for jobName, job := range workflow.Jobs {
		// validate job Name
		allErr = append(allErr, ValidateJobName(jobName)...)

		// validate resources
		allErr = append(allErr, ValidateResources(jobName, job.Resources)...)

		// validate tool
		allErr = append(allErr, ValidateTool(jobName, job.Tool)...)

		// validate commands
		allErr = append(allErr, ValidateCommands(jobName, job.Commands, workflow.Inputs)...)

		// validate commandsIter
		allErr = append(allErr, ValidateCommandsIter(jobName, job.CommandsIter, workflow.Inputs, workflow)...)

		// validate depends
		allErr = append(allErr, ValidateDepends(jobName, job.Depends, workflow.Jobs)...)
	}

	// detect cycle depends.
	if err := ValidateDependsCircle(workflow.Jobs); err != nil {
		allErr = append(allErr, err)
	}

	// validate volumes
	allErr = append(allErr, ValidateVolumes(workflow.Volumes, workflow.Inputs)...)

	// validate output
	for key, output := range workflow.Outputs {
		// validate paths
		allErr = append(allErr, ValidatePaths(key, output.Paths, workflow.Inputs)...)
		// validate pathsIter
		allErr = append(allErr, ValidatePathsIter(key, output.PathsIter, workflow.Inputs)...)
	}

	return allErr
}

func convert2ArrayOfIfs(data []Var) []interface{} {
	vars := make([]interface{}, 0)
	for _, arr := range data {
		for _, a := range arr {
			vars = append(vars, a)
		}
	}
	return vars
}

func InstantiateWorkflow(workflow *Workflow, inputs map[string]interface{}, tools map[string]Tool) error {
	// merge workflows input and input json.
	mergedInputs, err := MergeInputs(workflow.Inputs, inputs)
	if err != nil {
		return err
	}
	workflow.Inputs = mergedInputs

	inputsReplaceData := Inputs2ReplaceData(mergedInputs)

	// populate data for volumes.
	volumes := make(map[string]Volume, len(workflow.Volumes))
	for key, volume := range workflow.Volumes {
		volume.MountPath = ReplaceVariant(volume.MountPath, inputsReplaceData)
		volume.MountFrom.PVC = ReplaceVariant(volume.MountFrom.PVC, inputsReplaceData)
		volumes[key] = volume
	}
	workflow.Volumes = volumes

	// populate data for job.
	jobs := make(map[string]JobInfo, len(workflow.Jobs))

	for jobName, jobInfo := range workflow.Jobs {
		var tmpJob JobInfo

		tmpJob.Description = jobInfo.Description
		tmpJob.Tool = jobInfo.Tool

		tool, ok := tools[jobInfo.Tool]
		if !ok {
			return fmt.Errorf("workflows.%s.tool [%s] does not exist", jobName, jobInfo.Tool)
		}

		tmpJob.Image = tool.Image
		if len(jobInfo.Resources.Memory) != 0 {
			tmpJob.Resources.Memory = strings.ToUpper(jobInfo.Resources.Memory)
		}
		if len(jobInfo.Resources.Cpu) != 0 {
			tmpJob.Resources.Cpu = strings.ToUpper(jobInfo.Resources.Cpu)
		}
		if len(jobInfo.Commands) == 0 && IsCommandIterEmpty(jobInfo.CommandsIter) {
			tmpJob.Commands = append(tmpJob.Commands, tool.Command)
		}

		// populate data for commands
		newCommands := ReplaceArray(jobInfo.Commands, inputsReplaceData)

		// populate data for commandIter.vars
		prefix := fmt.Sprintf("***workflows.commands_iter.%s.vars", jobName)
		//fmt.Println(" before InstantiateVars", prefix, jobInfo.CommandsIter.Vars)
		vars, err := InstantiateVars(prefix, jobInfo.CommandsIter.Vars, inputsReplaceData)
		if err != nil {
			return err
		}
		//fmt.Println(" ****after InstantiateVars", vars)
		length, err := ValidateInstantiatedVars("workflows."+jobName, vars)
		if err != nil {
			return err
		}

		// populate data for commandIter.varsIter
		prefix = fmt.Sprintf("workflows.commands_iter.%s.varsIter", jobName)
		//fmt.Println(" before InstantiateVarsIter", prefix, jobInfo.CommandsIter.Vars)
		varsIter, dep, err := InstantiateVarsIter(prefix, jobInfo.CommandsIter.VarsIter, inputsReplaceData)
		if err != nil {
			return err
		}
		// if no get_result then len(dep) is zero
		if len(dep) == 0 {
			if length != 0 && len(varsIter) != 0 && len(varsIter) != length {
				return fmt.Errorf("workflows.%s: the length of vars is %d, but the length of varsIter is %d", jobName, length, len(varsIter))
			}

			// convert varsIter to var
			iterVars := VarIter2Vars(varsIter)

			// merge vars
			vars = append(vars, iterVars...)

			// populate data for CommandsIter.Command.
			command := ReplaceVariant(jobInfo.CommandsIter.Command, inputsReplaceData)

			// generate all commands.
			iterCommands := Iter2Array(command, vars)

			// merge jobInfo.commands and jobInfo.iterCommands
			newCommands = append(newCommands, iterCommands...)

			tmpJob.Commands = newCommands
			tmpJob.Depends = jobInfo.Depends
			jobs[jobName] = tmpJob

			//fmt.Println("tmpJob", tmpJob)
		} else {

			tmpJob.Commands = newCommands
			// populate data for CommandsIter.Command.
			command := ReplaceVariant(jobInfo.CommandsIter.Command, inputsReplaceData)

			tmpJob.CommandsIter.Command = command
			tmpJob.CommandsIter.Depends = dep
			tmpJob.CommandsIter.Vars = convert2ArrayOfIfs(vars)
			tmpJob.CommandsIter.VarsIter = convert2ArrayOfIfs(varsIter)

			tmpJob.Depends = jobInfo.Depends
			jobs[jobName] = tmpJob

		}
	}
	workflow.Jobs = jobs

	outPuts := make(map[string]OutputDesc, len(workflow.Outputs))
	for outputName, outputInfo := range workflow.Outputs {
		var output OutputDesc
		newPaths := ReplaceArray(outputInfo.Paths, inputsReplaceData)

		prefix := fmt.Sprintf("outputs.pathsIter.%s.vars", outputName)
		vars, err := InstantiateVars(prefix, outputInfo.PathsIter.Vars, inputsReplaceData)
		if err != nil {
			return err
		}
		length, err := ValidateInstantiatedVars("outputs."+outputName, vars)
		if err != nil {
			return err
		}

		prefix = fmt.Sprintf("outputs.pathsIter.%s.varsIter", outputName)
		varsIter, err := InstantiateVars(prefix, outputInfo.PathsIter.VarsIter, inputsReplaceData)
		if err != nil {
			return err
		}
		if length != 0 && len(varsIter) != 0 && len(varsIter) != length {
			return fmt.Errorf("outputs.%s: the length of varsIter is not equal to vars", outputName)
		}

		// convert varsIter to var
		varsIter = VarIter2Vars(varsIter)

		// merge vars
		vars = append(vars, varsIter...)

		// populate data for PathsIter.Path.
		path := ReplaceVariant(outputInfo.PathsIter.Path, inputsReplaceData)

		// generate all paths.
		iterPaths := Iter2Array(path, vars)

		newPaths = append(newPaths, iterPaths...)
		output.Paths = newPaths
		outPuts[outputName] = output
	}

	workflow.Outputs = outPuts

	return nil
}

func TransWorkflow2Execution(workflow *Workflow) (*execv1alpha1.Execution, error) {
	namespace := GetExecutionNamespace(workflow.Inputs)
	name := GetExecutionName(workflow.Inputs)
	execVolumes := TransVolume2ExecVolume(workflow.Volumes)

	// TODO make parallelism configurable
	parallelism := int64(5)

	exec := &execv1alpha1.Execution{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: execv1alpha1.ExecutionSpec{
			Parallelism: &parallelism,
			Tasks:       []execv1alpha1.Task{},
		},
	}

	for jobName, jobInfo := range workflow.Jobs {
		var task execv1alpha1.Task
		task.Name = jobName
		task.Type = "Job"
		task.Image = jobInfo.Image
		task.Volumes = execVolumes
		// we have alreay merge workflows command and commandIter.
		task.CommandSet = jobInfo.Commands

		//if job has get_result function in vars_iter
		if len(jobInfo.CommandsIter.Depends) > 0 {
			task.CommandsIter = TransCommandIter2ExecCommandIter(jobInfo.CommandsIter)
		}

		var cpuQuantity resource.Quantity
		var memoryQuantity resource.Quantity
		var err error

		// parse cpu
		if len(jobInfo.Resources.Cpu) > 0 {
			cpuNum := strings.TrimRight(jobInfo.Resources.Cpu, "cC")
			cpuQuantity, err = resource.ParseQuantity(cpuNum)
			if err != nil {
				return nil, fmt.Errorf("parse cpu quantity error: %v", err)
			}
		}
		// parse memory
		if len(jobInfo.Resources.Memory) > 0 {
			memoryQuantity, err = resource.ParseQuantity(jobInfo.Resources.Memory)
			if err != nil {
				return nil, fmt.Errorf("parse mem quantity error: %v", err)
			}
		}

		task.Resources = execv1alpha1.ResourceRequirements{
			Cpu:    cpuQuantity,
			Memory: memoryQuantity,
		}

		task.Dependents = TransDepend2ExecDepend(jobInfo.Depends)
		exec.Spec.Tasks = append(exec.Spec.Tasks, task)
	}

	return exec, nil
}
