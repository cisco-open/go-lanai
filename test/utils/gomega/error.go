package gomegautils

import (
	"errors"
	"fmt"
	errorutils "github.com/cisco-open/go-lanai/pkg/utils/error"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// IsError returns a types.GomegaMatcher that matches specified error. If the expected error is an errorutils.ErrorCoder
// the code and code mask is reported in the failure message
func IsError(expected error) types.GomegaMatcher {
	var code int64
	var errCoder errorutils.ErrorCoder
	if errors.As(expected, &errCoder) {
		code = errCoder.Code()
	}
	var mask int64
	var errCodeMask errorutils.ComparableErrorCoder
	if errors.As(expected, &errCodeMask) {
		mask = errCodeMask.CodeMask()
	}
	return &GomegaErrorMatcher{
		error: expected,
		code:  code,
		mask:  mask,
	}
}

// HaveErrorTypeCode returns a types.GomegaMatcher that matches errorutils.CodedError with given top-level type code
// the code and code mask is reported in the failure message
func HaveErrorTypeCode(typeCode int64) types.GomegaMatcher {
    return &GomegaErrorMatcher{
        error: errorutils.NewErrorType(typeCode, "error type"),
        code:  typeCode,
        mask:  errorutils.ErrorTypeMask,
    }
}

// HaveErrorSubTypeCode returns a types.GomegaMatcher that matches errorutils.CodedError with given sub-type code
// the code and code mask is reported in the failure message
func HaveErrorSubTypeCode(typeCode int64) types.GomegaMatcher {
    return &GomegaErrorMatcher{
        error: errorutils.NewErrorSubType(typeCode, "error sub-type"),
        code:  typeCode,
        mask:  errorutils.ErrorSubTypeMask,
    }
}

// HaveErrorCode returns a types.GomegaMatcher that matches errorutils.CodedError with given code
// the code and code mask is reported in the failure message
func HaveErrorCode(typeCode int64) types.GomegaMatcher {
	return &GomegaErrorMatcher{
		error: errorutils.NewCodedError(typeCode, "coded error"),
		code:  typeCode,
		mask:  errorutils.DefaultErrorCodeMask,
	}
}

// GomegaErrorMatcher implements types.GomegaMatcher for error type
type GomegaErrorMatcher struct {
	error error
	code  int64
	mask  int64
}

func (m *GomegaErrorMatcher) Match(actual interface{}) (success bool, err error) {
	actualErr, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf(`%T is not an error`, actual)
	}
	return errors.Is(actualErr, m.error), nil
}

func (m *GomegaErrorMatcher) FailureMessage(actual interface{}) (message string) {
	var msg string
	if m.code != 0 {
		msg = fmt.Sprintf("to be an error with type [%T], code [%#016x] and mask [%#016x]", m.error, uint64(m.code), uint64(m.mask))
	} else {
		msg = fmt.Sprintf(`to equals to error [%T] - %v`, m.error, m.error)
	}
	return fmt.Sprintf("Expected\n%s\n%s\n", m.formatActual(actual), msg)
}

func (m *GomegaErrorMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	var msg string
	if m.code != 0 {
		msg = fmt.Sprintf("not to be an error with type [%T], code [%#016x] and mask [%#016x]", m.error, uint64(m.code), uint64(m.mask))
	} else {
		msg = fmt.Sprintf(`to not equal to error [%T] - %v`, m.error, m.error)
	}
	return fmt.Sprintf("Expected\n%s\n%s\n", m.formatActual(actual), msg)
}

func (m *GomegaErrorMatcher) formatActual(actual interface{}) interface{} {
	desc := format.Object(actual, 1)
	actualErr, ok := actual.(error)
	if !ok {
		return desc
	}
	var coder errorutils.ErrorCoder
	if errors.As(actualErr, &coder) {
		desc = desc + fmt.Sprintf(" <Code %#016x>", uint64(coder.Code()))
	}
	return desc
}
