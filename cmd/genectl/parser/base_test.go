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
	"reflect"
	"strconv"
	"testing"

	"kubegene.io/kubegene/pkg/common"
)

func TestIsVariant(t *testing.T) {
	testCases := []struct {
		str    string
		expect bool
	}{
		{
			str:    "foo",
			expect: false,
		},
		{
			str:    "$foo",
			expect: false,
		},
		{
			str:    "foo$",
			expect: false,
		},
		{
			str:    "${foo}",
			expect: true,
		},
		{
			str:    "${.foo}",
			expect: true,
		},
	}

	for i, testCase := range testCases {
		result := IsVariant(testCase.str)
		if result != testCase.expect {
			t.Errorf("%d: unexpected result; got %v, expected %v", i, result, testCase.expect)
		}
	}
}

func TestGetVariantName(t *testing.T) {
	testCases := []struct {
		str    string
		expect string
	}{
		{
			str:    "${.foo}",
			expect: ".foo",
		},
		{
			str:    "${bar}",
			expect: "bar",
		},
	}

	for i, testCase := range testCases {
		result := GetVariantName(testCase.str)
		if result != testCase.expect {
			t.Errorf("%d: unexpected result; got %s, expected %s", i, result, testCase.expect)
		}
	}
}

func TestIsRangeFunc(t *testing.T) {
	testCases := []struct {
		str    string
		expect bool
	}{
		{
			str:    "range(2, ${npart}, 2)",
			expect: true,
		},
		{
			str:    "range(${start}, ${npart}, 2)",
			expect: true,
		},
		{
			str:    "range(${start}, ${npart}, ${end})",
			expect: true,
		},
		{
			str:    "range(1, 10, 2)",
			expect: true,
		},
		{
			str:    "range(1, 10)",
			expect: true,
		},
		{
			str:    "range(1:10)",
			expect: false,
		},
		{
			str:    "range(, 10)",
			expect: false,
		},
		{
			str:    "range(10)",
			expect: false,
		},
		{
			str:    "range{1, a}",
			expect: false,
		},
	}

	for i, testCase := range testCases {
		result := IsRangeFunc(testCase.str)
		if result != testCase.expect {
			t.Errorf("%d: unexpected result; got %v, expected %v", i, result, testCase.expect)
		}
	}
}

func TestGetRangeFuncParam(t *testing.T) {
	testCases := []struct {
		Str         string
		ExpectStart string
		ExpectEnd   string
		ExpectStep  string
	}{
		{
			Str:         "range(2, ${npart}, 2)",
			ExpectStart: "2",
			ExpectEnd:   "${npart}",
			ExpectStep:  "2",
		},
		{
			Str:         "range(${start}, ${npart}, 2)",
			ExpectStart: "${start}",
			ExpectEnd:   "${npart}",
			ExpectStep:  "2",
		},
		{
			Str:         "range(${start}, ${npart}, ${end})",
			ExpectStart: "${start}",
			ExpectEnd:   "${npart}",
			ExpectStep:  "${end}",
		},
		{
			Str:         "range(1, 10, 2)",
			ExpectStart: "1",
			ExpectEnd:   "10",
			ExpectStep:  "2",
		},
		{
			Str:         "range(1, 10)",
			ExpectStart: "1",
			ExpectEnd:   "10",
		},
	}

	for i, testCase := range testCases {
		start, end, step := GetRangeFuncParam(testCase.Str)
		if start != testCase.ExpectStart {
			t.Errorf("%d: unexpected start; got %s, expected %s", i, start, testCase.ExpectStart)
		}
		if end != testCase.ExpectEnd {
			t.Errorf("%d: unexpected end; got %s, expected %s", i, end, testCase.ExpectEnd)
		}
		if step != testCase.ExpectStep {
			t.Errorf("%d: unexpected step; got %s, expected %s", i, step, testCase.ExpectStep)
		}
	}
}

func TestReplaceVariant(t *testing.T) {
	testCases := []struct {
		str    string
		data   map[string]string
		expect string
	}{
		{
			str:    "${foo} kubegene ${bar}",
			data:   map[string]string{"foo": "hello", "bar": "world"},
			expect: "hello kubegene world",
		},
		{
			str:    "${foo} kubegene ${bar}",
			data:   map[string]string{"foo": "hello"},
			expect: "hello kubegene ${bar}",
		},
	}

	for i, testCase := range testCases {
		result := common.ReplaceVariant(testCase.str, testCase.data)
		if result != testCase.expect {
			t.Errorf("%d: unexpected result; got %s, expected %s", i, result, testCase.expect)
		}
	}
}

func TestVarIter2Vars(t *testing.T) {
	varIter := []common.Var{
		{1, 2},
		{3, 4},
		{5},
	}
	expect := []common.Var{
		{1, 3, 5},
		{1, 4, 5},
		{2, 3, 5},
		{2, 4, 5},
	}

	result := common.VarIter2Vars(varIter)
	if !reflect.DeepEqual(result, expect) {
		t.Errorf("unexpected result: got %v, expected %v", result, expect)
	}
}

func TestInstantiateRangeFunc(t *testing.T) {
	testCases := []struct {
		str          string
		Data         map[string]string
		ExpectResult common.Var
		ExpectErr    bool
	}{
		{
			str:          "range(1, 10, 2)",
			ExpectResult: common.Var{1, 3, 5, 7, 9},
			ExpectErr:    false,
		},
		{
			str:          "range(1, 10)",
			ExpectResult: common.Var{1, 2, 3, 4, 5, 6, 7, 8, 9},
			ExpectErr:    false,
		},
		{
			str:          "range(1, 10, -1)",
			ExpectResult: nil,
			ExpectErr:    true,
		},
		{
			str:          "range(1, ${end}, 2)",
			Data:         map[string]string{"end": "10"},
			ExpectResult: common.Var{1, 3, 5, 7, 9},
			ExpectErr:    false,
		},
		{
			str:          "range(${start}, ${end}, 2)",
			Data:         map[string]string{"start": "1", "end": "10"},
			ExpectResult: common.Var{1, 3, 5, 7, 9},
			ExpectErr:    false,
		},
		{
			str:          "range(${start}, ${end}, 2)",
			Data:         map[string]string{"start": "15", "end": "10"},
			ExpectResult: nil,
			ExpectErr:    true,
		},
		{
			str:          "range(${start}, ${end}, 2)",
			Data:         map[string]string{"start": "start", "end": "10"},
			ExpectResult: nil,
			ExpectErr:    true,
		},
	}

	for i, testCase := range testCases {
		result, err := InstantiateRangeFunc("test"+strconv.Itoa(i), testCase.str, testCase.Data)
		if testCase.ExpectErr == true && err == nil {
			t.Errorf("testCase %d: ExpectStart error, but got nil", i)
		}
		if testCase.ExpectErr == false && err != nil {
			t.Errorf("testCase %d: ExpectStart no error, but got error %v", i, err)
		}
		if !compareVar(result, testCase.ExpectResult) {
			t.Errorf("testCase %d: ExpectStart result %v, but got %v", i, testCase.ExpectResult, result)
		}
	}
}

func compareVar(a common.Var, b common.Var) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		aval := v.(float64)
		bval := b[i].(int)

		if int(aval) != bval {
			return false
		}
	}
	return true
}
