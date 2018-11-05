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
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/uuid"
	"kubegene.io/kubegene/cmd/genectl/templates"
	"kubegene.io/kubegene/cmd/genectl/util"
)

var subJobExample = `
		# Submit a single job from a file with specify resources and tools
        genectl sub job /kubegene/bwa_help.sh --memory 1g --cpu 1 --tool bwa:0.7.12 --pvc pvc-gene`

var subJobLong = `
		sub job command submits a job which execute a single shell script when
		perform gene sequencing. You should upload the shell script and sample data to
		the volume used by this job in preparation stage . The args[0] FILENAME is the
		absolute path of the shell script within the container.`

// default workflow yaml directory
var WorkflowFilePath = filepath.Join(Home, Kubegene, "workflows")

// UserInputs parameters
type UserInputs struct {
	Memory string
	Cpu    string
	Tool   string
	Shell  string
	// the absolute path of the shell script within the container.
	JobScript string
	PvcName   string
	MountPath string
	JobId     string
	Commands  []string
}

func NewSubJobCommand() *cobra.Command {
	var userInputs UserInputs
	var command = &cobra.Command{
		Use:     "job FILENAME",
		Short:   "submit a single job",
		Long:    subJobLong,
		Example: subJobExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			SubJob(cmd, args, &userInputs)
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

func SubJob(cmd *cobra.Command, args []string, userInputs *UserInputs) {
	myJobScript := args[0]
	userInputs.JobScript = myJobScript

	if len(userInputs.MountPath) == 0 {
		userInputs.MountPath = path.Dir(myJobScript)
	}

	// validate the input parameter.
	if err := validateInputs(userInputs); err != nil {
		ExitWithError(fmt.Errorf("check UserInputs failed: %v", err))
	}

	// generate jobid
	jobIdPrefix := strings.Replace(util.GetFileNameOnly(userInputs.JobScript), "_", "-", -1)
	jobId := generateJobId(jobIdPrefix)
	userInputs.JobId = jobId

	dstFile := jobId + ".yaml"

	// parse template.
	err := ParseTemplate(templates.SingleJobPpl, path.Join(WorkflowFilePath, dstFile), userInputs)
	if err != nil {
		ExitWithError(fmt.Errorf("parse template err %v", err))
	}

	ProcessWorkflow(cmd, path.Join(WorkflowFilePath, dstFile), nil)
}

func validateInputs(userInputs *UserInputs) error {
	if len(userInputs.Tool) == 0 {
		return fmt.Errorf("tool must be specified")
	}
	if len(userInputs.PvcName) == 0 {
		return fmt.Errorf("pvc must be specified")
	}

	return nil
}

func generateJobId(jobIdPrefix string) string {
	// time format YYYY-MMDD-HHMMSS
	formatTime := time.Now().Format("2006-0102-150405")

	uuid := uuid.NewUUID()
	randStr := strings.Replace(string(uuid), "-", "", -1)[0:5]

	jobId := fmt.Sprintf("%s-%s-%s", jobIdPrefix, formatTime, randStr)
	return jobId
}

func ParseTemplate(content string, dstFile string, userInputs *UserInputs) error {
	dstFile, err := filepath.Abs(dstFile)
	if err != nil {
		return err
	}

	// create directory if not exist.
	util.CreateDirectory(filepath.Dir(dstFile))

	file, err := os.OpenFile(
		dstFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	t := template.Must(template.New("job").Parse(content))

	return t.Execute(file, userInputs)
}
