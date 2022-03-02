package errorutils

import (
	"errors"
	"fmt"
)

// WrappedError is an embeddable struct that
// provide a convenient way to implement concrete error within certain error hierarchy without error code.
// This error implementation contains 3 components:
// - ErrIs is an anchor error used for comparison. Used for Is
// - Type is the parent error indicating its type, CodedError. Used for Unwrap
// - ErrMsg is the error's actual string value. Used for Error
type WrappedError struct {
	ErrIs error
	Type  *CodedError
	ErrMsg string
}

func (e WrappedError) Error() string {
	return e.ErrMsg
}

func (e WrappedError) Is(target error) bool {
	//nolint:errorlint // type assert is intentional
	switch t := target.(type) {
	case compareTargeter:
		wrappedE := t.target()
		return e == wrappedE || errors.Is(e.ErrIs, wrappedE.ErrIs) && errors.Is(e.Type, wrappedE.Type)
	default:
		return false
	}
}

// Unwrap returns type error,
// which makes sure that errors.Is(e, errorType) returns true when errors.Is(e.Type, errorType) is true
func (e WrappedError) Unwrap() error {
	return e.Type
}

// MarshalText implements encoding.TextMarshaler
func (e WrappedError) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}

func (e WrappedError) WithMessage(msg string, args ...interface{}) WrappedError {
	return WrappedError{
		ErrIs:  e.ErrIs,
		Type:   e.Type,
		ErrMsg: fmt.Sprintf(msg, args...),
	}
}

// compareTarget is an internal interface that makes Embedding implementation can be compared with another WrappedError
// overriding Is
type compareTargeter interface {
	target() WrappedError
}

func (e WrappedError) target() WrappedError {
	return e
}