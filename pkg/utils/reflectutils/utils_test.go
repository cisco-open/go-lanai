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

package reflectutils

import (
	"context"
	"github.com/cisco-open/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"reflect"
	"testing"
)

func TestStructAnalysis(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestFindStructField(), "TestFindStructField"),
		test.GomegaSubTest(SubTestListStructField(), "TestListStructField"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestFindStructField() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		sf, ok := FindStructField(reflect.TypeOf(TestSubject{}), func(t reflect.StructField) bool {
			return len(t.Tag.Get("mark")) != 0 && IsExportedField(t)
		})
		g.Expect(ok).To(BeTrue(), "field should be found")
		g.Expect(sf.Tag.Get("mark")).To(Equal("match"), "field's tag should be correct")
		// Note: search is in reversed order
		g.Expect(sf.Index).To(BeEquivalentTo([]int{1, 1, 0}), "field's index should be correct")
	}
}

func SubTestListStructField() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fields := ListStructField(reflect.TypeOf(TestSubject{}), func(t reflect.StructField) bool {
			return len(t.Tag.Get("mark")) != 0
		})
		// Should find all fields, includes shadowed:
		// - 2 from ExportedEmbedded -> ExportedNestedEmbedded
		// - 2 from ExportedEmbedded -> unexportedNestedEmbedded
		// - 2 from unexportedEmbedded -> ExportedNestedEmbedded
		// - 2 from unexportedEmbedded -> unexportedNestedEmbedded
		g.Expect(fields).To(HaveLen(8), "fields should be correct")
	}
}

/*************************
	Helpers
 *************************/

type TestSubject struct {
	ExportedEmbedded
	unexportedEmbedded
	ExportedField1   string
	ExportedField2   ExportedField
	ExportedField3   unexportedField
	unexportedField1 string
	unexportedField2 ExportedField
	unexportedField3 unexportedField
}

type ExportedEmbedded struct {
	ExportedNestedEmbedded
	unexportedNestedEmbedded
	ExportedField4   string
	ExportedField5   ExportedField
	ExportedField6   unexportedField
	unexportedField4 string
	unexportedField5 ExportedField
	unexportedField6 unexportedField
}

type unexportedEmbedded struct {
	ExportedNestedEmbedded
	unexportedNestedEmbedded
	ExportedField4   string
	ExportedField5   ExportedField
	ExportedField6   unexportedField
	unexportedField4 string
	unexportedField5 ExportedField
	unexportedField6 unexportedField
}

type ExportedNestedEmbedded struct {
	ExportedField1   string `mark:"match"`
	ExportedField2   ExportedField
	ExportedField3   unexportedField
	unexportedField1 string `mark:"match"`
	unexportedField2 ExportedField
	unexportedField3 unexportedField
}

type unexportedNestedEmbedded struct {
	ExportedField1   string `mark:"match"`
	ExportedField2   ExportedField
	ExportedField3   unexportedField
	unexportedField1 string `mark:"match"`
	unexportedField2 ExportedField
	unexportedField3 unexportedField
}

type ExportedField struct {
	Exported   string `mark:"shouldn't match'"`
	unexported int    `mark:"shouldn't match'"`
}

type unexportedField struct {
	Exported   string `mark:"shouldn't match'"`
	unexported int    `mark:"shouldn't match'"`
}
