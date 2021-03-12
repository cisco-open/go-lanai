package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

/*****************************
	Func Metadata
******************************/
// MvcHandlerFuncValidator validate MvcHandlerFunc signature
type MvcHandlerFuncValidator func(f *reflect.Value) error

type MvcHandlerFuncReturnMapper func(*[]reflect.Value) (interface{}, error)

type param struct {
	i int
	t reflect.Type
}

func (p param) isValid() bool {
	return p.i >= 0 && p.t != nil
}

// out parameters
type mvcOut struct {
	count int
	sc param
	header param
	response param
	error param
}

// in parameters
type mvcIn struct {
	count int
	context param
	request param
}

type mvcMetadata struct {
	function *reflect.Value
	request  reflect.Type
	response reflect.Type
	in 		 mvcIn
	out      mvcOut
}

// MakeFuncMetadata uses reflect to analyze the given rest function and create a endpointFuncMetadata
// this function panic if given function have incorrect signature
// Caller can provide an optional validator to further validate function signature on top of default validation
func MakeFuncMetadata(endpointFunc MvcHandlerFunc, validator MvcHandlerFuncValidator) *mvcMetadata {
	f := reflect.ValueOf(endpointFunc)
	err := validateFunc(&f, validator)
	if err != nil {
		//fatal error
		panic(err)
	}

	t := f.Type()
	unknown := param{-1, nil}
	meta := mvcMetadata{
		function: &f,
		in: mvcIn{
			context: unknown, request: unknown,
		},
		out: mvcOut{
			sc: unknown, header: unknown,
			response: unknown, error: unknown,
		},
	}

	// parse input params
	for i := t.NumIn() - 1; i >=0; i-- {
		switch it := t.In(i); {
		case it.ConvertibleTo(reflect.TypeOf((*context.Context)(nil)).Elem()):
			meta.in.context = param{i, it}
		case !meta.in.request.isValid() && isStructOrPtrToStruct(it):
			meta.in.request = param{i, it}
			meta.request = it
		default:
			panic(&errorInvalidMvcHandlerFunc{
				reason: errors.New(fmt.Sprintf("unknown input parameters at index %v", i)),
				target: &f,
			})
		}
		meta.in.count ++
	}

	// parse output params
	for i := t.NumOut() -1; i >=0; i-- {
		switch ot := t.Out(i); {
		case ot.ConvertibleTo(reflect.TypeOf(0)):
			meta.out.sc = param {i, ot}
		case ot.ConvertibleTo(reflect.TypeOf((*http.Header)(nil)).Elem()):
			meta.out.header = param {i, ot}
		case ot.ConvertibleTo(reflect.TypeOf((*error)(nil)).Elem()):
			meta.out.error = param {i, ot}
		case !meta.out.response.isValid() && isSupportedResponseType(ot):
			// we allow interface and map as response
			meta.out.response = param {i, ot}
			meta.response = ot
		default:
			panic(&errorInvalidMvcHandlerFunc{
				reason: errors.New(fmt.Sprintf("unknown return parameters at index %v", i)),
				target: &f,
			})
		}
		meta.out.count ++
	}

	if meta.request == nil || meta.response == nil || meta.in.count < 2 || meta.out.count < 2 {
		panic(&errorInvalidMvcHandlerFunc{
			reason: errors.New("unable to find request or response type"),
			target: &f,
		})
	}

	return &meta
}

func validateFunc(f *reflect.Value, validator MvcHandlerFuncValidator) (err error) {
	// For now, we check function signature at runtime.
	// I wish there is a way to check it at compile-time that I didn't know of
	t := f.Type()
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	switch {
	case f.Kind() != reflect.Func:
		return &errorInvalidMvcHandlerFunc{
			reason: errors.New("expecting a function"),
			target: f,
		}
	// In params validation
	case t.NumIn() < 2:
		fallthrough
	case !t.In(0).ConvertibleTo(ctxType):
		fallthrough
	case !isStructOrPtrToStruct(t.In(t.NumIn() - 1)):
		return &errorInvalidMvcHandlerFunc{
			reason: errors.New("function should have at least two input parameters, where the first is context.Context and the last is a struct or pointer to struct"),
			target: f,
		}
	// Out params validation
	case t.NumOut() < 2:
		fallthrough
	case !t.Out(t.NumOut() - 1).ConvertibleTo(errorType):
		return &errorInvalidMvcHandlerFunc{
			reason: errors.New("function should have at least two output parameters, where the the last is error"),
			target: f,
		}
	}

	if validator != nil {
		return validator(f)
	}
	return nil
}

func isStructOrPtrToStruct(t reflect.Type) (ret bool) {
	ret = t.Kind() == reflect.Struct
	ret = ret || t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
	return
}

func isSupportedResponseType(t reflect.Type) bool {
	if isStructOrPtrToStruct(t) {
		return true
	}
	switch t.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.String:
		return true
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return t.Elem().Kind() == reflect.Uint8
	}
	return false
}