package security

import "errors"


const (
	errorTypeOffset         = 16
	errorTypeMask           = 0xffffffff << errorTypeOffset

	errorSubTypeOffset      = 8
	errorSubTypeMask        = 0xffffffff << errorSubTypeOffset
)
// All "Type" values are used as mask
const (
	_ = iota
	ErrorTypeCodeAuthentication = iota << errorTypeOffset
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

type ErrorCodeMasker interface {
	Mask() int
}

type ErrorCoder interface {
	Code() int
}

type codedError struct {
	code int
	error
}

func (e *codedError) Code() int {
	return e.code
}

type parentCodedError struct {
	codedError
	mask int
}

func (e *parentCodedError) Mask() int {
	return e.mask
}

func newErrorType(code int, e error) *parentCodedError {
	return &parentCodedError{
		codedError: codedError{
			code:  code,
			error: e,
		},
		mask: errorTypeMask,
	}
}

func newErrorSubType(code int, e error) *parentCodedError {
	return &parentCodedError{
		codedError: codedError{
			code:  code,
			error: e,
		},
		mask: errorSubTypeMask,
	}
}

// Is return true if
//	1. target has same code, OR
//  2. target is a type/sub-type error and the receiver error is in same type/sub-type
func (e *codedError) Is(target error) bool {
	compare := e.code
	if masker, ok := target.(ErrorCodeMasker); ok {
		compare = e.code & masker.Mask()
	}

	coder, ok := target.(ErrorCoder)
	return  ok && compare == coder.Code()
}

/************************
	Constructors
*************************/
func NewAuthenticatorNotAvailableError(text string) error {
	return &codedError{
		code: ErrorCodeAuthenticatorNotAvailable,
		error: errors.New(text),
	}
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

