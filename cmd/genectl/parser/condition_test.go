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
	"testing"
)

func TestIsCheckResultFunc(t *testing.T) {
	testCases := []struct {
		str    string
		expect bool
	}{
		{
			str:    "check_result(job-target, 'c')",
			expect: true,
		},
		{
			str:    "check_result(job-a, ${npart})",
			expect: true,
		},
		{
			str:    "check_result(job-a, \"\n\")",
			expect: true,
		},
		{
			str:    "check_result(1, \"10\")",
			expect: true,
		},
		{
			str:    "check_result(1:,10)",
			expect: false,
		},
		{
			str:    "check_result(, 10)",
			expect: false,
		},
		{
			str:    "check_result()",
			expect: false,
		},
	}

	for i, testCase := range testCases {
		result := IsCheckResultFunc(testCase.str)
		if result != testCase.expect {
			t.Errorf("%d: unexpected result; got %v, expected %v", i, result, testCase.expect)
		}
	}
}

func TestCheckResultFuncParam(t *testing.T) {
	testCases := []struct {
		Str         string
		ExpectjName string
		ExpectSep   string
	}{

		{
			Str:         `check_result({start}, " ")`,
			ExpectjName: "{start}",
			ExpectSep:   " ",
		},
		{
			Str:         `check_result(job-a, "2")`,
			ExpectjName: "job-a",
			ExpectSep:   "2",
		},
		{
			Str:         `check_result(job-b, "${start}")`,
			ExpectjName: "job-b",
			ExpectSep:   "${start}",
		},
	}

	for i, testCase := range testCases {
		jName, sep := checkResultFuncParam(testCase.Str)
		if jName != testCase.ExpectjName {
			t.Errorf("%d: unexpected jobName; got %s, expected %s", i, jName, testCase.ExpectjName)
		}
		if sep != testCase.ExpectSep {
			t.Errorf("%d: unexpected separator; got %s, expected %s", i, sep, testCase.ExpectSep)
		}
	}
}
