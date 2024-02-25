// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package golden

import (
    "context"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/onsi/gomega"
    "os"
    "path/filepath"
    "testing"
)

func TestGetGoldenFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test a standard non subtest test",
			want: filepath.Join("testdata", "golden", "TestGetGoldenFilePath", "test_a_standard_non_subtest_test.json"),
		},
		{
			name: "test_a_standard_non_subtest_test_with_snake_case_already",
			want: filepath.Join("testdata", "golden", "TestGetGoldenFilePath", "test_a_standard_non_subtest_test_with_snake_case_already.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGoldenFilePath(t); got != tt.want {
				t.Errorf("GetGoldenFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithAppTest(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		test.GomegaSubTest(SubTestWithoutTableDriven(), "SubTestWithoutTableDriven"),
		test.GomegaSubTest(SubTestWithTableDriven(), "SubTestWithTableDriven"),
	)
}

func SubTestWithoutTableDriven() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		expectedPath := filepath.Join("testdata", "golden", "TestWithAppTest", "sub_test_without_table_driven.json")
		if got := GetGoldenFilePath(t); got != expectedPath {
			t.Errorf("GetGoldenFilePath() = %v, want %v", got, expectedPath)
		}
	}
}

func SubTestWithTableDriven() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tests := []struct {
			name string
			want string
		}{
			{
				name: "test a standard non subtest test",
				want: filepath.Join("testdata", "golden", "TestWithAppTest", "SubTestWithTableDriven", "test_a_standard_non_subtest_test.json"),
			},
			{
				name: "test_a_standard_non_subtest_test_with_snake_case_already",
				want: filepath.Join("testdata", "golden", "TestWithAppTest", "SubTestWithTableDriven", "test_a_standard_non_subtest_test_with_snake_case_already.json"),
			},
			{
				name: "TestName, ToBe.snake case with lowerCase",
				want: filepath.Join("testdata", "golden", "TestWithAppTest", "SubTestWithTableDriven", "test_name,_to_be_snake_case_with_lower_case.json"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := GetGoldenFilePath(t); got != tt.want {
					t.Errorf("GetGoldenFilePath() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}

type MockTestingT struct {
	*testing.T
	Failed bool
}

const (
	MockFatalPanic = "Mock Fatal Panic"
)

// Fatalf will not actually fatal the test, but will panic to exit the execution
func (c *MockTestingT) Fatalf(format string, args ...any) {
	c.Failed = true
	panic(MockFatalPanic)
}

// Errorf will not actually fail the test
func (c *MockTestingT) Errorf(format string, args ...any) {
	c.Failed = true
}

type GoodStruct struct {
	Hello string
}

type NoJsonStruct GoodStruct

func (NoJsonStruct) MarshalJSON() ([]byte, error) {
	return nil, errors.New("oops")
}

func TestAssert(t *testing.T) {
	type args struct {
		t    *testing.T
		data interface{}
	}
	tests := []struct {
		name       string
		args       args
		expectFail bool
	}{
		{
			name: "Test Correct Struct Data, expects no error",
			args: args{
				t: t,
				data: GoodStruct{
					Hello: "some string",
				},
			},
		},
		{
			name: "Test Incorrect Struct Data, expects no error",
			args: args{
				t: t,
				data: GoodStruct{
					Hello: "something else",
				},
			},
			expectFail: true,
		},
		{
			name: "Test nil Data, expects fatal error",
			args: args{
				t:    t,
				data: nil,
			},
			expectFail: true,
		},
		{
			name: "Test Non Struct Data, expects fatal error",
			args: args{
				t:    t,
				data: []byte("hello"),
			},
			expectFail: true,
		},
		{
			name: "Test Struct Data with marshalling error, expects fatal error",
			args: args{
				t:    t,
				data: NoJsonStruct{Hello: "hi"},
			},
			expectFail: true,
		},
	}
	for _, tt := range tests {
		mT := MockTestingT{t, false}
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && r != MockFatalPanic {
					t.Errorf("only expected %v, did not expect :%v panic", MockFatalPanic, r)
				}
				if tt.expectFail != mT.Failed {
					t.Errorf("expected fatal error but did not receive fatal error")
				}
			}()
			Assert(&mT, tt.args.data)
		})
	}
}

const PopulateTestOutputDir = `testdata/golden/TestPopulateGoldenFiles`

func TestPopulateGoldenFiles(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.Setup(CleanGoldenOutputDir()),
		test.GomegaSubTest(SubTestPopulateGoldenFiles(GoodStruct{Hello: "string"}, "good_struct.json", true), "GoodStruct"),
		test.GomegaSubTest(SubTestPopulateGoldenFiles(NoJsonStruct{Hello: "string"}, "bad_struct.json", false), "BadStruct"),
		test.GomegaSubTest(SubTestPopulateGoldenFiles("just string", "just_string.json", false), "JustString"),
		test.GomegaSubTest(SubTestPopulateGoldenFiles([]byte("just string"), "just_bytes.json", false), "JustBytes"),
	)
}

func CleanGoldenOutputDir() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		return ctx, os.RemoveAll(PopulateTestOutputDir)
	}
}

func SubTestPopulateGoldenFiles(data interface{}, expectedFile string, expectExists bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		mockT := &MockTestingT{T: t}
		func() {
			defer func() { recover() }()
			PopulateGoldenFiles(mockT, data)
		}()
		expectedPath := fmt.Sprintf(`%s/%s`, PopulateTestOutputDir, expectedFile)
		if expectExists {
			_, e := os.Stat(expectedPath)
			g.Expect(e).To(gomega.Succeed(), "file should exist [%s]", expectedPath)
		} else {
			_, e := os.Stat(expectedPath)
			g.Expect(os.IsNotExist(e)).To(gomega.BeTrue(), "file should not exist [%s]", expectedPath)
		}
	}
}
