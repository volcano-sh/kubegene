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
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"kubegene.io/kubegene/pkg/common"
)

func ReplaceArray(base []string, kv map[string]string) []string {
	array := make([]string, 0, len(base))
	for _, str := range base {
		newStr := common.ReplaceVariant(str, kv)
		array = append(array, newStr)
	}
	return array
}

// IsVariant checks whether a string is variant.
// valid variant format ${.var}, ${var}.
func IsVariant(str string) bool {
	if matched, _ := regexp.MatchString("^\\${.*}$", str); matched {
		return true
	}
	return false
}

// GetVariantName extract variant Name.
// such as, ${var} --> var
func GetVariantName(str string) string {
	re, _ := regexp.Compile("^\\${(.*)}$")
	submatch := re.FindStringSubmatch(str)
	return submatch[1]
}

const IsRangeFuncRegexFmt = "^range\\(([^,]+)\\s*,\\s*([^,]+)(,\\s*([^,]+))?\\)$"

// IsRangeFunc checks a string whether is a range function.
// Range function in the workflows must follow the format:
//
// 		range(start, end, step)
//
// start, end, and step must be integer type or variant that reference
// a integer data. If step is not specified, default to 1.
//
// range function example
//
// ---- range(2, ${npart}, 2)
// ---- range(1, 10, 2)
// ---- range(1, 10)
func IsRangeFunc(str string) bool {
	if matched, _ := regexp.MatchString(IsRangeFuncRegexFmt, str); matched {
		return true
	}
	return false
}

const GetRangeFuncRegexFmt = "range\\(([^,]+)\\s*,\\s*([^,]+)(,\\s*([^,]+))?\\)"

// GetRangeFuncParam extract parameter from range function.
// for example:
//
// 		Str ---> range(2,${npart},2)
// 		result ---> 2, ${npart}, 2
//
// 		Str ---> range(2, 10, 2)
// 		result ---> 2, 10, 2
func GetRangeFuncParam(str string) (string, string, string) {
	reg, _ := regexp.Compile(GetRangeFuncRegexFmt)
	submatch := reg.FindStringSubmatch(str)
	start := submatch[1]
	end := submatch[2]
	step := submatch[4]
	return start, end, step
}

func ValidateRangeFuncParam(prefix, param string, inputs map[string]Input) error {
	if IsVariant(param) {
		if err := ValidateVariant(prefix, param, []string{NumberType}, inputs); err != nil {
			return err
		}
	} else {
		_, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return fmt.Errorf("%s: the parameters of range should only be variant or number, but the real one is %s", prefix, param)
		}
	}

	return nil
}

// ValidateRangeFunc validate parameter of range function is valid.
func ValidateRangeFunc(prefix, str string, inputs map[string]Input) ErrorList {
	allErr := ErrorList{}
	start, end, step := GetRangeFuncParam(str)
	err := ValidateRangeFuncParam(prefix, start, inputs)
	if err != nil {
		allErr = append(allErr, err)
	}
	err = ValidateRangeFuncParam(prefix, end, inputs)
	if err != nil {
		allErr = append(allErr, err)
	}
	if step == "" {
		return allErr
	}
	err = ValidateRangeFuncParam(prefix, step, inputs)
	if err != nil {
		allErr = append(allErr, err)
	}
	return allErr
}

func InstantiateRangeFunc(prefix, str string, data map[string]string) (common.Var, error) {
	start, end, step := GetRangeFuncParam(str)

	// replace variant for start
	start = common.ReplaceVariant(start, data)
	startNum, err := strconv.ParseFloat(start, 64)
	if err != nil {
		return nil, fmt.Errorf("%s convert start to float err: %v", prefix, err)
	}

	// replace variant for end
	end = common.ReplaceVariant(end, data)
	endNum, err := strconv.ParseFloat(end, 64)
	if err != nil {
		return nil, fmt.Errorf("%s convert end to float err: %v", prefix, err)
	}

	var stepNum float64 = 1
	if len(step) != 0 {
		step = common.ReplaceVariant(step, data)
		stepNum, err = strconv.ParseFloat(step, 64)
		if err != nil {
			return nil, fmt.Errorf("%s convert step to float err: %v", prefix, err)
		}
	}

	if startNum >= endNum {
		return nil, fmt.Errorf("%s:In range function, start should be smaller than end", prefix)
	}

	if stepNum < 0 {
		return nil, fmt.Errorf("%s:In range function, step should be larger than 0", prefix)
	}

	numbers := make([]interface{}, 0)
	for i := startNum; i < endNum; i += stepNum {
		numbers = append(numbers, i)
	}

	return numbers, nil
}

func InstantiateVars(prefix string, vars []interface{}, data map[string]string) ([]common.Var, error) {
	result := make([]common.Var, 0, len(vars))
	for i, v := range vars {
		if strValue, ok := v.(string); ok {
			prefix := fmt.Sprintf("%s[%d]", prefix, i)
			if IsRangeFunc(strValue) {
				rangeValue, err := InstantiateRangeFunc(prefix, strValue, data)
				if err != nil {
					return nil, err
				}
				result = append(result, rangeValue)
				continue
			} else {
				variant := GetVariantName(strValue)
				var array common.Var
				err := json.Unmarshal([]byte(data[variant]), &array)
				if err != nil {
					return nil, fmt.Errorf("unmarshal %s error", prefix)
				}
				result = append(result, array)
			}
		} else if array, ok := v.([]interface{}); ok {
			for j, s := range array {
				if varStr, ok := s.(string); ok {
					array[j] = common.ReplaceVariant(varStr, data)
				}
			}
			result = append(result, array)
		}
	}

	return result, nil
}

func InstantiateVarsIter(prefix string, vars []interface{}, data map[string]string) ([]common.Var, bool, error) {
	result := make([]common.Var, 0, len(vars))
	dynamicjob := false
	for i, v := range vars {
		if strValue, ok := v.(string); ok {
			prefix := fmt.Sprintf("%s[%d]", prefix, i)
			if IsRangeFunc(strValue) {
				rangeValue, err := InstantiateRangeFunc(prefix, strValue, data)
				if err != nil {
					return nil, false, err
				}
				result = append(result, rangeValue)
				continue
			} else if IsGetResultFunc(strValue) {
				getresult := InstantiateGetResultFunc(prefix, strValue, data)

				dynamicjob = true
				result = append(result, getresult)
				continue
			} else {
				variant := GetVariantName(strValue)
				var array common.Var
				err := json.Unmarshal([]byte(data[variant]), &array)
				if err != nil {
					return nil, false, fmt.Errorf("unmarshal %s error", prefix)
				}
				result = append(result, array)
			}
		} else if array, ok := v.([]interface{}); ok {
			for j, s := range array {
				if varStr, ok := s.(string); ok {
					array[j] = common.ReplaceVariant(varStr, data)
				}
			}
			result = append(result, array)
		}
	}

	return result, dynamicjob, nil
}

func sliceContain(strArr []string, str string) int {
	for i, v := range strArr {
		if v == str {
			return i
		}
	}
	return -1
}

func IsPathIterEmpty(pathIter PathsIter) bool {
	if len(pathIter.Vars) == 0 && len(pathIter.VarsIter) == 0 {
		return true
	}
	return false
}

func IsValidType(typeStr string, typeList []string) bool {
	for _, t := range typeList {
		if t == typeStr {
			return true
		}
	}
	return false
}

func IsValidInputValue(val interface{}, typeStr string) bool {
	switch val.(type) {
	case float64, int:
		if typeStr == NumberType {
			return true
		}
	case bool:
		if typeStr == BoolType {
			return true
		}
	case string:
		if typeStr == StringType {
			return true
		}
	case []interface{}:
		if typeStr == ArrayType {
			return true
		}
	}
	return false
}

func GetInputType(val interface{}) string {
	switch val.(type) {
	case float64, int:
		return NumberType
	case bool:
		return BoolType
	case string:
		return StringType
	case []interface{}:
		return ArrayType
	}
	return ""
}

func ToArrayIndex(str string) (int, bool) {
	index, err := strconv.Atoi(str)
	if err != nil {
		return -1, false
	}
	if index < 0 {
		return -1, false
	}
	return index, true
}

func ValidateTemplate(template, prefix, typeStr string, inputs map[string]Input) (int, ErrorList) {
	errors := ErrorList{}
	regex, _ := regexp.Compile("\\${([^{}]*)}")
	subMatches := regex.FindAllStringSubmatch(template, -1)
	maxIndex := 0
	for _, subMatch := range subMatches {
		variant := subMatch[1]
		if variant == "item" {
			continue
		}
		if index, ok := ToArrayIndex(variant); ok {
			if index > maxIndex {
				maxIndex = index
			}
			continue
		}
		if _, ok := inputs[variant]; !ok {
			err := fmt.Errorf("%s: variant [%s] undefine", prefix, variant)
			errors = append(errors, err)
		}
	}
	return maxIndex, errors
}

func ValidateVarType(prefix string, varValue interface{}, inputs map[string]Input) error {
	switch v := varValue.(type) {
	case string:
		if IsVariant(v) {
			if err := ValidateVariant(prefix, v, []string{NumberType, StringType, BoolType}, inputs); err != nil {
				return err
			}
		}
	case []interface{}:
		return fmt.Errorf("%s: the value type should not be array", prefix)
	}
	return nil
}

func ValidateVarsTypes(prefix string, vars interface{}, inputs map[string]Input) ErrorList {
	allErr := ErrorList{}
	switch v := vars.(type) {
	case string:
		if IsVariant(v) {
			if err := ValidateVariant(prefix, v, []string{ArrayType}, inputs); err != nil {
				return append(allErr, err)
			}
		} else {
			err := fmt.Errorf("%s:the element of vars array should only be array variant or array, but the real one is %v", prefix, v)
			return append(allErr, err)
		}
	case []interface{}:
		for i, varValue := range v {
			prefix = fmt.Sprintf("%s[%d]", prefix, i)
			err := ValidateVarType(prefix, varValue, inputs)
			if err != nil {
				return append(allErr, err)
			}
		}
	default:
		err := fmt.Errorf("%s:the element of vars array should only be array variant or array, but the real one is %v", prefix, v)
		return append(allErr, err)
	}
	return allErr
}

func ValidateVarsIterTypes(prefix string, vars interface{}, inputs map[string]Input, jobName string, workflow *Workflow) ErrorList {
	allErr := ErrorList{}
	switch v := vars.(type) {
	case string:
		if IsVariant(v) {
			if err := ValidateVariant(prefix, v, []string{ArrayType}, inputs); err != nil {
				return append(allErr, err)
			}
		} else if strings.HasPrefix(v, "range") {
			if !IsRangeFunc(v) {
				err := fmt.Errorf("%s: the range function should be defined as range(start, end, step), but the real one is %s", prefix, v)
				return append(allErr, err)
			} else {
				errors := ValidateRangeFunc(prefix, v, inputs)
				if len(errors) != 0 {
					return append(allErr, errors...)
				}
			}
		} else if strings.HasPrefix(v, "get_result") {
			if !IsGetResultFunc(v) {
				err := fmt.Errorf("%s: the get_result function should be defined as get_result(jobName, sep), but the real one is %s", prefix, v)
				return append(allErr, err)
			} else {
				errors := validateGetResultFunc(prefix, v, inputs, jobName, workflow)
				if len(errors) != 0 {
					return append(allErr, errors...)
				}
			}
		} else {
			err := fmt.Errorf("%s:the element of vars array should only be array variant or array, but the real one is %v", prefix, v)
			return append(allErr, err)
		}
	case []interface{}:
		for i, varValue := range v {
			prefix = fmt.Sprintf("%s[%d]", prefix, i)
			err := ValidateVarType(prefix, varValue, inputs)
			if err != nil {
				return append(allErr, err)
			}
		}
	default:
		err := fmt.Errorf("%s:the element of vars array should only be array variant or array, but the real one is %v", prefix, v)
		return append(allErr, err)
	}
	return allErr
}

func ValidateVarsArray(prefix string, varsArray []interface{}, inputs map[string]Input) ErrorList {
	allErr := ErrorList{}
	for i, vars := range varsArray {
		prefix := fmt.Sprintf("%s[%d]", prefix, i)
		allErr = append(allErr, ValidateVarsTypes(prefix, vars, inputs)...)
	}
	return allErr
}

func ValidateVarsIterArray(prefix string, varsArray []interface{}, inputs map[string]Input, jobName string, workflow *Workflow) ErrorList {
	allErr := ErrorList{}
	for i, vars := range varsArray {
		prefix := fmt.Sprintf("%s[%d]", prefix, i)
		allErr = append(allErr, ValidateVarsIterTypes(prefix, vars, inputs, jobName, workflow)...)
	}
	return allErr
}

func ValidateVariant(prefix, varStr string, types []string, inputs map[string]Input) error {
	varName := GetVariantName(varStr)
	if input, ok := inputs[varName]; !ok {
		return fmt.Errorf("%s: the variant [%s] is not defined in the inputs", prefix, varName)
	} else {
		if !IsValidType(input.Type, types) {
			return fmt.Errorf("%s: the type of %s can only be in %v, but the real one is %s", prefix, varName, types, input.Type)
		}
	}
	return nil
}

func ValidateInstantiatedVars(prefix string, varsArray []common.Var) (int, error) {
	var length int
	if len(varsArray) != 0 {
		length = len(varsArray[0])
	}
	for i, vars := range varsArray {
		if len(vars) != length {
			return 0, fmt.Errorf("%s.vars: the length of 0 line is %d, but the length of %d line is %d", prefix, length, i, len(vars))
		}
	}
	return length, nil
}
