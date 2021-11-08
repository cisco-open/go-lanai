package errorutils

import (
	"encoding/gob"
	"fmt"
	"reflect"
)

var reserved = map[int64]string{}

func init() {
	gob.Register((*CodedError)(nil))
	gob.Register((*nestedError)(nil))
}

// Reserve is used for error hierarchy defining packages to validate and reserve its error code range
// it's usually called during init()
// this funciton does following things:
// 	1. validate given err implements error, ErrorCoder and ComparableErrorCoder
//  2. the mask equals ReservedMask (a category error created via NewErrorCategory)
//  3. bits lower than ReservedOffset of the given error's code are all 0
//  4. if the code is available (not registered by other packages)
//  5. try to register the error's implementation with gob
func Reserve(err interface{}) {
	switch err.(type) {
	case error, ErrorCoder, ComparableErrorCoder:
	default:
		panic(fmt.Errorf("cannot reserve error category %T", err))
	}

	if masker := err.(ComparableErrorCoder); masker.CodeMask() != ReservedMask {
		panic(fmt.Errorf("cannot reserve error category with code mask %x", masker.CodeMask()))
	}

	coder := err.(ErrorCoder)
	if coder.Code() & ^ReservedMask != 0 {
		panic(fmt.Errorf("cannot reserve error category with code %x, it's not a category level codes", coder.Code() ))
	}

	if pkg, ok := reserved[coder.Code()]; ok {
		panic(fmt.Errorf("error category with code %x is already registered by ", pkg ))
	}

	// try reserve
	gob.Register(err)
	reserved[coder.Code()] = reflect.TypeOf(err).PkgPath()
}

type ErrorCoder interface {
	Code() int64
}

type ComparableErrorCoder interface {
	CodeMask() int64
}

type NestedError interface {
	// Cause returns directly nested error
	Cause() error

	// RootCause returns the root cause of error, equivalent to calling Cause repeatedly
	RootCause() error
}

type ComparableError interface {
	Is(target error) bool
}

type Unwrapper interface {
	Unwrap() error
}

// Hasher is an interface for error implementations that not naturally hashable to be used as a map key
// e.g. if a struct containing Slice, Map, Array are not hashable. Therefore, use those errors as map key would panic
// For implementations that need use error as map key (e.g. map[error]interface{}) should use Hasher.Hash as key
// Note: CodedError doesn't implement this interface because itself is hashable
type Hasher interface {
	Hash() error
}