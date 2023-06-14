package opa

import "fmt"

var (
	InternalError           = NewError("internal error")
	AccessDeniedError       = NewError("Access Denied")
	QueriesNotResolvedError = NewError(`OPA cannot resolve partial queries`)
)

var errorCode int

type Error struct {
	code int
	msg  string
}

func (e Error) Error() string {
	return e.msg
}

func (e Error) Is(err error) bool {
	v, ok := err.(Error)
	return ok && v.code == e.code
}

func (e Error) WithMessage(tmpl string, args ...interface{}) Error {
	return Error{
		code: e.code,
		msg:  fmt.Sprintf(tmpl, args...),
	}
}

func NewError(tmpl string, args ...interface{}) Error {
	errorCode++
	return Error{
		code: errorCode,
		msg:  fmt.Sprintf(tmpl, args...),
	}
}
