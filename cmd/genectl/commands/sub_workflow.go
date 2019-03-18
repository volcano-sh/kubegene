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

package commands

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	"bytes"
	"github.com/renstrom/dedent"
	"html/template"
	"kubegene.io/kubegene/cmd/genectl/client"
	"kubegene.io/kubegene/cmd/genectl/parser"
	"kubegene.io/kubegene/cmd/genectl/util"
)

var submitSuccessMessage = template.Must(template.New("message").Parse(dedent.Dedent(`
		The workflow has been submitted successfully! And the execution has been created.
		Your can use the follow command to query the status of workflow execution.

			genectl get execution {{ .name }} -n {{ .namespace }}

		or use the follow command to query the detail info for workflow execution.

			genectl describe execution {{ .name }} -n {{ .namespace }}

		or use the follow command to delete the workflow execution.

			genectl delete execution {{ .name }} -n {{ .namespace }}

		`)))

var subWorkflowExample = `
		# Submit a workflow from a file with specify UserInputs"
		gcs sub workflow wf.yaml --input UserInputs.json`

type workflowFlags struct {
	input string
}

func NewSubWorkflowCommand() *cobra.Command {
	var workflowFlags workflowFlags
	var command = &cobra.Command{
		Use:   "workflow FILENAME ...",
		Short: "submit a workflow",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			SubWorkflow(cmd, args, &workflowFlags)
		},
	}

	command.Flags().StringVar(&workflowFlags.input, "input", "", "the input json file path.")

	return command
}

func SubWorkflow(cmd *cobra.Command, args []string, workflowFlags *workflowFlags) {
	// read input json
	inputs, err := readInputJson(workflowFlags.input)
	if err != nil {
		ExitWithError(fmt.Errorf("read input json file %s failed: %v", workflowFlags.input, err))
	}

	ProcessWorkflow(cmd, args[0], inputs)
}

func ProcessWorkflow(cmd *cobra.Command, workflowPath string, inputs map[string]interface{}) {
	// fetch all usable tools.
	tools, err := parser.FetchTools(cmd)
	if err != nil {
		ExitWithError(err)
	}

	// read workflow
	data, err := ioutil.ReadFile(workflowPath)
	if err != nil {
		ExitWithError(fmt.Errorf("read workflowFile %s failed: %v", workflowPath, err))
	}

	workflow, err := parser.UnmarshalWorkflow(data)
	if err != nil {
		ExitWithError(err)
	}

	// set default for workflow
	parser.SetDefaultWorkflow(workflow)

	// validate workflow
	errList := parser.ValidateWorkflow(workflow)
	if len(errList) > 0 {
		PrintErrList(errList)
		os.Exit(1)
	}
	// instantiate workflow
	err = parser.InstantiateWorkflow(workflow, inputs, tools)
	if err != nil {
		ExitWithError(err)
	}
	if util.GetFlagBool(cmd, "dry-run") {
		util.PrintYAML(workflow)
		return
	}

	// trans workflow to execution
	execution, err := parser.TransWorkflow2Execution(workflow)
	if err != nil {
		ExitWithError(err)
	}

	// get exec client
	geneClient, err := client.GetGeneClient(cmd)
	if err != nil {
		ExitWithError(err)
	}

	// submit execution to api server.
	newExec, err := geneClient.ExecutionV1alpha1().Executions(execution.Namespace).Create(execution)
	if err != nil {
		ExitWithError(fmt.Errorf("submit execution error: %v", err))
	}

	ctx := map[string]string{
		"name":      newExec.GetName(),
		"namespace": newExec.GetNamespace(),
	}

	var msg bytes.Buffer
	submitSuccessMessage.Execute(&msg, ctx)
}

func readInputJson(inputFile string) (map[string]interface{}, error) {
	if inputFile == "" {
		return nil, nil
	}
	content := make(map[string]interface{})
	inputJson, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inputJson, &content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func PrintErrList(errList parser.ErrorList) {
	for i, err := range errList {
		fmt.Printf("error %d: %v\n", i+1, err)
	}
}
