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

package common

import (
	"encoding/json"
	"strconv"
)

type Var []interface{}

func AddVar(varIter []Var, rowCnt, rowNum int, vars Var, result *[]Var) {
	for _, v := range varIter[rowNum] {
		newVar := make([]interface{}, rowNum, rowCnt)
		copy(newVar, vars)
		newVar = append(newVar, v)
		if rowNum+1 == rowCnt {
			*result = append(*result, newVar)
		} else {
			AddVar(varIter, rowCnt, rowNum+1, newVar, result)
		}
	}
}

// VarIter2Vars convert varIter to var.
//
// example
//
//   varIter ---> [[1, 2], [3, 4], [5]]
//   result  ---> [[1, 3, 5], [1, 4, 5], [2, 3, 5], [2, 4, 5]]
func VarIter2Vars(varIter []Var) []Var {
	var result []Var
	if len(varIter) == 0 {
		return result
	}
	vars := make([]interface{}, 0, len(varIter))
	AddVar(varIter, len(varIter), 0, vars, &result)

	return result
}

func ToString(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case uint:
		return strconv.Itoa(int(v))
	case uint32:
		return strconv.Itoa(int(v))
	case uint16:
		return strconv.Itoa(int(v))
	case int8:
		return strconv.Itoa(int(v))
	case bool:
		return strconv.FormatBool(v)
	default:
		jsonValue, err := json.Marshal(v)
		if err != nil {
			return "unknownType"
		}
		return string(jsonValue)
	}
}

// ReplaceVariant replace all the ${var} variant with the real data.
// for example:
// s = "${foo} kubegene ${bar}"
// data = {"foo": "hello", "bar": "world"}
// result: hello kubegene world
func ReplaceVariant(s string, data map[string]string) string {
	buf := make([]byte, 0, 2*len(s))
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+3 < len(s) && s[j+1] == '{' {
			buf = append(buf, s[i:j]...)
			var k int
			for k = j + 2; k < len(s); k++ {
				if s[k] == '}' {
					break
				}
			}
			if v, ok := data[s[j+2:k]]; ok {
				buf = append(buf, v...)
			} else {
				buf = append(buf, s[j:k+1]...)
			}
			j = k + 1
			i = j
		}
	}
	return string(buf) + s[i:]
}

func Iter2Array(base string, vars []Var) []string {
	result := make([]string, 0, len(vars))
	for i, varsRow := range vars {
		varMap := make(map[string]string)
		for j, varCol := range varsRow {
			varMap[strconv.Itoa(j+1)] = ToString(varCol)
			varMap["item"] = ToString(i)
		}
		result = append(result, ReplaceVariant(base, varMap))
	}
	return result
}
