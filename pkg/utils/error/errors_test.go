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

package errorutils

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCodedErrorTypeComparison(t *testing.T) {
	g := gomega.NewWithT(t)
	g.Expect(errors.Is(ErrorTypeA, ErrorCategoryTest)).To(BeTrue(), "ErrorType should match ErrorCategoryTest")
	g.Expect(errors.Is(ErrorTypeA, ErrorTypeA)).To(BeTrue(), "ErrorType should match itself")
	g.Expect(errors.Is(ErrorSubTypeA1, ErrorTypeA)).To(BeTrue(), "ErrorSubType should match its own ErrorType")
	g.Expect(errors.Is(ErrorTypeA, ErrorSubTypeA1)).To(BeFalse(), "ErrorType should not match its own ErrorSubType")
	g.Expect(errors.Is(ErrorTypeA, ErrorTypeB)).To(BeFalse(), "Different ErrorType should not match each other")
	g.Expect(errors.Is(ErrorSubTypeA1, ErrorSubTypeA2)).To(BeFalse(), "Different ErrorSubType should not match each other")
	g.Expect(errors.Is(ErrorSubTypeB2, ErrorTypeB)).To(BeTrue(), "ErrorSubTypeB1 should be ErrorTypeB error")

	g.Expect(ErrorTypeA.Cause()).To(BeNil(), "ErrorType shouldn't have cause")
	g.Expect(ErrorTypeA.RootCause()).To(BeNil(), "ErrorType shouldn't have root cause")
	g.Expect(ErrorTypeA.Cause()).To(BeNil(), "ErrorSubType shouldn't have root cause")
	g.Expect(ErrorTypeA.RootCause()).To(BeNil(), "ErrorSubType shouldn't have root cause")
}

func TestConcreteCodedError(t *testing.T) {
	g := gomega.NewWithT(t)
	// test marshalling of concrete error
	orig := NewCodedError(ErrorCodeA1_1, "whatever message")
	actual := assertGobEncodeAndDecode(g, orig)

	ref := ErrorA1_1
	cause := CauseA1_1
	nonCoded := errors.New("non-coded error")

	g.Expect(errors.Is(actual, ErrorCategoryTest)).To(BeTrue(), "actual error should match ErrorCategoryTest")
	g.Expect(errors.Is(actual, ErrorTypeA)).To(BeTrue(), "actual error should match ErrorTypeA")
	g.Expect(errors.Is(actual, ErrorSubTypeA1)).To(BeTrue(), "actual error should match ErrorSubTypeA1")
	g.Expect(errors.Is(actual, ref)).To(BeTrue(), "Two ErrorA1_1 should match each other regardless of actual message")
	g.Expect(errors.Is(ref, actual)).To(BeTrue(), "Two ErrorA1_1 should match each other regardless of actual message and comparison order")
	g.Expect(errors.Is(ref, cause)).To(BeTrue(), "ErrorA1_1 should match its cause")
	g.Expect(errors.Is(cause, ref)).To(BeFalse(), "cause should not match its CodedError counterpart (reversed comparison)")

	g.Expect(errors.Is(actual, nonCoded)).To(BeFalse(), "actual error should not match non-coded error")
	g.Expect(errors.Is(actual, ErrorTypeB)).To(BeFalse(), "actual error should not match ErrorTypeB")
	g.Expect(errors.Is(actual, ErrorSubTypeA2)).To(BeFalse(), "actual error should not match ErrorSubTypeA2")
	g.Expect(errors.Is(actual, ErrorSubTypeB1)).To(BeFalse(), "actual error should not match ErrorSubTypeB1")
}

func TestConcreteCodedTypeError(t *testing.T) {
	g := gomega.NewWithT(t)
	// test marshalling of concrete error
	orig := NewCodedError(ErrorSubTypeCodeA2, "whatever message")
	actual := assertGobEncodeAndDecode(g, orig)

	ref := ErrorA2
	cause := CauseA2
	nonCoded := errors.New("non-coded error")

	g.Expect(errors.Is(actual, ErrorCategoryTest)).To(BeTrue(), "actual error should match ErrorCategoryTest")
	g.Expect(errors.Is(actual, ErrorTypeA)).To(BeTrue(), "actual error should match ErrorTypeA")
	g.Expect(errors.Is(actual, ErrorSubTypeA2)).To(BeTrue(), "actual error should match ErrorSubTypeA2")
	g.Expect(errors.Is(actual, ref)).To(BeTrue(), "Two ErrorA2 should match each other regardless of actual message")
	g.Expect(errors.Is(ref, actual)).To(BeTrue(), "Two ErrorA2 should match each other regardless of actual message and comparison order")
	g.Expect(errors.Is(ref, cause)).To(BeTrue(), "ErrorA2 should match its cause")
	g.Expect(errors.Is(cause, ref)).To(BeFalse(), "cause should not match its CodedError counterpart (reversed comparison)")

	g.Expect(errors.Is(actual, nonCoded)).To(BeFalse(), "actual error should not match non-coded error")
	g.Expect(errors.Is(actual, ErrorTypeB)).To(BeFalse(), "actual error should not match ErrorTypeB")
	g.Expect(errors.Is(actual, ErrorSubTypeA1)).To(BeFalse(), "actual error should not match ErrorSubTypeA1")
	g.Expect(errors.Is(actual, ErrorSubTypeB1)).To(BeFalse(), "actual error should not match ErrorSubTypeB1")

	// Note: compare subtype's concrete error as target is undefined behavior, we don't check following
	//g.Expect(errors.Is(ErrorSubTypeA2, actual)).To(BeFalse(), "concrete sub type error should not match its own type as a target")
}

func TestCodedErrorWithCausesChain(t *testing.T) {
	g := gomega.NewWithT(t)
	causeDirect := ErrorB1_1
	causeIntermediate := ErrorB1_2
	causeRoot := ErrorA1_1
	expectedCauseRoot := CauseA1_1
	// Note: by design, the nested errors are not serialized
	actual := NewCodedError(ErrorSubTypeCodeA2, causeDirect, causeIntermediate, causeRoot)
	_ = assertGobEncodeAndDecode(g, actual)

	ref := ErrorA2
	nonCoded := errors.New("non-coded error")

	g.Expect(errors.Is(actual, ErrorCategoryTest)).To(BeTrue(), "actual error should match ErrorCategoryTest")
	g.Expect(errors.Is(actual, ErrorTypeA)).To(BeTrue(), "actual error should match ErrorTypeA")
	g.Expect(errors.Is(actual, ErrorSubTypeA2)).To(BeTrue(), "actual error should match ErrorSubTypeA2")
	g.Expect(errors.Is(actual, ref)).To(BeTrue(), "two ErrorA2 should match each other regardless of actual message")
	g.Expect(errors.Is(ref, actual)).To(BeTrue(), "two ErrorA2 should match each other regardless of actual message and comparison order")

	g.Expect(errors.Is(actual, nonCoded)).To(BeFalse(), "actual error should not match non-coded error")
	g.Expect(errors.Is(actual, causeDirect)).To(BeFalse(), "actual error should not match error's direct cause")
	g.Expect(errors.Is(actual, causeIntermediate)).To(BeFalse(), "actual error should not match error's intermediate cause")
	g.Expect(errors.Is(actual, causeRoot)).To(BeFalse(), "actual error should not match error's root cause")
	g.Expect(errors.Is(actual, ErrorTypeB)).To(BeFalse(), "actual error should not match coded cause's type (ErrorTypeB)")
	g.Expect(errors.Is(actual, ErrorSubTypeA1)).To(BeFalse(), "actual error should not match ErrorSubTypeA1")
	g.Expect(errors.Is(actual, ErrorSubTypeB1)).To(BeFalse(), "actual error should not match ErrorSubTypeB1")

	g.Expect(errors.Is(actual.Cause(), causeDirect)).To(BeTrue(), "actual error's direct cause should be correct")
	g.Expect(errors.Is(actual.RootCause(), expectedCauseRoot)).To(BeTrue(), "actual error's root cause should be correct")
}

func TestWrappedError(t *testing.T) {
	g := gomega.NewWithT(t)
	actual := ErrorA1X1.WithMessage("this is another instance")
	ref := ErrorA1X1
	another := NewWrappedTestError(actual.ErrIs, ErrorSubTypeB2)
	randErr := errors.New("some error")
	coded := ErrorA1_1

	g.Expect(errors.Is(actual, ErrorCategoryTest)).To(BeTrue(), "actual error should match ErrorCategoryTest")
	g.Expect(errors.Is(actual, ErrorTypeA)).To(BeTrue(), "actual error should match ErrorTypeA")
	g.Expect(errors.Is(actual, ErrorSubTypeA1)).To(BeTrue(), "actual error should match ErrorSubTypeA1")
	g.Expect(errors.Is(actual, ref)).To(BeTrue(), "two ErrorA2 should match each other regardless of actual message")
	g.Expect(errors.Is(ref, actual)).To(BeTrue(), "two ErrorA2 should match each other regardless of actual message and comparison order")

	g.Expect(errors.Is(actual, randErr)).To(BeFalse(), "actual error should not match random error")
	g.Expect(errors.Is(actual, coded)).To(BeFalse(), "actual error should not match coded error within same hierarchy")
	g.Expect(errors.Is(actual, actual.ErrIs)).To(BeFalse(), "actual error should not match error's internal")
	g.Expect(errors.Is(actual, another)).To(BeFalse(), "actual error should not match another error with different type")

	g.Expect(errors.Is(actual, ErrorTypeB)).To(BeFalse(), "actual error should not match coded cause's type (ErrorTypeB)")
	g.Expect(errors.Is(actual, ErrorSubTypeA2)).To(BeFalse(), "actual error should not match ErrorSubTypeA2")
	g.Expect(errors.Is(actual, ErrorSubTypeB1)).To(BeFalse(), "actual error should not match ErrorSubTypeB1")

	// Note: compare subtype's concrete error as target is undefined behavior, we don't check following
	//concrete := actual.Type.WithMessage("concrete subtype")
	//g.Expect(errors.Is(actual, concrete)).To(BeFalse(), "actual error should not match concrete version of its own type")
}

func TestCodeReserve(t *testing.T) {
	g := gomega.NewWithT(t)
	g.Expect(tryReserve(ErrorTypeA)).To(HaveOccurred(), "reserve non-category error should panic")
	g.Expect(tryReserve(ErrorCategoryTest)).To(Succeed(), "reserve category error for the first time should not panic")
	g.Expect(tryReserve(ErrorCategoryTest)).To(HaveOccurred(), "reserve category error with same code should panic")
}

/************************
	Helpers
 ************************/
func tryReserve(v interface{}) (err error) {
	defer func() {
		err, _ = recover().(error)
	}()
	Reserve(v)
	return
}

func assertGobEncodeAndDecode[T error](g *gomega.WithT, orig T) T {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf).Encode(orig)
	g.Expect(e).To(Succeed(), "encoding error to binary should not fail")
	var decoded T
	e = gob.NewDecoder(&buf).Decode(&decoded)
	g.Expect(e).To(Succeed(), "decoding error from binary should not fail")
	g.Expect(errors.Is(decoded, orig), "decoded error should be original error")
	return decoded
}

/************************
	Test Error Hierarchy
 ************************/
const (
	// security reserved
	testReserved = 998 << ReservedOffset
)

// Hierarchy:
// Category = 998
// |-- A
//     |-- A1
//         |-- A1_1
//         |-- A1_2
//     |-- A2
// |-- B
//     |-- B1
//         |-- B1_1
//         |-- B1_2
//     |-- B2
// All "Type" values are used as mask
const (
	_              = iota
	ErrorTypeCodeA = testReserved + iota<<ErrorTypeOffset
	ErrorTypeCodeB
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeA
const (
	_                  = iota
	ErrorSubTypeCodeA1 = ErrorTypeCodeA + iota<<ErrorSubTypeOffset
	ErrorSubTypeCodeA2
)

// ErrorSubTypeCodeA1
//goland:noinspection GoSnakeCaseUsage
const (
	_             = iota
	ErrorCodeA1_1 = ErrorSubTypeCodeA1 + iota
	ErrorCodeA1_2
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeB
const (
	_                  = iota
	ErrorSubTypeCodeB1 = ErrorTypeCodeB + iota<<ErrorSubTypeOffset
	ErrorSubTypeCodeB2
)

// ErrorSubTypeCodeB1
//goland:noinspection GoSnakeCaseUsage
const (
	_             = iota
	ErrorCodeB1_1 = ErrorSubTypeCodeB1 + iota
	ErrorCodeB1_2
)

// ErrorTypes, can be used in errors.Is
//goland:noinspection GoSnakeCaseUsage,GoUnusedGlobalVariable
var (
	// masked errors for comparison
	ErrorCategoryTest = NewErrorCategory(testReserved, errors.New("error type: test"))
	ErrorTypeA        = NewErrorType(ErrorTypeCodeA, errors.New("error type: A"))
	ErrorTypeB        = NewErrorType(ErrorTypeCodeB, errors.New("error type: B"))

	ErrorSubTypeA1 = NewErrorSubType(ErrorSubTypeCodeA1, errors.New("error sub-type: A1"))
	ErrorSubTypeA2 = NewErrorSubType(ErrorSubTypeCodeA2, errors.New("error sub-type: A2"))
	ErrorSubTypeB1 = NewErrorSubType(ErrorSubTypeCodeB1, errors.New("error sub-type: B1"))
	ErrorSubTypeB2 = NewErrorSubType(ErrorSubTypeCodeB2, errors.New("error sub-type: B2"))

	// causes
	CauseA1_1 = errors.New("error: A1-1")
	CauseA1_2 = errors.New("error: A1-2")
	CauseA2   = errors.New("error: A2")
	CauseB1_1 = errors.New("error: B1-1")
	CauseB1_2 = errors.New("error: B1-2")
	CauseB2   = errors.New("error: B2")
	CauseA1X1 = errors.New("error: A1X1")
	CauseA1X2 = errors.New("error: A1X1")
	CauseB1X1 = errors.New("error: B1X1")
	CauseB1X2 = errors.New("error: B1X2")

	// concrete errors
	ErrorA1_1 = NewTestError(ErrorCodeA1_1, CauseA1_1)
	ErrorA1_2 = NewTestError(ErrorCodeA1_2, CauseA1_2)
	ErrorA2   = NewTestError(ErrorSubTypeCodeA2, CauseA2)
	ErrorB1_1 = NewTestError(ErrorCodeB1_1, CauseB1_1)
	ErrorB1_2 = NewTestError(ErrorCodeB1_2, CauseB1_2)
	ErrorB2   = NewTestError(ErrorSubTypeCodeB2, CauseB2)

	ErrorA1X1 = NewWrappedTestError(CauseA1X1, ErrorSubTypeA1)
	ErrorA1X2 = NewWrappedTestError(CauseA1X2, ErrorSubTypeA1)
	ErrorB1X1 = NewWrappedTestError(CauseB1X1, ErrorSubTypeB1)
	ErrorB1X2 = NewWrappedTestError(CauseB1X2, ErrorSubTypeB1)
)

func NewTestError(code int64, e interface{}) *CodedTestError {
	return &CodedTestError{
		CodedError: *NewCodedError(code, e),
	}
}

type CodedTestError struct {
	CodedError
}

func (e CodedTestError) Unwrap() error {
	return e.CodedError.Nested
}

type WrappedTestError struct {
	WrappedError
}

func NewWrappedTestError(is error, typ error) WrappedTestError {
	return WrappedTestError{
		WrappedError{
			ErrIs:  is,
			Type:   typ.(*CodedError),
			ErrMsg: is.Error(),
		},
	}
}
