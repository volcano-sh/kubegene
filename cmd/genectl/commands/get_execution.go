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
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"kubegene.io/kubegene/cmd/genectl/client"
	"kubegene.io/kubegene/cmd/genectl/util"
	execv1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
)

var getExecutionExample = `
		# List executions in default namespace and in default output format.
		genectl get execution

		# List all executions in default output format.
		genectl get execution --all-namespaces

		# List all executions in yaml output format.
		genectl get execution --all-namespaces -o yaml

		# List executions that are running or succeeded in yaml output format.
		genectl get execution --all-namespaces –phase Running,Succeeded

		# List executions in exec-system namespace that are running or succeeded in yaml output format.
		genectl get execution -n exec-system –phase Running,Succeeded`

type getExecutionFlags struct {
	namespace     string
	allNamespaces bool
	phases        []string
	output        string
}

func NewGetExecutionCommand() *cobra.Command {
	var getExecutionFlags getExecutionFlags

	var command = &cobra.Command{
		Use:   "get",
		Short: "get execution",
		Run: func(cmd *cobra.Command, args []string) {
			getExecution(cmd, args, &getExecutionFlags)
		},
	}

	command.Flags().StringVarP(&getExecutionFlags.namespace, "namespace", "n", "default", "workflow execution namespace")
	command.Flags().StringSliceVar(&getExecutionFlags.phases, "phase", getExecutionFlags.phases, fmt.Sprintf("A comma-separated list for workflow execution phase. Available values: %v. This flag unset means 'list all phase workflow'", getAllPhases()))
	command.Flags().BoolVar(&getExecutionFlags.allNamespaces, "all-namespaces", getExecutionFlags.allNamespaces, "If present, list execution across all namespaces.")
	command.Flags().StringVarP(&getExecutionFlags.output, "output", "o", "wide", "Output format. One of: json|yaml|wide, default wide")

	return command
}

func getExecution(cmd *cobra.Command, args []string, getExecutionFlags *getExecutionFlags) {
	if len(args) == 0 {
		ExitWithError(fmt.Errorf("get execution need at least one args"))
	}

	if args[0] != "execution" && args[0] != "executions" {
		ExitWithError(fmt.Errorf("first args of get execution must be `execution` or `executions`"))
	}

	execNames := []string{}
	if len(args) > 1 {
		execNames = append(execNames, args[1:]...)
	}

	namespace := getExecutionFlags.namespace

	// get exec client
	geneClient, err := client.GetGeneClient(cmd)
	if err != nil {
		ExitWithError(err)
	}

	executions := []execv1alpha1.Execution{}
	if len(execNames) > 0 {
		if len(getExecutionFlags.phases) != 0 {
			ExitWithError(fmt.Errorf("phase flag should be set when query some certain executions"))
		}

		for _, name := range execNames {
			execution, err := geneClient.ExecutionV1alpha1().Executions(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				ExitWithError(err)
			}
			executions = append(executions, *execution)
		}
	} else {
		if getExecutionFlags.allNamespaces {
			namespace = metav1.NamespaceAll
		}
		// query exec list
		execList, err := geneClient.ExecutionV1alpha1().Executions(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			ExitWithError(fmt.Errorf("list execution error: %v", err))
		}

		filterExecList := make([]execv1alpha1.Execution, 0)
		if len(getExecutionFlags.phases) != 0 {
			for _, exec := range execList.Items {
				if checkPhase(string(exec.Status.Phase), getExecutionFlags.phases) {
					filterExecList = append(filterExecList, exec)
				}
			}
		} else {
			filterExecList = execList.Items
		}

		sort.Sort(PhaseOrder(filterExecList))

		executions = filterExecList
	}

	PrintExecutionList(executions, getExecutionFlags.output)
}

func PrintExecutionList(execList []execv1alpha1.Execution, output string) {
	switch output {
	case "wide":
		PrintExecutionListWide(execList)
	case "json":
		util.PrintJSON(execList)
	case "yaml":
		util.PrintYAML(execList)
	default:
		fmt.Printf("unsupported format: %v\n", output)
	}
}

func PrintExecutionListWide(execList []execv1alpha1.Execution) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)
	fmt.Fprint(out, "Name\tAge\tPhase\tMessage\n")
	fmt.Fprint(out, "----\t---\t-----\t-------\n")

	for _, exec := range execList {
		age := duration.HumanDuration(time.Since(exec.CreationTimestamp.Time))
		fmt.Fprintf(out, "%v\t%v\t%v\t%v\n", exec.Name, age, exec.Status.Phase, exec.Status.Message)
	}

	out.Flush()
	str := string(buf.String())
	fmt.Fprintf(os.Stdout, "%s\n", str)
}

func getAllPhases() []execv1alpha1.VertexPhase {
	return []execv1alpha1.VertexPhase{
		execv1alpha1.VertexRunning,
		execv1alpha1.VertexSucceeded,
		execv1alpha1.VertexFailed,
		execv1alpha1.VertexError,
	}
}

func checkPhase(phase string, phases []string) bool {
	for _, item := range phases {
		if item == phase {
			return true
		}
	}

	return false
}

type PhaseOrder []execv1alpha1.Execution

func (s PhaseOrder) Len() int      { return len(s) }
func (s PhaseOrder) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s PhaseOrder) Less(i, j int) bool {
	iphase := s[i].Status.Phase
	jphase := s[j].Status.Phase
	return iphase > jphase
}
