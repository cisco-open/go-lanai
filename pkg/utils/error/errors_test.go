package errorutils

import (
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestTypeComparison(t *testing.T) {
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

func TestConcreteError(t *testing.T) {
	g := gomega.NewWithT(t)
	actual := NewCodedError(ErrorCodeA1_1, "whatever message")
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

func TestConcreteTypeError(t *testing.T) {
	g := gomega.NewWithT(t)
	actual := NewCodedError(ErrorSubTypeCodeA2, "whatever message")
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
func tryReserve(v interface{}) (err error){
	defer func() {
		err, _ = recover().(error)
	}()
	Reserve(v)
	return
}

/************************
	Test Error Hierarchy
 ************************/
const (
	// security reserved
	Reserved        		= 998 << ReservedOffset

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
	_ = iota
	ErrorTypeCodeA = Reserved + iota << ErrorTypeOffset
	ErrorTypeCodeB
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeA
const (
	_ = iota
	ErrorSubTypeCodeA1             = ErrorTypeCodeA + iota << ErrorSubTypeOffset
	ErrorSubTypeCodeA2
)

// ErrorSubTypeCodeA1
//goland:noinspection GoSnakeCaseUsage
const (
	_ = iota
	ErrorCodeA1_1 = ErrorSubTypeCodeA1 + iota
	ErrorCodeA1_2
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeB
const (
	_ = iota
	ErrorSubTypeCodeB1 = ErrorTypeCodeB + iota << ErrorSubTypeOffset
	ErrorSubTypeCodeB2
)

// ErrorSubTypeCodeB1
//goland:noinspection GoSnakeCaseUsage
const (
	_ = iota
	ErrorCodeB1_1 = ErrorSubTypeCodeB1 + iota
	ErrorCodeB1_2
)

// ErrorTypes, can be used in errors.Is
//goland:noinspection GoSnakeCaseUsage,GoUnusedGlobalVariable
var (
	// masked errors for comparison
	ErrorCategoryTest = NewErrorCategory(Reserved, errors.New("error type: test"))
	ErrorTypeA        = NewErrorType(ErrorTypeCodeA, errors.New("error type: A"))
	ErrorTypeB        = NewErrorType(ErrorTypeCodeB, errors.New("error type: B"))

	ErrorSubTypeA1 = NewErrorSubType(ErrorSubTypeCodeA1, errors.New("error sub-type: A1"))
	ErrorSubTypeA2 = NewErrorSubType(ErrorSubTypeCodeA2, errors.New("error sub-type: A2"))
	ErrorSubTypeB1 = NewErrorSubType(ErrorSubTypeCodeB1, errors.New("error sub-type: B1"))
	ErrorSubTypeB2 = NewErrorSubType(ErrorSubTypeCodeB2, errors.New("error sub-type: B2"))

	// causes
	CauseA1_1 = errors.New("error: A1-1")
	CauseA1_2 = errors.New("error: A1-2")
	CauseA2 = errors.New("error: A2")
	CauseB1_1 = errors.New("error: B1-1")
	CauseB1_2 = errors.New("error: B1-2")
	CauseB2 = errors.New("error: B2")

	// concrete errors
	ErrorA1_1 = NewCodedError(ErrorCodeA1_1, CauseA1_1)
	ErrorA1_2 = NewCodedError(ErrorCodeA1_2, CauseA1_2)
	ErrorA2   = NewCodedError(ErrorSubTypeCodeA2, CauseA2)
	ErrorB1_1 = NewCodedError(ErrorCodeB1_1, CauseB1_1)
	ErrorB1_2 = NewCodedError(ErrorCodeB1_2, CauseB1_2)
	ErrorB2   = NewCodedError(ErrorSubTypeCodeB2, CauseB2)
)