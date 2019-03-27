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
	"reflect"
	"testing"
)

func TestSetAdjMatrix(t *testing.T) {
	graph, vertices := newTestGraph(4)

	vertices[0].AddChild(vertices[2])
	vertices[1].AddChild(vertices[3])

	graph.SetAdjMatrix()
	expectedAdjMatrix := []int{0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}

	if !reflect.DeepEqual(graph.AdjMatrix, expectedAdjMatrix) {
		t.Errorf("expected matrix %#v, got %#v", expectedAdjMatrix, graph.AdjMatrix)
	}
}

func TestIsDAG(t *testing.T) {
	graph, vertices := newTestGraph(5)

	vertices[0].AddChild(vertices[2])
	vertices[0].AddChild(vertices[3])
	vertices[1].AddChild(vertices[2])
	vertices[1].AddChild(vertices[3])
	vertices[2].AddChild(vertices[4])
	vertices[3].AddChild(vertices[2])
	graph.SetAdjMatrix()

	isDAG := graph.IsDAG()
	if !isDAG {
		t.Errorf("expected graph bening DAG")
	}

	// set cycle
	vertices[4].AddChild(vertices[3])
	graph.SetAdjMatrix()
	isDAG = graph.IsDAG()
	if isDAG {
		t.Errorf("expected graph not a DAG")
	}
}

func TestDirectedTraverse(t *testing.T) {
	graph, vertices := newTestGraph(5)

	vertices[0].AddChild(vertices[2])
	vertices[1].AddChild(vertices[3])
	vertices[2].AddChild(vertices[4])
	graph.SetAdjMatrix()

	expectedSeqs := []int{0, 2, 4}
	seqs, err := graph.DirectedTraverse(0)
	if err != nil {
		t.Errorf("DirectedTraverse failed %v", err)
	}
	if !reflect.DeepEqual(expectedSeqs, seqs) {
		t.Errorf("expected sequences %#v, got %#v", expectedSeqs, seqs)
	}

	vertices[0].AddChild(vertices[2])
	vertices[0].AddChild(vertices[3])
	vertices[1].AddChild(vertices[2])
	vertices[1].AddChild(vertices[3])
	vertices[2].AddChild(vertices[4])
	vertices[3].AddChild(vertices[2])

	graph.SetAdjMatrix()

	expectedSeqs = []int{0, 1, 3, 2, 4}
	seqs, err = graph.DirectedTraverse(0)
	if err != nil {
		t.Errorf("DirectedTraverse failed %v", err)
	}
	if !reflect.DeepEqual(expectedSeqs, seqs) {
		t.Errorf("expected sequences %#v, got %#v", expectedSeqs, seqs)
	}
}

func newTestGraph(size int) (*Graph, []*Vertex) {
	graph := NewGraph(size)
	vertices := make([]*Vertex, size)
	jobInfo := &JobInfo{}

	for i := 0; i < size; i++ {
		vertex := NewVertex(jobInfo, false)
		vertices[i] = vertex

		graph.AddVertex(vertex)
	}

	return graph, vertices
}
