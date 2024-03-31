package dsync

import (
	"errors"
	"fmt"
)

var errorTypeCounter int

type comparableError struct {
	typ   int
	msg   string
	cause error
}

func newError(tmpl string, args...interface{}) comparableError {
	errorTypeCounter ++
	return comparableError {
		typ: errorTypeCounter,
		msg: fmt.Sprintf(tmpl, args...),
	}
}

func (w comparableError) Error() string {
	return w.msg
}

func (w comparableError) Is(target error) bool {
	var comparableTarget comparableError
	if errors.As(target, &comparableTarget) {
		return w.typ == comparableTarget.typ
	}
	return errors.Is(w.cause, target)
}

func (w comparableError) Unwrap() error {
	return w.cause
}

func (w comparableError) WithMessage(tmpl string, args ...interface{}) comparableError {
	return comparableError{
		typ:   w.typ,
		msg:   fmt.Sprintf(tmpl, args...),
		cause: w.cause,
	}
}

func (w comparableError) WithCause(err error) comparableError {
	return comparableError{
		typ:   w.typ,
		msg:   fmt.Sprintf(`%s: %v`, w.msg, err),
		cause: err,
	}
}
