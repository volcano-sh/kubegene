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
	"os"
	"path/filepath"
)

const (
	Kubegene = "kubegene"
	// CLIName is the name of the CLI
	CLIName = "genectl"
)

// User home
var Home = os.Getenv("HOME")

// default tool repository directory
var ToolDir = filepath.Join(Home, Kubegene, "tools")

// NewCommand returns a new instance of an gcs command
func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   CLIName,
		Short: "A simple command line client for kubegene.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.PersistentFlags().String("kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")

	command.AddCommand(NewSubCommand())
	command.AddCommand(NewDeleteCommand())
	command.AddCommand(NewDescribeExecutionCommand())
	command.AddCommand(NewGetExecutionCommand())
	command.AddCommand(NewVersionCommand())

	return command
}
