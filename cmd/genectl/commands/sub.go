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
	"github.com/spf13/cobra"
)

type subOptions struct {
	output string
	dryRun bool
}

func NewSubCommand() *cobra.Command {

	var subOptions subOptions

	var command = &cobra.Command{
		Use:   "sub",
		Short: "submit a workflow",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.PersistentFlags().String("tool-repo", ToolDir, "directory or URL to tool repository, if it is a URL, it must point to tool file.")
	command.PersistentFlags().BoolVarP(&subOptions.dryRun, "dry-run", "", false, "If true, display results but do not submit workflow")

	command.AddCommand(NewSubJobCommand())
	command.AddCommand(NewSubRepJobCommand())
	command.AddCommand(NewSubWorkflowCommand())

	return command
}
