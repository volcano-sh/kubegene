/*
Copyright 2019 The Kubegene Authors.

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

	"kubegene.io/kubegene/pkg/version"
)

var versionExample = `genectl version`

func NewVersionCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "version",
		Short:   "Print the version information",
		Long:    "Print the version information",
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			versionInfo(cmd, args)
		},
	}
	return command
}

func versionInfo(cmd *cobra.Command, args []string) {
	version := version.GetVersion()
	fmt.Printf("  genectl Version: %s\n", version)
}
