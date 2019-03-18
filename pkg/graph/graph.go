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

package graph

import (
	"container/list"
	"fmt"
	batch "k8s.io/api/batch/v1"
	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	"strings"
	"sync"
)

// JobInfo stores job information for running
type JobInfo struct {
	Finished   bool
	Job        *batch.Job
	TaskType   genev1alpha1.TaskType
	DynamicJob *genev1alpha1.Task
}

func NewJobInfo(job *batch.Job, finished bool, taskType genev1alpha1.TaskType, dynamicJob *genev1alpha1.Task) *JobInfo {
	return &JobInfo{
		Job:        job,
		Finished:   finished,
		TaskType:   taskType,
		DynamicJob: dynamicJob,
	}
}

type Vertex struct {
	Data     *JobInfo
	Children []*Vertex
	dynamic  bool
}

type Graph struct {
	sync.RWMutex
	NumOfSuccess int
	VertexCount  int
	Size         int
	VertexArray  []*Vertex
	AdjMatrix    []int
}

func NewGraph(size int) *Graph {
	return &Graph{
		Size:        size,
		VertexArray: make([]*Vertex, size),
		AdjMatrix:   make([]int, size*size),
	}
}

func NewVertex(data *JobInfo, flag bool, children ...*Vertex) *Vertex {
	vertex := &Vertex{
		Data:     data,
		Children: make([]*Vertex, 0),
		dynamic:  flag,
	}
	vertex.Children = append(vertex.Children, children...)

	return vertex
}

func (n *Vertex) AddChild(vertex *Vertex) {
	if vertex != nil {
		n.Children = append(n.Children, vertex)
	}
}

func (g *Graph) AddVertex(Vertex *Vertex) {
	if g.VertexCount >= g.Size {
		return
	}
	g.VertexArray[g.VertexCount] = Vertex
	g.VertexCount++
}

func (g *Graph) SetAdjMatrix() bool {
	if len(g.VertexArray) != g.Size {
		return false
	}

	for row := 0; row < len(g.VertexArray); row++ {
		for col := 0; col < len(g.VertexArray); col++ {
			for _, child := range g.VertexArray[row].Children {
				if child == g.VertexArray[col] {
					g.AdjMatrix[row*g.Size+col] = 1
				}
			}
		}
	}

	return true
}

func (g *Graph) DFS(stack *list.List, onStack map[int]bool, visited map[int]bool) error {
	indexEle := stack.Back()
	vertexPos := indexEle.Value.(int)
	onStack[vertexPos] = true
	visited[vertexPos] = true
	for col := 0; col < g.VertexCount; col++ {
		if g.AdjMatrix[vertexPos*g.Size+col] != 0 {
			visit, ok := visited[col]
			if !ok || !visit {
				ele := stack.PushBack(col)
				err := g.DFS(stack, onStack, visited)
				if err != nil {
					return err
				}
				stack.Remove(ele)
			} else if VertexOnStack, ok := onStack[col]; ok && VertexOnStack {
				return fmt.Errorf("have Cycle")
			}
		}
	}
	onStack[vertexPos] = false
	return nil
}

func (g *Graph) CheckCycle(checkEnd int) (bool, []int) {
	visited := make(map[int]bool)
	onStack := make(map[int]bool)
	var processIds []int
	for index := 0; index < checkEnd; index++ {
		stack := list.New()
		ele := stack.PushBack(index)
		err := g.DFS(stack, onStack, visited)
		if err != nil {
			for e := stack.Front(); e != nil; e = e.Next() {
				vertexPos := e.Value.(int)
				processIds = append(processIds, vertexPos)
			}
			return false, processIds
		}
		stack.Remove(ele)
	}
	return true, nil
}

func (g *Graph) IsDAG() bool {
	isDAG, _ := g.CheckCycle(g.Size)
	return isDAG
}

func (g *Graph) DirectedTraverse(start int) ([]int, error) {
	if start >= g.Size {
		return nil, fmt.Errorf("start Vertex exceed the graph size")
	}
	visited := make(map[int]bool)
	queue := list.New()
	g.dirTraverse(start, visited, queue)

	vertices := []int{}
	for e := queue.Front(); e != nil; e = e.Next() {
		vertexPos := e.Value.(int)
		vertices = append(vertices, vertexPos)
	}

	return vertices, nil
}

func (g *Graph) FindDependents(vertex int) []int {
	dependents := []int{}
	for row := 0; row < g.VertexCount; row++ {
		if g.AdjMatrix[row*g.Size+vertex] != 0 {
			dependents = append(dependents, row)
		}
	}

	return dependents
}

func (g *Graph) FindChildren(vertex int) []int {
	children := []int{}
	for col := 0; col < g.VertexCount; col++ {
		if g.AdjMatrix[vertex*g.Size+col] != 0 {
			children = append(children, col)
		}
	}

	return children
}

func (g *Graph) dirTraverse(start int, visited map[int]bool, queue *list.List) {
	if _, ok := visited[start]; ok {
		return
	}
	visited[start] = true

	dependents := g.FindDependents(start)
	for _, dependent := range dependents {
		if _, ok := visited[dependent]; !ok {
			g.dirTraverse(dependent, visited, queue)
		}
	}

	queue.PushBack(start)

	children := g.FindChildren(start)
	for _, child := range children {
		if _, ok := visited[child]; !ok {
			g.dirTraverse(child, visited, queue)
		}
	}
}

func (g *Graph) FindChildrenByName(jobName string) []*Vertex {
	for _, vertex := range g.VertexArray {
		if jobInfo := vertex.Data; jobInfo.Job.Name == jobName {
			return vertex.Children
		}
	}

	return nil
}

func (g *Graph) FindDependentsByName(jobName string) []int {
	for i, vertex := range g.VertexArray {
		if jobInfo := vertex.Data; jobInfo.Job.Name == jobName {
			return g.FindDependents(i)
		}
	}

	return nil
}

func (g *Graph) GetRootVertex() []*Vertex {
	rootVertex := make([]*Vertex, 0)
	for i, vertex := range g.VertexArray {
		if len(g.FindDependents(i)) == 0 {
			rootVertex = append(rootVertex, vertex)
		}
	}
	return rootVertex
}

func (g *Graph) FindVertexByName(jobName string) *Vertex {
	for _, vertex := range g.VertexArray {
		if jobInfo := vertex.Data; jobInfo.Job.Name == jobName {
			return vertex
		}
		//jobNamePrefix := execution.Name + Separator + task.Name + Separator
		if vertex.dynamic {
			if jobInfo := vertex.Data; strings.HasPrefix(jobName, jobInfo.Job.Name) {
				return vertex
			}
		}
	}

	return nil
}

func (g *Graph) FindVertex(vertex int) *Vertex {
	if vertex > g.VertexCount {
		return nil
	}

	return g.VertexArray[vertex]
}

func (g *Graph) PlusNumOfSuccess() {
	g.Lock()
	defer g.Unlock()
	g.NumOfSuccess++
}
