package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"io/ioutil"
	"net/http"
	"reflect"
)

type MappingFunc func(engine *gin.Engine, handlerFunc gin.HandlerFunc)

// EndpointFunc is a function with following signature
// 	- two input parameters with 1st as context.Context and 2nd as <request>
// 	- two output parameters with 1st as <response> and 2nd as error
// where
// <request>:   a struct or a pointer to a struct whose fields are properly tagged
// <response>:  a struct or a pointer to a struct whose fields are properly tagged.
// 				if decoding is not supported (endpoint not used by any go client), it can be an interface{}
type EndpointFunc interface{}

type Mapping struct {
	Path string
	Method string
	EndpointFunc EndpointFunc

	// Overrides
	MappingFunc MappingFunc
	Endpoint endpoint.Endpoint
	DecodeRequestFunc httptransport.DecodeRequestFunc
	EncodeRequestFunc httptransport.EncodeRequestFunc
	DecodeResponseFunc httptransport.DecodeResponseFunc
	EncodeResponseFunc httptransport.EncodeResponseFunc
}

type EndpointFuncMetadata struct {
	function *reflect.Value
	request reflect.Type
	response reflect.Type
}

func MakeMapping(method string, path string, endpointFunc EndpointFunc) *Mapping {
	// TODO more validation and better error handling
	if path == "" || method == "" {
		panic("invalid mapping")
	}

	if endpointFunc == nil {
		return &Mapping{
			Path: path,
			Method: method,
			// Overrides
			EncodeRequestFunc: GenericEncodeRequestFunc,
			DecodeRequestFunc: httptransport.NopRequestDecoder,
			EncodeResponseFunc: GenericEncodeResponseFunc,
			//DecodeResponseFunc: TODO
		}
	}

	metadata := MakeEndpointFuncMetadata(endpointFunc)
	return &Mapping {
		Path: path,
		Method: method,
		EndpointFunc: endpointFunc,
		// Overrides
		Endpoint: MakeEndpoint(metadata),
		EncodeRequestFunc: GenericEncodeRequestFunc,
		DecodeRequestFunc: MakeGinBindingDecodeRequestFunc(metadata),
		EncodeResponseFunc: GenericEncodeResponseFunc,
		//DecodeResponseFunc: TODO
	}
}

// MakeEndpointFuncMetadata uses reflect to analyze the given endpoint function and create a EndpointFuncMetadata
// this function panic if given function have incorrect signature
func MakeEndpointFuncMetadata(endpointFunc EndpointFunc) *EndpointFuncMetadata {
	f := reflect.ValueOf(endpointFunc)
	t, err := validateEndpointFunc(&f)
	if err != nil {
		//TODO better fatal error handling
		panic(err)
	}

	return &EndpointFuncMetadata{
		request: t.In(1),
		response: t.Out(0),
		function: &f,
	}
}

func validateEndpointFunc(f *reflect.Value) (reflect.Type, error) {
	// TODO define error type
	errTemplate := "invalid HTTP endpoint signature: %v"
	// For now, we check function signature at runtime.
	//I wish there is a way to check it at compile-time that I didn't know of
	t := f.Type()
	switch {
	case f.Kind() != reflect.Func:
		return nil, errors.New(fmt.Sprintf(errTemplate, "endpoint should be a function"))
	case t.NumIn() < 2: // TODO|| t.In(0).ConvertibleTo(context.Context)
		return nil, errors.New(fmt.Sprintf(errTemplate, "endpoint should have at least two input parameters, " +
			"which the first is context.Context and the second is a struct or pointer to struct"))
	case t.NumOut() < 2: // TODO|| t.In(0).ConvertibleTo(context.Context)
		return nil, errors.New(fmt.Sprintf(errTemplate, "endpoint should have at least two output parameters, " +
			"which the first is struct or pointer to struct and the second is error"))
	}

	return t, nil
}

/*********************
	go-kit Endpoint
**********************/
// MakeEndpoint convert given EndpointFuncMetadata to kit/endpoint.Endpoint
func MakeEndpoint(s *EndpointFuncMetadata) endpoint.Endpoint {
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
	go-kit Request Decode/Encode
***********************************/
// bindable requestType can only be struct or pointer of struct
func MakeGinBindingDecodeRequestFunc(s *EndpointFuncMetadata) httptransport.DecodeRequestFunc {
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

func GenericEncodeRequestFunc(_ context.Context, r *http.Request, request interface{}) error {
	// review this part
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

/**********************************
	go-kit Response Decode/Encode
***********************************/
func GenericEncodeResponseFunc(_ context.Context, w http.ResponseWriter, response interface{}) error {
	// TODO review this part
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// TODO Response Decode function, used for client
