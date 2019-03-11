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
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubegene.io/kubegene/cmd/genectl/client"
)

var delExecExample = `genectl delete execution my-exec –n gene-system`

type delExecutionFlags struct {
	namespace string
}

func NewDeleteCommand() *cobra.Command {
	var delExecutionFlags delExecutionFlags

	var command = &cobra.Command{
		Use:     "delete execution NAME [flags]",
		Short:   "delete a execution",
		Args:    cobra.ExactArgs(2),
		Example: delExecExample,
		Run: func(cmd *cobra.Command, args []string) {
			DeleteWorkflow(cmd, args, &delExecutionFlags)
		},
	}

	command.Flags().StringVarP(&delExecutionFlags.namespace, "namespace", "n", "default", "workflow execution namespace")

	return command
}

func DeleteWorkflow(cmd *cobra.Command, args []string, delExecutionFlags *delExecutionFlags) {
	if args[0] != "execution" && args[0] != "executions" {
		ExitWithError(fmt.Errorf("first args of del execution must be `execution` or `executions` "))
	}
	executionName := args[1]
	if len(executionName) == 0 {
		ExitWithError(fmt.Errorf("executionName can not be empty"))
	}
	namespace := delExecutionFlags.namespace

	// get exec client
	geneClient, err := client.GetGeneClient(cmd)
	if err != nil {
		ExitWithError(err)
	}

	// delete exec
	err = geneClient.ExecutionV1alpha1().Executions(namespace).Delete(executionName, &metav1.DeleteOptions{})
	if err != nil {
		ExitWithError(err)
	}

	fmt.Printf("delete execution %v successfully\n", executionName)
}
