package security

import "errors"


const (
	Reserved        		= 11 << 24

	errorTypeOffset         = 16
	errorTypeMask           = ^int(0) << errorTypeOffset

	errorSubTypeOffset      = 10
	errorSubTypeMask        = ^int(0) << errorSubTypeOffset

	defaultErrorCodeMask    = ^int(0)
)
// All "Type" values are used as mask
const (
	_ = iota
	ErrorTypeCodeAuthentication = Reserved + iota << errorTypeOffset
	ErrorTypeCodeAccessControl
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeAuthentication
const (
	_ = iota
	ErrorSubTypeCodeInternal             = ErrorTypeCodeAuthentication + iota << errorSubTypeOffset
	ErrorSubTypeCodeUsernamePasswordAuth
)

// ErrorSubTypeAuthInternal
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

// ErrorTypes, can be used in errors.Is
var (
	ErrorTypeAuthentication          = newErrorType(ErrorTypeCodeAuthentication, errors.New("error type: authentication"))
	ErrorTypeAccessControl           = newErrorType(ErrorTypeCodeAccessControl, errors.New("error type: access control"))

	ErrorSubTypeInternalError        = newErrorSubType(ErrorSubTypeCodeInternal, errors.New("error sub-type: internal"))
	ErrorSubTypeUsernamePasswordAuth = newErrorSubType(ErrorSubTypeCodeUsernamePasswordAuth, errors.New("error sub-type: internal"))
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

/************************
	Constructors
*************************/
func newCodedError(code int, e error, mask int) error {
	return &codedError{
		code: code,
		error: e,
		mask: mask,
	}
}

func newErrorType(code int, e error) error {
	return newCodedError(code, e, errorTypeMask)
}

func newErrorSubType(code int, e error) error {
	return newCodedError(code, e, errorSubTypeMask)
}

func NewAuthenticationError(text string) error {
	return newCodedError(ErrorTypeCodeAuthentication, errors.New(text), errorTypeMask)
}

func NewAccessControlError(text string) error {
	return newCodedError(ErrorTypeCodeAccessControl, errors.New(text), errorTypeMask)
}

func NewAuthenticatorNotAvailableError(text string) error {
	return newCodedError(ErrorCodeAuthenticatorNotAvailable, errors.New(text), defaultErrorCodeMask)
}

func NewUsernameNotFoundError(text string) error {
	return &codedError{
		code: ErrorCodeUsernameNotFound,
		error: errors.New(text),
	}
}

func NewBadCredentialsError(text string) error {
	return &codedError{
		code: ErrorCodeBadCredentials,
		error: errors.New(text),
	}
}

