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
	"fmt"
)

func ValidatePaths(outputName string, paths []string, inputs map[string]Input) ErrorList {
	allErr := ErrorList{}
	for i, path := range paths {
		prefix := fmt.Sprintf("outputs.%s.Paths[%d]", outputName, i)
		_, errors := ValidateTemplate(path, prefix, "path", inputs)
		allErr = append(allErr, errors...)
	}
	return allErr
}

func ValidatePathsIter(outputName string, pathIter PathsIter, inputs map[string]Input) ErrorList {
	allErr := ErrorList{}
	if len(pathIter.Path) == 0 && IsPathIterEmpty(pathIter) {
		return allErr
	}
	if len(pathIter.Path) == 0 && !IsPathIterEmpty(pathIter) {
		err := fmt.Errorf("outputs.%s.vars or varsIter is not empty but path is empty", outputName)
		return append(allErr, err)
	}
	if len(pathIter.Path) != 0 && IsPathIterEmpty(pathIter) {
		err := fmt.Errorf("outputs.%s.pathIter: vars or varsIter is empty but command is not empty", outputName)
		return append(allErr, err)
	}

	prefix := fmt.Sprintf("outputs.%s.paths_iter.path", outputName)
	maxIndex, errors := ValidateTemplate(pathIter.Path, prefix, "path", inputs)
	allErr = append(allErr, errors...)

	if maxIndex > len(pathIter.VarsIter) {
		err := fmt.Errorf("outputs.%s.pathsIter.path: ${%d} is larger than pathsIter's rows", outputName, maxIndex)
		allErr = append(allErr, err)
	}

	prefix = fmt.Sprintf("outputs.%s.pathsIter.vars", outputName)
	allErr = append(allErr, ValidateVarsArray(prefix, pathIter.Vars, inputs)...)

	prefix = fmt.Sprintf("outputs.%s.pathsIter.varsIter", outputName)
	allErr = append(allErr, ValidateVarsArray(prefix, pathIter.VarsIter, inputs)...)

	return allErr
}
