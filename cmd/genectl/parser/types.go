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

package parser

const (
	StringType = "string"
	NumberType = "number"
	BoolType   = "bool"
	ArrayType  = "array"
)

var InputTypeList = []string{StringType, NumberType, BoolType, ArrayType}

// Tool is an abstraction of the gene sequencing container. It contains the basic information
// about a gene sequencing container. such as image, version, description and so on. When we
// use it in gene sequencing workflows, we can simply specify the tool as Name:version, and the
// real image and other information will be replaced when running.
//
// tool example
//
// Name: GATK
// version: 4.0.1
// image: 1.0.0.21:/root/GATK:4.0.1
// command: gatk hello world
// type: basic
// description: software package to analyze next-generation sequencing data
//
// use example
//
// job-GATK:
//   tool: GATK:4.0.1
//   resources:
//     memory: 2G
//     cpu: 2C
//   command:
//     - sh ${obs-path}/${jobid}/bwa_mem.sh obs/path/sample1.fastq.gz obs/path/hg19.fa >obs/path/sample1.sam
//     - sh ${obs-path}/${jobid}/bwa_mem.sh obs/path/sample1.fastq.gz obs/path/hg19.fa >obs/path/sample2.sam
//
// the final workflows job
//
// job-GATK:
//   tool: GATK:4.0.1
//   image: 1.0.0.21:/root/GATK:4.0.1
//   resources:
//     memory: 2G
//     cpu: 2C
//   command:
//     - sh ${obs-path}/${jobid}/bwa_mem.sh obs/path/sample1.fastq.gz obs/path/hg19.fa >obs/path/sample1.sam
//     - sh ${obs-path}/${jobid}/bwa_mem.sh obs/path/sample1.fastq.gz obs/path/hg19.fa >obs/path/sample2.sam
type Tool struct {
	// The Name of tool.
	// Required.
	Name string `json:"Name" yaml:"Name"`
	// The version of tool.
	// Required.
	Version string `json:"version" yaml:"version"`
	// Docker image Name.
	// Required.
	Image string `json:"image" yaml:"image"`
	// Command is the task that will be run for gene sequencing.
	// If set, it will append to workflows commands.
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
	// the type of tool.
	Type string `json:"type" yaml:"type"`
	// Description describes what the tool is used for.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Input defines input parameter that used for gene sequencing.
//
// input example
//
// sample:
//   default: /sample/data
//   description: the path that stores the sample data.
//   type: string
type Input struct {
	// Default is the default value that will be used for an input parameter
	// if a value was not provided.
	Default interface{} `json:"default,omitempty" yaml:"default,omitempty"`

	// Value is the literal value to use for the parameter.
	Value interface{} `json:"value,omitempty" yaml:"value,omitempty"`

	// Description is the information about what the parameter is used for.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Type defines the type for the input parameter.
	// One of string, number, bool, array.
	// The default type is string if not specified.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// Compute Resources required by this container.
type Resources struct {
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
	Cpu    string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
}

type Var []interface{}

// CommandsIter defines command for workflows job. If both Vars and Vars_iter are specified,
// the generate command will be merged.
type CommandsIter struct {
	// Command is the base command that contains variables.
	Command string `json:"command" yaml:"command"`

	// Vars list all the parameters for the command.
	//
	// commandsIter example
	//
	//    commands_iter:
	//      command: sh /tmp/scripts/step1.splitfq.sh ${1} ${2} /tmp/data ${3}
	//      vars:
	//        - ["sample1", 0, 25] # Each member of the array will be the ${1}, ${2}, ${3}
	//        - ["sample2", 0, 25]
	//        - ["sample1", 1, 25]
	//        - ["sample2", 1, 25]
	//
	// then the final command will be:
	//
	// sh /tmp/scripts/step1.splitfq.sh sample1 0 ${sample-path} 25
	// sh /tmp/scripts/step1.splitfq.sh sample2 0 ${sample-path} 25
	// sh /tmp/scripts/step1.splitfq.sh sample1 1 ${sample-path} 25
	// sh /tmp/scripts/step1.splitfq.sh sample2 1 ${sample-path} 25
	Vars []interface{} `json:"vars,omitempty" yaml:"vars,omitempty"`

	// VarsIter list all the possible parameters for every position in the command line.
	// And we will use algorithm Of Full Permutation to generate all the permutation and
	// combinations for these parameter that will be used to replace the ${number} variable.
	//
	// commandsIter example
	//
	//    commands_iter:
	//      command: sh /tmp/scripts/step1.splitfq.sh ${1} ${2} /tmp/data ${3}
	//      vars_iter:
	//        - ["sample1", "sample2"]
	//        - [0, 1]
	//        - [25]
	//
	// then the final command will be:
	//
	// sh /tmp/scripts/step1.splitfq.sh sample1 0 /tmp/data 25
	// sh /tmp/scripts/step1.splitfq.sh sample2 0 /tmp/data 25
	// sh /tmp/scripts/step1.splitfq.sh sample1 1 /tmp/data 25
	// sh /tmp/scripts/step1.splitfq.sh sample2 1 /tmp/data 25
	VarsIter []interface{} `json:"vars_iter,omitempty" yaml:"vars_iter,omitempty"`

	Depends map[string]bool `json:"depends,omitempty" yaml:"depends,omitempty"`
}

const (
	WholeDependType   = "whole"
	IterateDependType = "iterate"
)

type Depend struct {
	// Target is the Name of job this depends on.
	Target string `json:"target" yaml:"target"`

	// Type is the depends type.
	// One of whole, iterate
	// Default to `whole`.
	//
	// Examples:
	//
	// "whole" - A-->B": jobs of B depends on all jobs of A execution done.
	// "iterate" - A[1,2,3]-->B[1,2,3]: jobs of B depends on jobs of A one by one.
	// That is to say: A[1]->B[1], A[2]->B[2], A[3]->B[3]
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// job information.
type JobInfo struct {
	// Description describes what this job is to do.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// tool Name
	Tool string `json:"tool" yaml:"tool"`
	// docker image Name.
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// Compute Resources required by this job.
	Resources Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	// command to run for gene sequencing.
	Commands []string `json:"commands,omitempty" yaml:"commands,omitempty"`
	// CommandsIter defines batch command for workflows job.
	CommandsIter CommandsIter `json:"commands_iter,omitempty" yaml:"commands_iter,omitempty"`
	// Depends is the Name of task this depends on.
	Depends []Depend `json:"depends,omitempty" yaml:"depends,omitempty"`
}

// PathsIter similar to CommandsIter.
type PathsIter struct {
	Path     string        `json:"path" yaml:"path"`
	Vars     []interface{} `json:"vars,omitempty" yaml:"vars,omitempty"`
	VarsIter []interface{} `json:"vars_iter,omitempty" yaml:"vars_iter,omitempty"`
}

type VolumeSource struct {
	PVC string `json:"pvc" yaml:"pvc"`
}

type Volume struct {
	MountPath string       `json:"mount_path" yaml:"mount_path"`
	MountFrom VolumeSource `json:"mount_from" yaml:"mount_from"`
}

type OutputDesc struct {
	Paths     []string  `json:"paths" yaml:"paths"`
	PathsIter PathsIter `json:"paths_iter,omitempty" yaml:"paths_iter,omitempty"`
}

//
type Workflow struct {
	// Workflow version
	Version string                `json:"version" yaml:"version"`
	Inputs  map[string]Input      `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Jobs    map[string]JobInfo    `json:"workflow" yaml:"workflow"`
	Volumes map[string]Volume     `json:"volumes" yaml:"volumes"`
	Outputs map[string]OutputDesc `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Tools   map[string]Tool       `json:"tools" yaml:"tools"`
}

// ErrorList holds a set of Errors.
type ErrorList []error
