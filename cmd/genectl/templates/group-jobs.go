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

package templates

const (
	GroupJobsTpl = `version: genecontainer_0_1
inputs:
  memory:
    default: {{.Memory}}
    type: string
  cpu:
    default: {{.Cpu}}
    type: number
  tool:
    default: {{.Tool}}
    type: string
  job-script:
    default: {{.JobScript}}
    type: string
  mount-path:
    default: {{.MountPath}}
    type: string
  pvc-name:
    default: {{.PvcName}}
    type: string
  jobid:
    default: {{.JobId}}
    type: string

workflow:
  {{.JobId}}:
    tool: {{.Tool}}
    resources:
      memory: {{.Memory}}
      cpu: {{.Cpu}}c
    commands:
      {{- range .Commands}}
      - {{.}}
      {{- end}}
volumes:
  sample-data:
    mount_path: ${mount-path}
    mount_from:
      pvc: ${pvc-name}
`
)
