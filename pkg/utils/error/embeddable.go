package errorutils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// CodedError implements error, Code, CodeMask, NestedError, ComparableError, Unwrapper
// encoding.TextMarshaler, json.Marshaler, encoding.BinaryMarshaler, encoding.BinaryUnmarshaler
type CodedError struct {
	ErrMsg  string
	ErrCode int64
	ErrMask int64
	Nested  error
}

func (e CodedError) Error() string {
	return e.ErrMsg
}

func (e CodedError) Code() int64 {
	return e.ErrCode
}

func (e CodedError) CodeMask() int64 {
	return e.ErrMask
}

func (e CodedError) Cause() error {
	return e.Nested
}

func (e CodedError) RootCause() error {
	if nested, ok := e.Nested.(NestedError); ok {
		return nested.RootCause()
	}
	return e.Nested
}

// WithMessage make a concrete error with given error message
func (e CodedError) WithMessage(msg string, args ...interface{}) *CodedError {
	return NewCodedError(e.ErrCode, fmt.Errorf(msg, args...))
}

// WithCause make a concrete error with given cause and error message
func (e CodedError) WithCause(cause error, msg string, args ...interface{}) *CodedError {
	return NewCodedError(e.ErrCode, fmt.Errorf(msg, args...), cause)
}

// MarshalText implements encoding.TextMarshaler
func (e CodedError) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface
// ErrCode, ErrMask, error.Error() are written into byte array in the mentioned order
// ErrCode and ErrMask are written as 64 bits with binary.BigEndian
// Note: currently we don't serialize Cause() to avoid cyclic reference
func (e CodedError) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	if err := binary.Write(buffer, binary.BigEndian, e.ErrCode); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, e.ErrMask); err != nil {
		return nil, err
	}
	if _, err := buffer.WriteString(e.Error()); err != nil {
		return nil, err
	}
	if err := buffer.WriteByte(byte(0)); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface
func (e *CodedError) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	var code, mask int64
	if err := binary.Read(buffer, binary.BigEndian, &code); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.BigEndian, &mask); err != nil {
		return err
	}

	errBytes, err := buffer.ReadBytes(byte(0))
	if err != nil {
		return err
	}

	e.ErrCode = code
	e.ErrMask = mask
	e.ErrMsg = string(errBytes[:len(errBytes)-1])
	return nil
}

// Is return true if
//	1. target has same ErrCode, OR
//  2. target is a type/sub-type error and the receiver error is in same type/sub-type
func (e CodedError) Is(target error) bool {
	compare := e.ErrCode
	if masker, ok := target.(ComparableErrorCoder); ok {
		compare = e.ErrCode & masker.CodeMask()
	}

	if coder, ok := target.(ErrorCoder); ok && compare == coder.Code() {
		return true
	}

	return false
}

// nestedError implements error, NestedError, encoding.BinaryMarshaler, encoding.BinaryUnmarshaler
type nestedError struct {
	error
	nested error
}

func (e nestedError) Is(target error) bool {
	return errors.Is(e.error, target) || e.nested != nil && errors.Is(e.nested, target)
}

func (e nestedError) Cause() error {
	return e.nested
}

func (e nestedError) RootCause() error {
	for root := e.nested; root != nil; {
		if nested, ok := root.(NestedError); ok {
			root = nested.Cause()
		} else {
			return root
		}
	}
	return e.error
}

// MarshalBinary implements encoding.BinaryMarshaler interface
// error.Error(), is written into byte array in the mentioned order
// Note: currently we don't serialize Cause() to avoid cyclic reference
func (e nestedError) MarshalBinary() ([]byte, error) {

	buffer := bytes.NewBuffer([]byte{})
	if _, err := buffer.WriteString(e.Error()); err != nil {
		return nil, err
	}
	if err := buffer.WriteByte(byte(0)); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface
func (e *nestedError) UnmarshalBinary(data []byte) (err error) {
	buffer := bytes.NewBuffer(data)
	errBytes, err := buffer.ReadString(byte(0))
	if err != nil {
		return err
	}
	e.error = errors.New(errBytes[:len(errBytes)-1])
	return
}

/************************
	Constructors
*************************/
func newCodedError(code int64, msg string, mask int64, cause error) *CodedError {
	return &CodedError{
		ErrMsg:  msg,
		ErrCode: code,
		ErrMask: mask,
		Nested:  cause,
	}
}

func NewErrorCategory(code int64, e interface{}) *CodedError {
	code = code & ReservedMask
	return newCodedError(code, fmt.Sprintf("%v", e), ReservedMask, nil)
}

func NewErrorType(code int64, e interface{}) *CodedError {
	code = code & ErrorTypeMask
	return newCodedError(code, fmt.Sprintf("%v", e), ErrorTypeMask, nil)
}

func NewErrorSubType(code int64, e interface{}) *CodedError {
	code = code & ErrorSubTypeMask
	return newCodedError(code, fmt.Sprintf("%v", e), ErrorSubTypeMask, nil)
}

// construct error from supported item: string, error, fmt.Stringer
func construct(e interface{}) error {
	var err error
	switch e.(type) {
	case error:
		err = e.(error)
	case fmt.Stringer:
		err = errors.New(e.(fmt.Stringer).String())
	case string:
		err = errors.New(e.(string))
	default:
		err = fmt.Errorf("%v", e)
	}
	return err
}

// NewCodedError creates concrete error. it cannot be used as ErrorType or ErrorSubType comparison
// supported item are string, error, fmt.Stringer
func NewCodedError(code int64, e interface{}, causes ...interface{}) *CodedError {
	// special case, cause is not specified, use "e" as cause
	if len(causes) == 0 {
		causes = []interface{}{e}
	}

	// chain causes
	var cause error
	for i := len(causes) - 1; i >= 0; i-- {
		current := construct(causes[i])
		if cause == nil {
			cause = current
		} else {
			cause = &nestedError{
				error:  cause,
				nested: current,
			}
		}
	}

	return newCodedError(code, fmt.Sprintf("%v", e), DefaultErrorCodeMask, cause)
}
