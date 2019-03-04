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
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubegene.io/kubegene/cmd/genectl/client"
	execv1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
)

var descExecExample = `genectl describe execution my-exec –n gene-system`

type describeFlags struct {
	namespace string
}

func NewDescribeExecutionCommand() *cobra.Command {
	var describeFlags describeFlags

	var command = &cobra.Command{
		Use:     "describe execution NAME [flags]",
		Short:   "describe execution info",
		Args:    cobra.ExactArgs(2),
		Example: descExecExample,
		Run: func(cmd *cobra.Command, args []string) {
			DescribeWorkflow(cmd, args, &describeFlags)
		},
	}

	command.Flags().StringVarP(&describeFlags.namespace, "namespace", "n", "default", "workflow execution namespace")

	return command
}

func DescribeWorkflow(cmd *cobra.Command, args []string, describeFlags *describeFlags) {
	if args[0] != "execution" {
		ExitWithError(fmt.Errorf("first args of describe execution must be `execution` "))
	}
	executionName := args[1]
	if len(executionName) == 0 {
		ExitWithError(fmt.Errorf("executionName can not be empty"))
	}
	namespace := describeFlags.namespace

	// get exec client
	geneClient, err := client.GetGeneClient(cmd)
	if err != nil {
		ExitWithError(err)
	}

	// query exec
	exec, err := geneClient.ExecutionV1alpha1().Executions(namespace).Get(executionName, metav1.GetOptions{})
	if err != nil {
		ExitWithError(err)
	}

	DescribeExecution(exec)
}

func FindExecutionTask(exec *execv1alpha1.Execution, name string) *execv1alpha1.Task {
	for _, task := range exec.Spec.Tasks {
		if task.Name == name {
			return &task
		}
	}

	panic("can not get task for " + name)
}

func DescribeExecution(exec *execv1alpha1.Execution) {
	buf := &bytes.Buffer{}
	tabWriter := newTabWriter(buf)
	writer := NewExecutionWriter(tabWriter)

	writer.Write(0, "Name:\t%s\n", exec.Name)
	writer.Write(0, "Namespace:\t%s\n", exec.Namespace)

	printLabels(writer, exec.Labels)
	printAnnotations(writer, exec.Annotations)

	phase := "Running"
	if len(exec.Status.Phase) != 0 {
		phase = string(exec.Status.Phase)
	}
	writer.Write(0, "Phase:\t%s\n", phase)

	if len(exec.Status.Message) != 0 {
		writer.Write(0, "Message:\t%s\n", exec.Status.Message)
	}

	status := make(map[string][]execv1alpha1.VertexStatus)
	for _, task := range exec.Spec.Tasks {
		status[task.Name] = make([]execv1alpha1.VertexStatus, 0)
	}

	for key, vertex := range exec.Status.Vertices {
		// we use . to construct job name.
		items := strings.Split(key, ".")
		taskName := items[len(items)-2]
		if _, ok := status[taskName]; !ok {
			status[taskName] = make([]execv1alpha1.VertexStatus, 0)
		}
		status[taskName] = append(status[taskName], vertex)
	}

	writer.Write(0, "workflow:\n")

	for taskName, vertices := range status {
		task := FindExecutionTask(exec, taskName)
		totalJob := len(task.CommandSet)
		var succeedJob, failedJob, runningJob, errorJob int
		for _, vertex := range vertices {
			switch vertex.Phase {
			case execv1alpha1.VertexFailed:
				failedJob++
			case execv1alpha1.VertexError:
				errorJob++
			case execv1alpha1.VertexRunning:
				runningJob++
			case execv1alpha1.VertexSucceeded:
				succeedJob++
			}
		}

		info := fmt.Sprintf("(total: %d; success: %d; failed: %d; running: %d; error: %d)", totalJob, succeedJob, failedJob, runningJob, errorJob)
		writer.Write(1, taskName+info+":\n")

		if len(vertices) == 0 {
			if exec.Status.Phase == execv1alpha1.VertexError || exec.Status.Phase == execv1alpha1.VertexFailed {
				writer.Write(2, "the status of workflow is error or failed, this job will not run\n")
			} else {
				writer.Write(2, "wait for execute\n")
			}
		}

		writer.Write(2, "Subtask\tPhase\tMessage\n")
		writer.Write(2, "-------\t-----\t-------\n")

		for _, vertex := range vertices {
			writer.Write(2, "%v\t%v\t%v\n", vertex.Name, vertex.Phase, vertex.Message)
		}
	}

	tabWriter.Flush()
	str := string(buf.String())
	fmt.Fprintf(os.Stdout, "%s\n", str)
}

func newTabWriter(buf *bytes.Buffer) *tabwriter.Writer {
	out := new(tabwriter.Writer)
	out.Init(buf, 0, 8, 2, ' ', 0)
	return out
}

type ExecutionWriter struct {
	out io.Writer
}

func NewExecutionWriter(out io.Writer) ExecutionWriter {
	return ExecutionWriter{out: out}
}

func (pw *ExecutionWriter) Write(level int, format string, a ...interface{}) {
	levelSpace := "  "
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += levelSpace
	}
	fmt.Fprintf(pw.out, prefix+format, a...)
}

func (pw *ExecutionWriter) WriteLine(a ...interface{}) {
	fmt.Fprintln(pw.out, a...)
}

func printLabels(w ExecutionWriter, labels map[string]string) {
	innerIndent := "\t"
	w.Write(0, "%s:%s", "Labels", innerIndent)

	if labels == nil || len(labels) == 0 {
		w.WriteLine("<none>")
		return
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, key := range keys {
		if i != 0 {
			w.Write(0, "%s", innerIndent)
		}
		w.Write(0, "%s=%s\n", key, labels[key])
	}
}

var maxAnnotationLen = 140

func printAnnotations(w ExecutionWriter, annotations map[string]string) {
	innerIndent := "\t"

	w.Write(0, "%s:%s", "Annotations", innerIndent)

	if len(annotations) == 0 {
		w.WriteLine("<none>")
		return
	}

	keys := make([]string, 0, len(annotations))
	for key := range annotations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, key := range keys {
		if i != 0 {
			w.Write(0, innerIndent)
		}
		value := strings.TrimSuffix(annotations[key], "\n")
		if (len(value)+len(key)+2) > maxAnnotationLen || strings.Contains(value, "\n") {
			w.Write(0, "%s:\n", key)
			for _, s := range strings.Split(value, "\n") {
				w.Write(0, "%s  %s\n", innerIndent, shorten(s, maxAnnotationLen-2))
			}
		} else {
			w.Write(0, "%s: %s\n", key, value)
		}
	}
}

func shorten(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength] + "..."
	}
	return s
}
