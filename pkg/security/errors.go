package security

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// security reserved
	ReservedOffset			= 24
	Reserved        		= 11 << ReservedOffset
	ReservedMask			= ^int(0) << ReservedOffset

	// error type bits
	errorTypeOffset         = 16
	errorTypeMask           = ^int(0) << errorTypeOffset

	// error sub type bits
	errorSubTypeOffset      = 10
	errorSubTypeMask        = ^int(0) << errorSubTypeOffset

	defaultErrorCodeMask    = ^int(0)
)
// All "Type" values are used as mask
const (
	_ = iota
	ErrorTypeCodeAuthentication = Reserved + iota << errorTypeOffset
	ErrorTypeCodeAccessControl
	ErrorTypeCodeInternal
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeAuthentication
const (
	_ = iota
	ErrorSubTypeCodeInternal             = ErrorTypeCodeAuthentication + iota << errorSubTypeOffset
	ErrorSubTypeCodeUsernamePasswordAuth
)

// ErrorSubTypeCodeInternal
const (
	_ = iota
	ErrorCodeAuthenticatorNotAvailable = ErrorSubTypeCodeInternal + iota
)

// ErrorSubTypeCodeUsernamePasswordAuth
const (
	_ = iota
	ErrorCodeUsernameNotFound = ErrorSubTypeCodeUsernamePasswordAuth + iota
	ErrorCodeBadCredentials
	ErrorCodeCredentialsExpired
	ErrorCodeMaxAttemptsReached
	ErrorCodeAccountStatus
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeAccessControl
const (
	_ = iota
	ErrorSubTypeCodeAccessDenied = ErrorTypeCodeAccessControl + iota << errorSubTypeOffset
	ErrorSubTypeCodeInsufficientAuth
	ErrorSubTypeCodeCsrf
)

const (
	_ = iota
	ErrorCodeMissingCsrfToken = ErrorSubTypeCodeCsrf + iota
	ErrorCodeInvalidCsrfToken
)

// ErrorTypes, can be used in errors.Is
var (
	ErrorTypeSecurity				 = newCodedError(Reserved, errors.New("error type: security"), ReservedMask, nil)
	ErrorTypeAuthentication          = newErrorType(ErrorTypeCodeAuthentication, errors.New("error type: authentication"))
	ErrorTypeAccessControl           = newErrorType(ErrorTypeCodeAccessControl, errors.New("error type: access control"))
	ErrorTypeInternal                = newErrorType(ErrorTypeCodeInternal, errors.New("error type: internal"))

	ErrorSubTypeInternalError        = newErrorSubType(ErrorSubTypeCodeInternal, errors.New("error sub-type: internal"))
	ErrorSubTypeUsernamePasswordAuth = newErrorSubType(ErrorSubTypeCodeUsernamePasswordAuth, errors.New("error sub-type: internal"))

	ErrorSubTypeAccessDenied         = newErrorSubType(ErrorSubTypeCodeAccessDenied, errors.New("error sub-type: access denied"))
	ErrorSubTypeInsufficientAuth     = newErrorSubType(ErrorSubTypeCodeInsufficientAuth, errors.New("error sub-type: insufficient auth"))
	ErrorSubTypeCsrf 				 = newErrorSubType(ErrorSubTypeCodeCsrf, errors.New("error sub-type: csrf"))


)

type ErrorCoder interface {
	Code() int
}

type ComparableErrorCoder interface {
	CodeMask() int
}

type NestedError interface {
	Cause() error
}

// codedError implements Code, CodeMask, NestedError
type codedError struct {
	code int
	error
	mask int
	cause error
}

func (e *codedError) Code() int {
	return e.code
}

func (e *codedError) CodeMask() int {
	return e.mask
}

func (e *codedError) Cause() error {
	return e.cause
}

// Is return true if
//	1. target has same code, OR
//  2. target is a type/sub-type error and the receiver error is in same type/sub-type
func (e *codedError) Is(target error) bool {
	compare := e.code
	if masker, ok := target.(ComparableErrorCoder); ok {
		compare = e.code & masker.CodeMask()
	}

	coder, ok := target.(ErrorCoder)
	return  ok && compare == coder.Code()
}

// encoding.BinaryMarshaler interface
// code, mask, error.Error() are written into byte array in the mentioned order
// code and mask are written as 64 bits with binary.BigEndian
// Note: currently we don't serialize Cause() to avoid cyclic reference
func (e *codedError) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	if err := binary.Write(buffer, binary.BigEndian, int64(e.code)); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, int64(e.mask)); err != nil {
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
func (e *codedError) UnmarshalBinary(data []byte) (error) {
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

	e.code = int(code)
	e.mask = int(mask)
	e.error = errors.New(string(errBytes[:len(errBytes) - 1]))
	return nil
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
func newCodedError(code int, e error, mask int, cause error) error {
	return &codedError{
		code: code,
		error: e,
		mask:   mask,
		cause: cause,
	}
}

func newErrorType(code int, e error) error {
	return newCodedError(code, e, errorTypeMask, nil)
}

func newErrorSubType(code int, e error) error {
	return newCodedError(code, e, errorSubTypeMask, nil)
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
func NewCodedError(code int, e interface{}, causes...interface{}) error {
	err := construct(e)
	if len(causes) == 0 {
		return newCodedError(code, err, defaultErrorCodeMask, nil)
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
	return newCodedError(code, err, defaultErrorCodeMask, &cause)
}

/* InternalError family */
func NewInternalError(text string, causes...interface{}) error {
	return NewCodedError(ErrorTypeCodeInternal, errors.New(text), causes...)
}

/* AuthenticationError family */
func NewAuthenticationError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorTypeCodeAuthentication, value, causes...)
}

func NewInternalAuthenticationError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeInternal, value, causes...)
}

func NewAuthenticatorNotAvailableError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeAuthenticatorNotAvailable, value, causes...)
}

func NewUsernameNotFoundError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeUsernameNotFound, value, causes...)
}

func NewBadCredentialsError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeBadCredentials, value, causes...)
}

func NewCredentialsExpiredError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeCredentialsExpired, value, causes...)
}

func NewMaxAttemptsReachedError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeMaxAttemptsReached, value, causes...)
}

func NewAccountStatusError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeAccountStatus, value, causes...)
}

/* AccessControlError family */
func NewAccessControlError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorTypeCodeAccessControl, value, causes...)
}

func NewAccessDeniedError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeAccessDenied, value, causes...)
}

func NewInsufficientAuthError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeInsufficientAuth, value, causes...)
}

func NewMissingCsrfTokenError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeMissingCsrfToken, value, causes...)
}

func NewInvalidCsrfTokenError(value interface{}, causes...interface{}) error {
	return NewCodedError(ErrorCodeInvalidCsrfToken, value, causes...)
}