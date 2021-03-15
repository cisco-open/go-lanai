package errorutils

import (
	"errors"
	"testing"
)

func TestTypeComparison(t *testing.T) {
	switch {
	case !errors.Is(ErrorTypeA, ErrorCategoryTest):
		t.Errorf("ErrorType should match ErrorCategoryTest")

	case !errors.Is(ErrorTypeA, ErrorTypeA):
		t.Errorf("ErrorType should match itself")

	case !errors.Is(ErrorSubTypeA1, ErrorTypeA):
		t.Errorf("ErrorSubType should match its own ErrorType")

	case errors.Is(ErrorTypeA, ErrorSubTypeA1):
		t.Errorf("ErrorType should not match its own ErrorSubType")

	case errors.Is(ErrorTypeA, ErrorTypeB):
		t.Errorf("Different ErrorType should not match each other")

	case errors.Is(ErrorSubTypeA1, ErrorSubTypeA2):
		t.Errorf("Different ErrorSubType should not match each other")

	case !errors.Is(ErrorSubTypeB2, ErrorTypeB):
		t.Errorf("ErrorSubTypeB1 should be ErrorTypeB error")
	}
}

func TestConcreteError(t *testing.T) {
	actual := NewCodedError(ErrorCodeA1_1, "whatever message")
	ref := ErrorA1_1
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(actual, ErrorCategoryTest):
		t.Errorf("actual error should match ErrorCategoryTest")

	case !errors.Is(actual, ErrorTypeA):
		t.Errorf("actual error should match ErrorTypeA")

	case !errors.Is(actual, ErrorSubTypeA1):
		t.Errorf("actual error should match ErrorSubTypeA1")

	case !errors.Is(actual, ref):
		t.Errorf("Two ErrorA1_1 should match each other regardless of actual message")

	case errors.Is(actual, nonCoded):
		t.Errorf("actual error should not match non-coded error")

	case errors.Is(actual, ErrorTypeB):
		t.Errorf("actual error should not match ErrorTypeB")

	case errors.Is(actual, ErrorSubTypeA2):
		t.Errorf("actual error should not match ErrorSubTypeA2")

	case errors.Is(actual, ErrorSubTypeB1):
		t.Errorf("actual error should not match ErrorSubTypeB1")
	}
}

func TestConcreteTypeError(t *testing.T) {
	actual := NewCodedError(ErrorSubTypeCodeA2, "whatever message")
	ref := ErrorA2
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(actual, ErrorCategoryTest):
		t.Errorf("actual error should match ErrorCategoryTest")

	case !errors.Is(actual, ErrorTypeA):
		t.Errorf("actual error should match ErrorTypeA")

	case !errors.Is(actual, ErrorSubTypeA2):
		t.Errorf("actual error should match ErrorSubTypeA2")

	case !errors.Is(actual, ref):
		t.Errorf("Two ErrorA2 should match each other regardless of actual message")

	case errors.Is(actual, nonCoded):
		t.Errorf("actual error should not match non-coded error")

	case errors.Is(actual, ErrorTypeB):
		t.Errorf("actual error should not match ErrorTypeB")

	case errors.Is(actual, ErrorSubTypeA1):
		t.Errorf("actual error should not match ErrorSubTypeA1")

	case errors.Is(actual, ErrorSubTypeB1):
		t.Errorf("actual error should not match ErrorSubTypeB1")
	}
}

func TestCodeReserve(t *testing.T) {
	if e := tryReserve(ErrorTypeA); e == nil {
		t.Errorf("reserve non-category error should panic")
	}

	if e := tryReserve(ErrorCategoryTest); e != nil {
		t.Errorf("reserve category error for the first time should not panic")
	}

	if e := tryReserve(ErrorCategoryTest); e == nil {
		t.Errorf("reserve category error with same code should panic")
	}
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
	Test Err Hierarchy
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
const (
	_ = iota
	ErrorCodeB1_1 = ErrorSubTypeCodeB1 + iota
	ErrorCodeB1_2
)

// ErrorTypes, can be used in errors.Is
var (
	// masked errors for comparison
	ErrorCategoryTest = NewErrorCategory(Reserved, errors.New("error type: test"))
	ErrorTypeA        = NewErrorType(ErrorTypeCodeA, errors.New("error type: A"))
	ErrorTypeB        = NewErrorType(ErrorTypeCodeB, errors.New("error type: B"))

	ErrorSubTypeA1 = NewErrorSubType(ErrorSubTypeCodeA1, errors.New("error sub-type: A1"))
	ErrorSubTypeA2 = NewErrorSubType(ErrorSubTypeCodeA2, errors.New("error sub-type: A2"))
	ErrorSubTypeB1 = NewErrorSubType(ErrorSubTypeCodeB1, errors.New("error sub-type: B1"))
	ErrorSubTypeB2 = NewErrorSubType(ErrorSubTypeCodeB2, errors.New("error sub-type: B2"))

	// concrete errors
	ErrorA1_1 = NewCodedError(ErrorCodeA1_1, errors.New("error: A1-1"))
	ErrorA1_2 = NewCodedError(ErrorCodeA1_2, errors.New("error: A1-2"))
	ErrorA2   = NewCodedError(ErrorSubTypeCodeA2, errors.New("error: A2"))
	ErrorB1_1 = NewCodedError(ErrorCodeB1_1, errors.New("error: B1-1"))
	ErrorB1_2 = NewCodedError(ErrorCodeB1_2, errors.New("error: B1-2"))
	ErrorB2   = NewCodedError(ErrorSubTypeCodeB2, errors.New("error: B2"))
)