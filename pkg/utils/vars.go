package utils

import (
	"fmt"
	"github.com/google/uuid"
	"reflect"
)


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

func BoolPtr(v bool) *bool {
	if v {
		return &TRUE
	} else {
		return &FALSE
	}
}

func IntPtr(v int) *int {
	return &v
}

func UIntPtr(v uint) *uint {
	return &v
}

func Float64Ptr(v float64) *float64 {
	return &v
}

func StringPtr(v string) *string {
	return &v
}

func UuidPtr(v uuid.UUID) *uuid.UUID {
	return &v
}
