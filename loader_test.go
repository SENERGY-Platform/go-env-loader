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
	Var1  string            `env_var:"VAR_1"`
	Var2  int               `env_var:"VAR_2"`
	Var3  int8              `env_var:"VAR_3"`
	Var4  int16             `env_var:"VAR_4"`
	Var5  int32             `env_var:"VAR_5"`
	Var6  int64             `env_var:"VAR_6"`
	Var7  uint              `env_var:"VAR_7"`
	Var8  uint8             `env_var:"VAR_8"`
	Var9  uint16            `env_var:"VAR_9"`
	Var10 uint32            `env_var:"VAR_10"`
	Var11 uint64            `env_var:"VAR_11"`
	Var12 float32           `env_var:"VAR_12"`
	Var13 float64           `env_var:"VAR_13"`
	Var14 []string          `env_var:"VAR_14"`
	Var15 map[string]string `env_var:"VAR_15"`
	Var16 []TestItem        `env_var:"VAR_16"`
	Var17 TestSubStruct     `env_var:"VAR_17"`
	Var18 string
	Var19 complex64  `env_var:"VAR_19"`
	Var20 complex128 `env_var:"VAR_20"`
}

const (
	defaultString  string     = "default"
	testString     string     = "test"
	testInt64      int64      = 1
	testFloat64    float64    = 1.0
	testComplex128 complex128 = 2 - 3i
)

func newTestStruct() TestStruct {
	return TestStruct{
		Var1:  defaultString,
		Var18: defaultString,
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

func TestLoadInt(t *testing.T) {
	intStr := strconv.FormatInt(testInt64, 10)
	testCaseA := []TestCaseA{
		{
			a:   intStr,
			env: "VAR_2",
		},
		{
			a:   intStr,
			env: "VAR_3",
		},
		{
			a:   intStr,
			env: "VAR_4",
		},
		{
			a:   intStr,
			env: "VAR_5",
		},
		{
			a:   intStr,
			env: "VAR_6",
		},
		{
			a:   intStr,
			env: "VAR_7",
		},
		{
			a:   intStr,
			env: "VAR_8",
		},
		{
			a:   intStr,
			env: "VAR_9",
		},
		{
			a:   intStr,
			env: "VAR_10",
		},
		{
			a:   intStr,
			env: "VAR_11",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var2,
			want: int(testInt64),
		},
		{
			b:    testStruct.Var3,
			want: int8(testInt64),
		},
		{
			b:    testStruct.Var4,
			want: int16(testInt64),
		},
		{
			b:    testStruct.Var5,
			want: int32(testInt64),
		},
		{
			b:    testStruct.Var6,
			want: testInt64,
		},
		{
			b:    testStruct.Var7,
			want: uint(testInt64),
		},
		{
			b:    testStruct.Var8,
			want: uint8(testInt64),
		},
		{
			b:    testStruct.Var9,
			want: uint16(testInt64),
		},
		{
			b:    testStruct.Var10,
			want: uint32(testInt64),
		},
		{
			b:    testStruct.Var11,
			want: uint64(testInt64),
		},
	}
	testValues(t, testCasesB)
}

func TestLoadFloat(t *testing.T) {
	testCaseA := []TestCaseA{
		{
			a:   strconv.FormatFloat(testFloat64, 'f', 1, 32),
			env: "VAR_12",
		},
		{
			a:   strconv.FormatFloat(testFloat64, 'f', 1, 64),
			env: "VAR_13",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var12,
			want: float32(testFloat64),
		},
		{
			b:    testStruct.Var13,
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
			env: "VAR_14",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var14,
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
			env: "VAR_15",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var15,
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
			env: "VAR_16",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var16,
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
			env: "VAR_17",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var17,
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
			b:    testStruct.Var17.Var,
			want: testString,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadNoTag(t *testing.T) {
	testCasesA := []TestCaseA{
		{
			a:   testString,
			env: "VAR_18",
		},
	}
	testStruct := initTestStruct(t, testCasesA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var18,
			want: defaultString,
		},
	}
	testValues(t, testCasesB)
}

func TestLoadComplex(t *testing.T) {
	testCaseA := []TestCaseA{
		{
			a:   strconv.FormatComplex(testComplex128, 'f', 1, 64),
			env: "VAR_19",
		},
		{
			a:   strconv.FormatComplex(testComplex128, 'f', 1, 128),
			env: "VAR_20",
		},
	}
	testStruct := initTestStruct(t, testCaseA)
	testCasesB := []TestCaseB{
		{
			b:    testStruct.Var19,
			want: complex64(testComplex128),
		},
		{
			b:    testStruct.Var20,
			want: testComplex128,
		},
	}
	testValues(t, testCasesB)
}
