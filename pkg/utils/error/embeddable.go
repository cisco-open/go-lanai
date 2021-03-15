package errorutils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Err error

// CodedError implements error, Code, CodeMask, NestedError
// encoding.TextMarshaler, json.Marshaler, encoding.BinaryMarshaler, encoding.BinaryUnmarshaler
type CodedError struct {
	Err
	ErrCode int64
	ErrMask int64
	Nested  error
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

// encoding.TextMarshaler
func (e CodedError) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}

// encoding.BinaryMarshaler interface
// ErrCode, ErrMask, error.Error() are written into byte array in the mentioned order
// ErrCode and ErrMask are written as 64 bits with binary.BigEndian
// Note: currently we don't serialize Cause() to avoid cyclic reference
func (e *CodedError) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	if err := binary.Write(buffer, binary.BigEndian, int64(e.ErrCode)); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, int64(e.ErrMask)); err != nil {
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

// encoding.BinaryUnmarshaler interface
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
	e.Err = errors.New(string(errBytes[:len(errBytes) - 1]))
	return nil
}

// Is return true if
//	1. target has same ErrCode, OR
//  2. target is a type/sub-type error and the receiver error is in same type/sub-type
func (e *CodedError) Is(target error) bool {
	compare := e.ErrCode
	if masker, ok := target.(ComparableErrorCoder); ok {
		compare = e.ErrCode & masker.CodeMask()
	}

	coder, ok := target.(ErrorCoder)
	return  ok && compare == coder.Code()
}

// nestedError implements NestedError
type nestedError struct {
	error
	cause error
}

func (e *nestedError) Cause() error {
	return e.cause
}

// encoding.BinaryMarshaler interface
// error.Error(), is written into byte array in the mentioned order
// Note: currently we don't serialize Cause() to avoid cyclic reference
func (e *nestedError) MarshalBinary() ([]byte, error) {

	buffer := bytes.NewBuffer([]byte{})
	if _, err := buffer.WriteString(e.Error()); err != nil {
		return nil, err
	}
	if err := buffer.WriteByte(byte(0)); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// encoding.BinaryUnmarshaler interface
func (e *nestedError) UnmarshalBinary(data []byte) (err error) {
	buffer := bytes.NewBuffer(data)
	errBytes, err := buffer.ReadString(byte(0))
	if err != nil {
		return err
	}
	e.error = errors.New(string(errBytes[:len(errBytes) - 1]))
	return
}

/************************
	Constructors
*************************/
func newCodedError(code int64, e error, mask int64, cause error) *CodedError {
	return &CodedError{
		Err:     e,
		ErrCode: code,
		ErrMask: mask,
		Nested:  cause,
	}
}

func NewErrorCategory(code int64, e error) *CodedError {
	code = code & ReservedMask
	return newCodedError(code, e, ReservedMask, nil)
}

func NewErrorType(code int64, e error) error {
	code = code & ErrorTypeMask
	return newCodedError(code, e, ErrorTypeMask, nil)
}

func NewErrorSubType(code int64, e error) error {
	code = code & ErrorSubTypeMask
	return newCodedError(code, e, ErrorSubTypeMask, nil)
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
func NewCodedError(code int64, e interface{}, causes...interface{}) *CodedError {
	err := construct(e)
	if len(causes) == 0 {
		return newCodedError(code, err, DefaultErrorCodeMask, nil)
	}

	// chain causes
	cause := nestedError{
		error: construct(causes[0]),
	}
	nested := &cause
	for i, c := range causes[1:] {
		if i < len(causes) - 1 {
			nested.cause = &nestedError{
				error: construct(c),
			}
			nested = nested.cause.(*nestedError)
		} else {
			nested.cause = construct(c)
		}
	}
	return newCodedError(code, err, DefaultErrorCodeMask, &cause)
}