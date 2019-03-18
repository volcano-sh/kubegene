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

package parser

import (
	"fmt"
	"testing"
)

func TestIsGetResultFuncFunc(t *testing.T) {
	testCases := []struct {
		str    string
		expect bool
	}{
		{
			str:    "get_result(job-target)",
			expect: true,
		},
		{
			str:    "get_result(job-target, 'c')",
			expect: true,
		},
		{
			str:    "get_result(job-a, ${npart})",
			expect: true,
		},
		{
			str:    "get_result(job-a, \"\n\")",
			expect: true,
		},
		{
			str:    "get_result(1, \"10\")",
			expect: true,
		},
		{
			str:    "get_result(1:,10)",
			expect: false,
		},
		{
			str:    "get_result(, 10)",
			expect: false,
		},
		{
			str:    "get_result()",
			expect: false,
		},
	}

	for i, testCase := range testCases {
		result := IsGetResultFunc(testCase.str)
		if result != testCase.expect {
			t.Errorf("%d: unexpected result; got %v, expected %v", i, result, testCase.expect)
		}
	}
}

func TestGetgetResultFuncParam(t *testing.T) {
	testCases := []struct {
		Str         string
		ExpectjName string
		ExpectSep   string
	}{
		{
			Str:         "get_result(job-target)",
			ExpectjName: "job-target",
			ExpectSep:   "",
		},

		{
			Str:         `get_result({start}, " ")`,
			ExpectjName: "{start}",
			ExpectSep:   " ",
		},
	}

	for i, testCase := range testCases {
		jName, sep := GetgetResultFuncParam(testCase.Str)
		if jName != testCase.ExpectjName {
			t.Errorf("%d: unexpected jobName; got %s, expected %s", i, jName, testCase.ExpectjName)
		}
		if sep != testCase.ExpectSep {
			t.Errorf("%d: unexpected separator; got %s, expected %s", i, sep, testCase.ExpectSep)
		}
		fmt.Println("succes===")

	}
}
