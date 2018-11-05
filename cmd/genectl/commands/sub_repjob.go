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
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"kubegene.io/kubegene/cmd/genectl/templates"
	"kubegene.io/kubegene/cmd/genectl/util"
)

var subRepJobExample = `
		# Submit a set of jobs from a file with specify resources and tools"
        genectl sub repjob /kubegene/bwa_mem_work.sh --memory 1g --cpu 1 --tool bwa:0.7.12 --pvc pvc-gene`

var subRepJobLong = `
		sub repjob command submits a group of job. The args[0] FILENAME is the
		absolute path of the shell script within the container. And every line
		in the shell script is a single job and it follow the format:

			bash/sh             scriptPath                args...
			   |                    |                        |
			linux shell  abs path within the container    shell args

		for example:

			genectl sub repjob /kubegene/bwa_mem_work.sh --memory 1g --cpu 1 --tool bwa:0.7.12 --pvc pvc-gene

		/kubegene/bwa_mem_work.sh is the the absolute path of the shell script within the container.
		the content of bwa_mem_work.sh:

			sh /kubegene/bwa_mem.sh obs/path/sample1.fastq.gz obs/path/hg19.fa >obs/path/sample1.sam
			sh /kubegene/bwa_mem.sh obs/path/sample2.fastq.gz obs/path/hg19.fa >obs/path/sample2.sam

		And the script path in your host should keep consistent with the path within the container.
		You should upload all the shell script and sample data that will be used to the storage volume used by this job. `

const GROUPJOBSCRPTCOMMAND string = `^\s*(bash|sh)\s*.*$`

func NewSubRepJobCommand() *cobra.Command {
	var userInputs UserInputs
	var command = &cobra.Command{
		Use:   "repjob FILENAME ...",
		Short: "submit a set of jobs",
		Long:  subRepJobLong,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			SubRepJob(cmd, args, &userInputs)
		},
	}

	command.Flags().StringVar(&userInputs.Memory, "memory", "1G", "memory resource required to run this job, default 1G")
	command.Flags().StringVar(&userInputs.Cpu, "cpu", "1", "cpu resource required to run this job, default 1")
	command.Flags().StringVar(&userInputs.Shell, "shell", "sh", "linux shell used to execute the job script, default sh")
	command.Flags().StringVar(&userInputs.Tool, "tool", "", "tool used by the job, format: toolName:toolVersion")
	command.Flags().StringVar(&userInputs.PvcName, "pvc", "", "the name of pvc that used by the job, the backend storage that the pvc refer to should pre-populate the job script and sample data used by the job")
	command.Flags().StringVar(&userInputs.MountPath, "mount-path", "", "path within the container at which the volume that pvc refer to should be mounted, default the same as the host dir of job script")

	return command
}

func SubRepJob(cmd *cobra.Command, args []string, userInputs *UserInputs) {
	userInputs.JobScript = args[0]

	// validate the input parameter.
	if err := validateInputs(userInputs); err != nil {
		ExitWithError(fmt.Errorf("check UserInputs failed %v", err))
	}

	if len(userInputs.MountPath) == 0 {
		userInputs.MountPath = path.Dir(userInputs.JobScript)
	}

	// generate jobid
	jobIdPrefix := strings.Replace(util.GetFileNameOnly(userInputs.JobScript), "_", "-", -1)
	jobId := generateJobId(jobIdPrefix)
	userInputs.JobId = jobId

	//解析commands命令
	var commands []string
	err := validateScript(userInputs.JobScript, &commands)
	if err != nil {
		ExitWithError(fmt.Errorf("validate script failed %v", err))
	}
	userInputs.Commands = commands

	dstFile := jobId + ".yaml"

	// parse template.
	err = ParseTemplate(templates.GroupJobsTpl, path.Join(WorkflowFilePath, dstFile), userInputs)
	if err != nil {
		ExitWithError(fmt.Errorf("%v", err))
	}

	ProcessWorkflow(cmd, path.Join(WorkflowFilePath, dstFile), nil)
}

func validateScript(script string, commands *[]string) error {
	file, err := os.Open(script)
	if err != nil {
		return fmt.Errorf("open script %s failed %v", script, err)
	}
	defer file.Close()

	br := bufio.NewReader(file)
	for {
		bytes, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		command := string(bytes)
		if regexp.MustCompile(GROUPJOBSCRPTCOMMAND).MatchString(command) {
			//bash xx/xx/xx.sh args...
			lists := strings.Fields(command)
			if len(lists) < 2 {
				return fmt.Errorf("invalid command %s", command)
			}
			scriptPath := lists[1]
			if !filepath.IsAbs(scriptPath) {
				return fmt.Errorf("script path should be absolute path in command %s", command)
			}
			*commands = append(*commands, command)
		}
	}

	return nil
}
