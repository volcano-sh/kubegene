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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubegene.io/kubegene/cmd/genectl/util"
)

func GetExecutionNamespace(inputs map[string]Input) string {
	namespace := GetStringValue("namespace", inputs)
	if len(namespace) == 0 {
		namespace = metav1.NamespaceDefault
	}
	return namespace
}

func GetExecutionName(inputs map[string]Input) string {
	name := GetStringValue("executionName", inputs)
	if len(name) == 0 {
		name = util.GenerateExecName("execution")
	}
	return name
}
