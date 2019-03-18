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

package controller

import (
	"strconv"
	"strings"
	"sync"

	"fmt"
	batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	"kubegene.io/kubegene/pkg/graph"
)

// Separator used to construct job name.
const Separator = "."

// GraphBuilder: based on the executions supplied by the informers, GraphBuilder updates
// jobs map, a struct that caches the execution uid to jobs
type GraphBuilder struct {
	sync.RWMutex
	graphs map[string]*graph.Graph
}

func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		graphs: make(map[string]*graph.Graph),
	}
}

func (gb *GraphBuilder) AddGraph(execution *genev1alpha1.Execution) {
	g := newGraph(execution)
	gb.Lock()
	defer gb.Unlock()
	gb.graphs[execution.Namespace+"/"+execution.Name] = g
}

func (gb *GraphBuilder) DeleteGraph(key string) {
	gb.Lock()
	defer gb.Unlock()
	delete(gb.graphs, key)
}

func (gb *GraphBuilder) GetGraph(key string) *graph.Graph {
	gb.RLock()
	defer gb.RUnlock()

	graph, ok := gb.graphs[key]
	if !ok {
		return nil
	}
	return graph
}

func newGraph(execution *genev1alpha1.Execution) *graph.Graph {
	jobInfos := []*graph.JobInfo{}
	vertices := []*graph.Vertex{}
	for _, task := range execution.Spec.Tasks {
		if len(task.Tolerations) == 0 {
			task.Tolerations = execution.Spec.Tolerations
		}
		if len(task.NodeSelector) == 0 {
			task.NodeSelector = execution.Spec.NodeSelector
		}
		if task.Affinity == nil {
			task.Affinity = execution.Spec.Affinity
		}

		jobNamePrefix := execution.Name + Separator + task.Name + Separator

		if len(task.CommandsIter.Depends) > 0 {
			localtask := task
			fmt.Println("task.CommandSet", task.CommandSet)
			jobName := jobNamePrefix //+ strconv.Itoa(0)
			// make up k8s job resource
			job := newJob(jobName, "", execution, &task)
			jobInfo := graph.NewJobInfo(job, false, task.Type, &localtask)
			jobInfos = append(jobInfos, jobInfo)
			vertices = append(vertices, graph.NewVertex(jobInfo, true))

		} else {

			for index, command := range task.CommandSet {
				jobName := jobNamePrefix + strconv.Itoa(index)
				// make up k8s job resource
				job := newJob(jobName, command, execution, &task)
				jobInfo := graph.NewJobInfo(job, false, task.Type, nil)
				jobInfos = append(jobInfos, jobInfo)
				vertices = append(vertices, graph.NewVertex(jobInfo, false))
			}
		}
	}

	for vertexIndex, jobInfo := range jobInfos {
		items := strings.Split(jobInfo.Job.Name, Separator)
		taskName := items[len(items)-2]
		jobSuffix := items[len(items)-1]

		for _, task := range execution.Spec.Tasks {
			if task.Name != taskName {
				continue
			}
			for _, dependent := range task.Dependents {
				switch dependent.Type {
				case genev1alpha1.DependTypeWhole, "":
					for index, jobInfo := range jobInfos {
						items := strings.Split(jobInfo.Job.Name, Separator)
						if dependent.Target == items[len(items)-2] {
							vertices[index].AddChild(vertices[vertexIndex])
						}
					}
				case genev1alpha1.DependTypeIterate:
					for index, jobInfo := range jobInfos {
						items := strings.Split(jobInfo.Job.Name, Separator)
						if dependent.Target == items[len(items)-2] && jobSuffix == items[len(items)-1] {
							vertices[index].AddChild(vertices[vertexIndex])
						}
					}
				}
			}
			break
		}
	}

	// create graph
	g := graph.NewGraph(len(vertices))

	// add vertex
	for _, vertex := range vertices {
		g.AddVertex(vertex)
	}

	// initialize the adjacency matrix
	g.SetAdjMatrix()

	return g
}

func newJob(name, command string, exec *genev1alpha1.Execution, task *genev1alpha1.Task) *batch.Job {
	volumes := []v1.Volume{}
	volumeMounts := []v1.VolumeMount{}
	for name, volume := range task.Volumes {
		volumes = append(volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: volume.MountFrom.Pvc,
				},
			},
		})

		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: volume.MountPath,
		})
	}

	controllerRef := metav1.NewControllerRef(exec, execKind)
	containerName := strings.Replace(name, ".", "-", -1)

	return &batch.Job{
		TypeMeta: metav1.TypeMeta{Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       exec.Namespace,
			Labels:          map[string]string{"controller-uid": string(exec.UID)},
			OwnerReferences: []metav1.OwnerReference{*controllerRef},
		},
		Spec: batch.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyOnFailure,
					Containers: []v1.Container{
						{
							Name:            containerName,
							Image:           task.Image,
							Command:         []string{"sh", "-c", command},
							VolumeMounts:    volumeMounts,
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					NodeSelector: task.NodeSelector,
					Affinity:     task.Affinity,
					Tolerations:  task.Tolerations,
					Volumes:      volumes,
				},
			},
		},
	}
}
