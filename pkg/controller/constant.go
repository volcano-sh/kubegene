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

const (
	// Number of retry when update execution status.
	statusUpdateRetries = 3

	// Number of retry when update execution spec.
	specUpdateRetries = 3

	executionSuccessMessage = "execution has run successfully"
	executionRunningMessage = "execution is running"
	missVertexMessage       = "execution is running but can not find vertex in the graph"
	vertexRunningMessage    = "vertex is running"
)
