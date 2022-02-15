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
const (
	errorMsgExpectFunc       = "expecting a function"
	errorMsgInputParams      = "function should have one or two input parameters, where the first is context.Context and the second is a struct or pointer to struct"
	errorMsgOutputParams     = "function should have at least two output parameters, where the the last is error"
	errorMsgInvalidSignature = "unable to find request or response type"
)

var (
	specialTypeContext        = reflect.TypeOf((*context.Context)(nil)).Elem()
	specialTypeHttpRequestPtr = reflect.TypeOf(&http.Request{})
	specialTypeInt            = reflect.TypeOf(int(0))
	specialTypeHttpHeader     = reflect.TypeOf((*http.Header)(nil)).Elem()
	specialTypeError          = reflect.TypeOf((*error)(nil)).Elem()
)

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
	count    int
	sc       param
	header   param
	response param
	error    param
}

// in parameters
type mvcIn struct {
	count   int
	context param
	request param
}

type mvcMetadata struct {
	function *reflect.Value
	request  reflect.Type
	response reflect.Type
	in       mvcIn
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
	for i := t.NumIn() - 1; i >= 0; i-- {
		switch it := t.In(i); {
		case it.ConvertibleTo(specialTypeContext):
			meta.in.context = param{i, it}
		case !meta.in.request.isValid() && isSupportedRequestType(it):
			meta.in.request = param{i, it}
			meta.request = it
		default:
			panic(&errorInvalidMvcHandlerFunc{
				reason: errors.New(fmt.Sprintf("unknown input parameters at index %v", i)),
				target: &f,
			})
		}
		meta.in.count++
	}

	// parse output params
	for i := t.NumOut() - 1; i >= 0; i-- {
		switch ot := t.Out(i); {
		case ot.ConvertibleTo(specialTypeInt):
			meta.out.sc = param{i, ot}
		case ot.ConvertibleTo(specialTypeHttpHeader):
			meta.out.header = param{i, ot}
		case ot.ConvertibleTo(specialTypeError):
			meta.out.error = param{i, ot}
		case !meta.out.response.isValid() && isSupportedResponseType(ot):
			// we allow interface and map as response
			meta.out.response = param{i, ot}
			meta.response = ot
		default:
			panic(&errorInvalidMvcHandlerFunc{
				reason: errors.New(fmt.Sprintf("unknown return parameters at index %v", i)),
				target: &f,
			})
		}
		meta.out.count++
	}

	if meta.response == nil || meta.in.count < 1 || meta.out.count < 2 || meta.in.count > 1 && meta.request == nil {
		panic(&errorInvalidMvcHandlerFunc{
			reason: errors.New(errorMsgInvalidSignature),
			target: &f,
		})
	}

	return &meta
}

func validateFunc(f *reflect.Value, validator MvcHandlerFuncValidator) (err error) {
	// For now, we check function signature at runtime.
	// I wish there is a way to check it at compile-time that I didn't know of
	t := f.Type()
	switch {
	case f.Kind() != reflect.Func:
		return &errorInvalidMvcHandlerFunc{
			reason: errors.New(errorMsgExpectFunc),
			target: f,
		}
	// In params validation
	case t.NumIn() < 1 || t.NumIn() > 2:
		fallthrough
	case !t.In(0).ConvertibleTo(specialTypeContext):
		fallthrough
	case t.NumIn() == 2 && !isSupportedRequestType(t.In(t.NumIn()-1)):
		return &errorInvalidMvcHandlerFunc{
			reason: errors.New(errorMsgInputParams),
			target: f,
		}

	// Out params validation
	case t.NumOut() < 2:
		fallthrough
	case !t.Out(t.NumOut() - 1).ConvertibleTo(specialTypeError):
		return &errorInvalidMvcHandlerFunc{
			reason: errors.New(errorMsgOutputParams),
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

// isHttpRequestPtr returns true if given type is *http.Request
func isHttpRequestPtr(t reflect.Type) bool {
	return t == specialTypeHttpRequestPtr
}

func isSupportedRequestType(t reflect.Type) bool {
	return isStructOrPtrToStruct(t)
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
	default:
		return false
	}
}
