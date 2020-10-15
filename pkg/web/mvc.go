package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	"reflect"
)

type mvcMapping struct {
	name               string
	path               string
	method             string
	endpoint           endpoint.Endpoint
	decodeRequestFunc  httptransport.DecodeRequestFunc
	encodeRequestFunc  httptransport.EncodeRequestFunc
	decodeResponseFunc httptransport.DecodeResponseFunc
	encodeResponseFunc httptransport.EncodeResponseFunc
}

func NewMvcMapping(name, path, method string,
	endpoint endpoint.Endpoint,
	decodeRequestFunc httptransport.DecodeRequestFunc,
	encodeRequestFunc httptransport.EncodeRequestFunc,
	decodeResponseFunc httptransport.DecodeResponseFunc,
	encodeResponseFunc httptransport.EncodeResponseFunc) MvcMapping {
	return &mvcMapping{
		name: name,
		path: path,
		method: method,
		endpoint: endpoint,
		decodeRequestFunc: decodeRequestFunc,
		encodeRequestFunc: encodeRequestFunc,
		decodeResponseFunc: decodeResponseFunc,
		encodeResponseFunc: encodeResponseFunc,
	}
}

/*****************************
	MvcMapping Interface
******************************/
func (m *mvcMapping) Name() string {
	return m.name
}

func (m *mvcMapping) Path() string {
	return m.path
}

func (m *mvcMapping) Method() string {
	return m.method
}

func (m *mvcMapping) Endpoint() endpoint.Endpoint {
	return m.endpoint
}

func (m *mvcMapping) DecodeRequestFunc() httptransport.DecodeRequestFunc {
	return m.decodeRequestFunc
}

func (m *mvcMapping) EncodeRequestFunc() httptransport.EncodeRequestFunc {
	return m.encodeRequestFunc
}

func (m *mvcMapping) DecodeResponseFunc() httptransport.DecodeResponseFunc {
	return m.decodeResponseFunc
}

func (m *mvcMapping) EncodeResponseFunc() httptransport.EncodeResponseFunc {
	return m.encodeResponseFunc
}

/*****************************
	Func Metadata
******************************/
// MvcHandlerFuncValidator validate MvcHandlerFunc signature
type MvcHandlerFuncValidator func(f *reflect.Value) error

type mvcMetadata struct {
	function *reflect.Value
	request reflect.Type
	response reflect.Type
}

// MakeFuncMetadata uses reflect to analyze the given rest function and create a endpointFuncMetadata
// this function panic if given function have incorrect signature
// Caller can provide an optional validator to further validate function signature on top of default validation
func MakeFuncMetadata(endpointFunc MvcHandlerFunc, validator MvcHandlerFuncValidator) *mvcMetadata {
	f := reflect.ValueOf(endpointFunc)
	err := validateFunc(&f, validator)
	if err != nil {
		//TODO better fatal error handling
		panic(err)
	}

	t := f.Type()
	return &mvcMetadata{
		request: t.In(1),
		response: t.Out(0),
		function: &f,
	}
}

func validateFunc(f *reflect.Value, validator MvcHandlerFuncValidator) (err error) {
	// TODO define error type
	errTemplate := "invalid HTTP rest signature: %v"
	// For now, we check function signature at runtime.
	//I wish there is a way to check it at compile-time that I didn't know of
	t := f.Type()
	switch {
	case f.Kind() != reflect.Func:
		return errors.New(fmt.Sprintf(errTemplate, "rest should be a function"))
	case t.NumIn() < 2: // TODO|| t.In(0).ConvertibleTo(context.Context)
		return errors.New(fmt.Sprintf(errTemplate, "rest should have at least two input parameters, " +
			"which the first is context.Context and the second is a struct or pointer to struct"))
	case t.NumOut() < 2: // TODO|| t.In(0).ConvertibleTo(context.Context)
		return errors.New(fmt.Sprintf(errTemplate, "rest should have at least two output parameters, " +
			"which the first is struct or pointer to struct and the second is error"))
	}

	if validator != nil {
		return validator(f)
	}
	return nil
}

/*********************
	go-kit Endpoint
**********************/
// MakeEndpoint convert given mvcMetadata to kit/endpoint.Endpoint
func MakeEndpoint(s *mvcMetadata) endpoint.Endpoint {
	return func(c context.Context, request interface{}) (response interface{}, err error) {
		params := []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(request)}
		rets := s.function.Call(params)
		switch {
		case rets[1].IsZero():
			return rets[0].Interface(), nil
		default:
			return rets[0].Interface(), rets[1].Interface().(error)
		}
	}
}

/**********************************
	Generic Request Decoder
***********************************/
// bindable requestType can only be struct or pointer of struct
func MakeGinBindingDecodeRequestFunc(s *mvcMetadata) httptransport.DecodeRequestFunc {
	return func(c context.Context, r *http.Request) (response interface{}, err error) {
		if _,ok := c.(*gin.Context); !ok {
			// TODO return proper error
			return nil, nil
		}

		req := instantiateByType(s.request)
		ginCtx := c.(*gin.Context)

		// We always try to bind Header, Uri and Query. other bindings are determined by Content-Type (in ShouldBind)
		err = doBindRequest(req,
			ginCtx.ShouldBindHeader,
			ginCtx.ShouldBindUri,
			ginCtx.ShouldBindQuery,
			ginCtx.ShouldBind)

		if err != nil {
			fmt.Println(err)
		}
		return req, err
	}
}

type requestBindingFunc func(interface{}) error

func doBindRequest(obj interface{}, bindFuncs ...requestBindingFunc) (err error) {
	for _, bindFunc := range bindFuncs {
		if err = bindFunc(obj); err != nil {
			return
		}
	}
	return nil
}

func instantiateByType(t reflect.Type) interface{} {
	var obj reflect.Value
	switch t.Kind() {
	case reflect.Ptr:
		t = t.Elem()
		obj = reflect.New(t)
	default:
		obj = reflect.New(t)
	}
	return obj.Interface()
}

/**********************************
	Generic Response Decoder
***********************************/

// TODO Response Decode function, used for client

