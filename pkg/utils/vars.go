package utils

import (
	"fmt"
	"github.com/google/uuid"
	"reflect"
)

type primitives interface {
	~bool |
		~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64 |
		~complex64 | ~complex128 |
		~string |
		~int | ~uint | ~uintptr
}

// MustSetIfNotNil takes "src" pointer (e.g. *bool) and set its dereference value to "dst" if not nil
// this function panic if:
// - "dst" and "src" are not pointer
// - "src" is not convertable to "dst"
// - "dst" not point to a settable value
func MustSetIfNotNil(dst interface{}, src interface{}) {
	dstV := reflect.ValueOf(dst)
	srcV := reflect.ValueOf(src)
	if srcV.IsNil() {
		return
	}
	dstEV := dstV.Elem()
	srcEV := srcV.Elem()
	dstEV.Set(srcEV.Convert(dstEV.Type()))
}

// SetIfNotNil is equivalent of MustSetIfNotNil, this function returns error instead of panic
func SetIfNotNil(dst interface{}, src interface{}) (err error) {
	defer func() {
		switch e := recover().(type) {
		case error:
			err = e
		default:
			err = fmt.Errorf("%v", e)
		}
	}()
	MustSetIfNotNil(dst, src)
	return
}

// MustSetIfNotZero takes "src" value (e.g. bool) and set its value to "dst" if not zero
// this function panic if:
// - "dst" is not pointer or not point to a settable value
// - "src" is not convertable to "dst"
func MustSetIfNotZero(dst interface{}, src interface{}) {
	dstV := reflect.ValueOf(dst)
	srcV := reflect.ValueOf(src)
	if srcV.IsZero() {
		return
	}
	dstEV := dstV.Elem()
	dstEV.Set(srcV.Convert(dstEV.Type()))
}

// SetIfNotZero is equivalent of MustSetIfNotZero, this function returns error instead of panic
func SetIfNotZero(dst interface{}, src interface{}) (err error) {
	defer func() {
		switch e := recover().(type) {
		case error:
			err = e
		default:
			err = fmt.Errorf("%v", e)
		}
	}()
	MustSetIfNotZero(dst, src)
	return
}

var (
	TRUE  = true
	FALSE = false
)

// FromPtr will take a pointer type and return its value if it is not nil. Otherwise,
// it will return the default value for that type. ex,
// 	var s *string
//  FromPtr(s) // results in ""
//  *s = "hello"
//  FromPtr(s) // results in "hello"
//  var b *bool
//  FromPtr(b) // results in false
// 	...
//  // Custom Types with underlying types of primitives
//  type String string
//  var s *String
//  FromPtr(s) // results in "" - but typed String
//  *s = String("hello")
//  FromPtr(s) // results in "hello" - but typed String
func FromPtr[T primitives](t *T) T {
	if t != nil {
		return *t
	}
	var defaultValueOfTypeT T
	return defaultValueOfTypeT
}

// ToPtr will return a pointer to any given input
// Example usage:
//
//  var stringPtr *string
//	stringPtr = ToPtr("hello world")
//
// 	// or some complex types
//  var funcPtr *[]func(arg *argType)
//  funcPtr = ToPtr([]func(arg *argType){})
func ToPtr[T any](t T) *T {
	return &t
}

// BoolPtr
// Deprecated: make use of ToPtr instead
func BoolPtr(v bool) *bool {
	if v {
		return &TRUE
	} else {
		return &FALSE
	}
}

// IntPtr
// Deprecated: make use of ToPtr instead
func IntPtr(v int) *int {
	return &v
}

// UIntPtr
// Deprecated: make use of ToPtr instead
func UIntPtr(v uint) *uint {
	return &v
}

// Float64Ptr
// Deprecated: make use of ToPtr instead
func Float64Ptr(v float64) *float64 {
	return &v
}

// StringPtr
// Deprecated: Make use of ToPtr instead
func StringPtr(v string) *string {
	return &v
}

// UuidPtr
// Deprecated: make use of ToPtr instead
func UuidPtr(v uuid.UUID) *uuid.UUID {
	return &v
}
