/*
   Copyright 2022 Yann Dumont

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

package envldr

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"
)

type TestCaseA struct {
	a, env string
}

type TestCaseB struct {
	b, want interface{}
}

func setEnv(testCases []TestCaseA) (err error) {
	for _, testCase := range testCases {
		if err = os.Setenv(testCase.env, testCase.a); err != nil {
			return
		}
	}
	return
}

func unsetEnv(testCases []TestCaseA) (err error) {
	for _, testCase := range testCases {
		if err = os.Unsetenv(testCase.env); err != nil {
			return
		}
	}
	return
}

func testValues(t *testing.T, testCases []TestCaseB) {
	for _, testCase := range testCases {
		if !reflect.DeepEqual(testCase.b, testCase.want) {
			t.Errorf("b = %s; want %s", testCase.b, testCase.want)
		}
	}
}

type TestItem struct {
	Var string
}

type TestSubStruct struct {
	Var string `env_var:"SUB_VAR"`
}

type TestStruct struct {
	Var1 string            `env_var:"VAR_1"`
	Var2 int64             `env_var:"VAR_2"`
	Var3 float64           `env_var:"VAR_3"`
	Var4 []string          `env_var:"VAR_4"`
	Var5 map[string]string `env_var:"VAR_5"`
	Var6 []TestItem        `env_var:"VAR_6"`
	Var7 TestSubStruct     `env_var:"VAR_7"`
	Var8 string
}

const (
	defaultString string  = "default"
	testString    string  = "test"
	testInt64     int64   = 1
	testFloat64   float64 = 1.0
)

func newTestStruct() TestStruct {
	return TestStruct{
		Var1: defaultString,
		Var8: defaultString,
	}
}

func initTestStruct(t *testing.T, testCasesA []TestCaseA) TestStruct {
	if testCasesA != nil {
		if err := setEnv(testCasesA); err != nil {
			panic(err)
		}
	}
	testStruct := newTestStruct()
	if err := LoadEnv(&testStruct); err != nil {
		t.Error(err)
	}
	if testCasesA != nil {
		if err := unsetEnv(testCasesA); err != nil {
			panic(err)
		}
	}
	return testStruct
}

func TestDefaultValue(t *testing.T) {
	testStruct := initTestStruct(t, nil)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var1,
			want: defaultString,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadString(t *testing.T) {
	testCaseA := []TestCaseA{
		{
			a:   testString,
			env: "VAR_1",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var1,
			want: testString,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadInt64(t *testing.T) {
	testCaseA := []TestCaseA{
		{
			a:   strconv.FormatInt(testInt64, 10),
			env: "VAR_2",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var2,
			want: testInt64,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadFloat64(t *testing.T) {
	testCaseA := []TestCaseA{
		{
			a:   strconv.FormatFloat(testFloat64, 'f', 1, 64),
			env: "VAR_3",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var3,
			want: testFloat64,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadSlice(t *testing.T) {
	testSlice := []string{testString}
	testSliceByte, err := json.Marshal(testSlice)
	if err != nil {
		panic(err)
	}
	testCaseA := []TestCaseA{
		{
			a:   string(testSliceByte),
			env: "VAR_4",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var4,
			want: testSlice,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadMap(t *testing.T) {
	testMap := map[string]string{testString: testString}
	testMapByte, err := json.Marshal(testMap)
	if err != nil {
		panic(err)
	}
	testCaseA := []TestCaseA{
		{
			a:   string(testMapByte),
			env: "VAR_5",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var5,
			want: testMap,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadStructSlice(t *testing.T) {
	testStructSlice := []TestItem{{Var: testString}}
	testStructSliceByte, err := json.Marshal(testStructSlice)
	if err != nil {
		panic(err)
	}
	testCaseA := []TestCaseA{
		{
			a:   string(testStructSliceByte),
			env: "VAR_6",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var6,
			want: testStructSlice,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadStruct(t *testing.T) {
	testSubStruct := TestSubStruct{Var: testString}
	testStructByte, err := json.Marshal(testSubStruct)
	if err != nil {
		panic(err)
	}
	testCaseA := []TestCaseA{
		{
			a:   string(testStructByte),
			env: "VAR_7",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var7,
			want: testSubStruct,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadSubStructVar(t *testing.T) {
	testCasesA := []TestCaseA{
		{
			a:   testString,
			env: "SUB_VAR",
		},
	}
	testStruct := initTestStruct(t, testCasesA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var7.Var,
			want: testString,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadNoTag(t *testing.T) {
	testCasesA := []TestCaseA{
		{
			a:   testString,
			env: "VAR_8",
		},
	}
	testStruct := initTestStruct(t, testCasesA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var8,
			want: defaultString,
		},
	}
	testValues(t, testCasesB)
}
