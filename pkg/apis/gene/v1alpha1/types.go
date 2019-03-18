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

package v1alpha1

import (
	"encoding/json"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VertexPhase is a label for the condition of a node at the current time.
type VertexPhase string

// Vertex status in the execution.
const (
	VertexRunning   VertexPhase = "Running"
	VertexSucceeded VertexPhase = "Succeeded"
	VertexFailed    VertexPhase = "Failed"
	VertexError     VertexPhase = "Error"
)

// TaskType is the type of a job
type TaskType string

// Possible Task types
const (
	JobTaskType   TaskType = "Job"
	SparkTaskType TaskType = "Spark"
)

// VertexType is the type of a vertex
type VertexType string

// DAG vertex types
const (
	DAGVertexType        VertexType = "DAG"
	JobVertexType        VertexType = "Job"
	JobGroupVertexType   VertexType = "JobGroup"
	SparkVertexType      VertexType = "Spark"
	SparkGroupVertexType VertexType = "SparkGroup"
)

// DependType is the type of depend
type DependType string

// DependType
const (
	DependTypeWhole   DependType = "whole"
	DependTypeIterate DependType = "iterate"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Execution is the definition of kubegene workflow.
type Execution struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Spec ExecutionSpec `json:"spec,omitempty"`
	// +optional
	Status ExecutionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExecutionList is a collection of executions.
type ExecutionList struct {
	metav1.TypeMeta `json:",inline" `
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of executions.
	Items []Execution `json:"items"`
}

type ExecutionSpec struct {
	// Tasks is a list of Tasks used in a workflow
	Tasks []Task `json:"tasks"`

	// NodeSelector is a selector which will result in all pods of the workflow
	// to be scheduled on the selected node(s). This is able to be overridden by
	// a nodeSelector specified in the job.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity sets the scheduling constraints for all pods in the workflow.
	// Can be overridden by an affinity specified in the Job
	// +optional
	Affinity *apiv1.Affinity `json:"affinity,omitempty"`

	// Tolerations to apply to workflow pods.
	// +optional
	Tolerations []apiv1.Toleration `json:"tolerations,omitempty"`

	// Parallelism limits the max total parallel jobs that can execute at the same time in a workflow
	// +optional
	Parallelism *int64 `json:"parallelism,omitempty"`
}

// Task is a unit of execution in an Execution
type Task struct {
	// Name is the name of the task
	Name string `json:"name"`

	// Type is the type of the task
	Type TaskType `json:"type"`

	// NodeSelector is a selector to schedule this step of the workflow to be
	// run on the selected node(s). Overrides the selector set at the execution level.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity sets the pod's scheduling constraints
	// Overrides the affinity set at the execution level (if any)
	// +optional
	Affinity *apiv1.Affinity `json:"affinity,omitempty"`

	// Tolerations to apply to task pods.
	// Overrides the tolerations set at the execution level (if any)
	// +optional
	Tolerations []apiv1.Toleration `json:"tolerations,omitempty"`

	// CommandSet is a list of commands run by this task.
	CommandSet []string `json:"commandSet,omitempty"`

	// CommandsIter defines batch command for workflows job.
	CommandsIter CommandsIter `json:"commands_iter,omitempty"`
	// Docker image name.
	// More info: https://kubernetes.io/docs/concepts/containers/images
	Image string `json:"image,omitempty"`

	// Volumes is a list of volumes that can be mounted by containers in the task.
	// +optional
	Volumes map[string]Volume `json:"volumes,omitempty"`

	// +optional
	Resources ResourceRequirements `json:"resources,omitempty"`

	// Specifies the duration in seconds relative to the startTime that the job may be active
	// before the system tries to terminate it; value must be positive integer
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// Specifies the number of retries before marking this job failed.
	// If not set use the k8s job default.
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// Parallelism limits the max total parallel jobs that can execute at the same time within the
	// boundaries of this job invocation.
	// Overrides the parallelism set at the execution level (if any)
	Parallelism *int64 `json:"parallelism,omitempty"`

	// Specifies the dependency by this task
	// +optional
	Dependents []Dependent `json:"dependents"`
}

// +k8s:openapi-gen=false
type ExecutionStatus struct {
	// Phase a simple, high-level summary of where the workflow is in its lifecycle.
	Phase VertexPhase `json:"phase,omitempty"`

	// Time at which this workflow started
	StartedAt metav1.Time `json:"startedAt,omitempty"`

	// Time at which this workflow completed
	FinishedAt metav1.Time `json:"finishedAt,omitempty"`

	// A human readable message indicating details about why the workflow is in this condition.
	Message string `json:"message,omitempty"`

	// Vertices is a mapping between a vertex ID and the vertex's status.
	Vertices map[string]VertexStatus `json:"vertices,omitempty"`
}

// CommandsIter defines command for workflows job. If both Vars and Vars_iter are specified,
// the generate command will be merged. This is used for the dynamically generating task
// based on the get_result
type CommandsIter struct {
	// Command is the base command that contains variables.
	Command string `json:"command"`

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
	Vars []interface{} `json:"vars,omitempty"`

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
	VarsIter []interface{} `json:"vars_iter,omitempty"`

	Depends map[string]bool `json:"depends,omitempty"`
}

type VertexStatus struct {
	// ID is a unique identifier of a vertex within the worklow
	// It is implemented as a hash of the vertex name, which makes the ID deterministic
	ID string `json:"id"`

	// Name is unique name in the graph tree used to generate the vertex ID
	Name string `json:"name"`

	// Type indicates type of vertex
	Type VertexType `json:"type"`

	// Phase a simple, high-level summary of where the vertex is in its lifecycle.
	// Can be used as a state machine.
	Phase VertexPhase `json:"phase,omitempty"`

	// A human readable message indicating details about why the vertex is in this condition.
	Message string `json:"message,omitempty"`

	// Time at which this vertex started
	StartedAt metav1.Time `json:"startedAt,omitempty"`

	// Time at which this vertex completed
	FinishedAt metav1.Time `json:"finishedAt,omitempty"`

	// Children is a list of child vertex IDs
	Children []string `json:"children,omitempty"`
}

type Volume struct {
	MountPath string       `json:"mountPath"`
	MountFrom VolumeSource `json:"mountFrom"`
}

type VolumeSource struct {
	Pvc string `json:"pvc"`
}

type ResourceRequirements struct {
	Memory resource.Quantity `json:"memory"`
	Cpu    resource.Quantity `json:"cpu"`
}

type Dependent struct {
	// Target is the name of task this depends on.
	Target string `json:"target"`
	// Type is the depends type.
	// Default to `whole`.
	// Examples:
	//  "whole" - A-->B": jobs of B depends on all jobs of A execution done.
	//  "iterate" - A[1,2,3]-->B[1,2,3]: jobs of B depends on jobs of A one by one.
	//   That is to say: A[1]->B[1], A[2]->B[2], A[3]->B[3]
	Type DependType `json:"type,omitempty"`
}

// DeepCopyInto is an custom deepcopy function to deal with our use of the interface{} type
func (i *CommandsIter) DeepCopyInto(out *CommandsIter) {

	inBytes, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(inBytes, out)
	if err != nil {
		panic(err)
	}
}
