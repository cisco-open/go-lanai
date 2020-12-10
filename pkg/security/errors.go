package security

import (
	"bytes"
	"encoding/binary"
	"errors"
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
	ErrorTypeSecurity				 = newCodedError(Reserved, errors.New("error type: security"), ReservedMask)
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

type codedError struct {
	code int
	error
	mask int
}

func (e *codedError) Code() int {
	return e.code
}

func (e *codedError) CodeMask() int {
	return e.mask
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
func (e *codedError) MarshalBinary() (data []byte, err error) {
	errString := e.Error()

	buffer := bytes.NewBuffer([]byte{})
	err = binary.Write(buffer, binary.BigEndian, int64(e.code))
	err = binary.Write(buffer, binary.BigEndian, int64(e.mask))
	_, err = buffer.Write([]byte(errString))

	if err == nil {
		data = buffer.Bytes()
	}
	return
}

// encoding.BinaryUnmarshaler interface
func (e *codedError) UnmarshalBinary(data []byte) (err error) {
	buffer := bytes.NewBuffer(data)
	var code, mask int64
	err = binary.Read(buffer, binary.BigEndian, &code)
	err = binary.Read(buffer, binary.BigEndian, &mask)
	errString := buffer.String()
	if err != nil {
		return err
	}

	e.code = int(code)
	e.mask = int(mask)
	e.error = errors.New(errString)
	return
}


/************************
	Constructors
*************************/
func newCodedError(code int, e error, mask int) error {
	return &codedError{
		code: code,
		error: e,
		mask:   mask,
	}
}

func newErrorType(code int, e error) error {
	return newCodedError(code, e, errorTypeMask)
}

func newErrorSubType(code int, e error) error {
	return newCodedError(code, e, errorSubTypeMask)
}

// NewCodedError creates concrete error. it cannot be used as ErrorType or ErrorSubType comparison
func NewCodedError(code int, e error) error {
	return newCodedError(code, e, defaultErrorCodeMask)
}

/* InternalError family */
func NewInternalError(text string) error {
	return NewCodedError(ErrorTypeCodeInternal, errors.New(text))
}

/* AuthenticationError family */

func NewAuthenticationError(text string) error {
	return NewCodedError(ErrorTypeCodeAuthentication, errors.New(text))
}

func NewInternalAuthenticationError(text string) error {
	return NewCodedError(ErrorSubTypeCodeInternal, errors.New(text))
}

func NewAuthenticatorNotAvailableError(text string) error {
	return NewCodedError(ErrorCodeAuthenticatorNotAvailable, errors.New(text))
}

func NewUsernameNotFoundError(text string) error {
	return NewCodedError(ErrorCodeUsernameNotFound, errors.New(text))
}

func NewBadCredentialsError(text string) error {
	return NewCodedError(ErrorCodeBadCredentials, errors.New(text))
}

/* AccessControlError family */
func NewAccessControlError(text string) error {
	return NewCodedError(ErrorTypeCodeAccessControl, errors.New(text))
}

func NewAccessDeniedError(text string) error {
	return NewCodedError(ErrorSubTypeCodeAccessDenied, errors.New(text))
}

func NewInsufficientAuthError(text string) error {
	return NewCodedError(ErrorSubTypeCodeInsufficientAuth, errors.New(text))
}

func NewMissingCsrfTokenError(text string) error {
	return NewCodedError(ErrorCodeMissingCsrfToken, errors.New(text))
}

func NewInvalidCsrfTokenError(text string) error {
	return NewCodedError(ErrorCodeInvalidCsrfToken, errors.New(text))
}