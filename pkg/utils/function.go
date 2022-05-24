package utils

import (
	"fmt"
	"reflect"
)

const (
	errTmplIncorrectSignature = `incorrect signature [%T] for RecoverableFunc, check the document for usage`
)

var (
	typeError = reflect.TypeOf((*error)(nil)).Elem()
)

// RecoverableFunc wrap a panicing function with following signature
// - func()
// - func() error
// into a func() error, where the recovered value is converted to error
// This function panics if the given function has incorrect signature
func RecoverableFunc(panicingFunc interface{}) func() error {
	rv := reflect.ValueOf(panicingFunc)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		panic("unable to recover a non-function type")
	}
	if rt.NumIn() != 0 {
		panic(fmt.Sprintf(errTmplIncorrectSignature, panicingFunc))
	}

	var fn func() error
	switch rt.NumOut() {
	case 0:
		fn = func() error {
			rv.Call(nil)
			return nil
		}
	case 1:
		if !rt.Out(0).AssignableTo(typeError) {
			panic(fmt.Sprintf(errTmplIncorrectSignature, panicingFunc))
		}
		fn = func() error {
			ret := rv.Call(nil)
			if ret[0].IsNil() {
				return nil
			}
			return ret[0].Interface().(error)
		}
	default:
		panic(fmt.Sprintf(errTmplIncorrectSignature, panicingFunc))
	}

	return func() (err error) {
		defer func() {
			switch v := recover().(type) {
			case error:
				err = v
			case nil:
			default:
				err = fmt.Errorf("unable to run gateway: %v", v)
			}
		}()

		return fn()
	}
}
