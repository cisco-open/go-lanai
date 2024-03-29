// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"io"
	"net/http"
	"reflect"
)

type mvcMapping struct {
	name               string
	group              string
	path               string
	method             string
	condition          RequestMatcher
	endpoint           endpoint.Endpoint
	decodeRequestFunc  httptransport.DecodeRequestFunc
	encodeRequestFunc  httptransport.EncodeRequestFunc
	decodeResponseFunc httptransport.DecodeResponseFunc
	encodeResponseFunc httptransport.EncodeResponseFunc
	errorEncoder       httptransport.ErrorEncoder
}

// NewMvcMapping exported for inter-package usage only. Use builders.
func NewMvcMapping(name, group, path, method string, condition RequestMatcher,
	endpoint endpoint.Endpoint,
	decodeRequestFunc httptransport.DecodeRequestFunc,
	encodeRequestFunc httptransport.EncodeRequestFunc,
	decodeResponseFunc httptransport.DecodeResponseFunc,
	encodeResponseFunc httptransport.EncodeResponseFunc,
	errorEncoder httptransport.ErrorEncoder) MvcMapping { //SuppressWarnings go:S107 This function is internal use only, converting params to a struct provide no additional benefit
	return &mvcMapping{
		name:               name,
		group:              group,
		path:               path,
		method:             method,
		condition:          condition,
		endpoint:           endpoint,
		decodeRequestFunc:  decodeRequestFunc,
		encodeRequestFunc:  encodeRequestFunc,
		decodeResponseFunc: decodeResponseFunc,
		encodeResponseFunc: encodeResponseFunc,
		errorEncoder:       errorEncoder,
	}
}

/*****************************
	MvcMapping Interface
******************************/

func (m *mvcMapping) Name() string {
	return m.name
}

func (m *mvcMapping) Group() string {
	return m.group
}

func (m *mvcMapping) Path() string {
	return m.path
}

func (m *mvcMapping) Method() string {
	return m.method
}

func (m *mvcMapping) Condition() RequestMatcher {
	return m.condition
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

func (m *mvcMapping) ErrorEncoder() httptransport.ErrorEncoder {
	return m.errorEncoder
}

/*********************
	Response
**********************/

type Response struct {
	SC int
	H  http.Header
	B  interface{}
}

// StatusCode implements httptransport.StatusCoder and StatusCoder
func (r Response) StatusCode() int {
	if i, ok := r.B.(StatusCoder); ok {
		return i.StatusCode()
	}
	return r.SC
}

// Headers implements httptransport.Headerer and Headerer
func (r Response) Headers() http.Header {
	if i, ok := r.B.(Headerer); ok {
		return i.Headers()
	}
	return r.H
}

// Body implements BodyContainer
func (r Response) Body() interface{} {
	if i, ok := r.B.(BodyContainer); ok {
		return i.Body()
	}
	return r.B
}

/**********************************
	LazyHeaderWriter
***********************************/

// LazyHeaderWriter makes sure that status code and headers is overwritten at last second (when invoke Write([]byte) (int, error).
// Calling WriteHeader(int) would not actually send the header. Calling it multiple times to update status code
// Doing so allows response encoder and error handling to send different header and status code
type LazyHeaderWriter struct {
	http.ResponseWriter
	sc     int
	header http.Header
}

func (w *LazyHeaderWriter) Header() http.Header {
	return w.header
}

func (w *LazyHeaderWriter) WriteHeader(code int) {
	w.sc = code
}

func (w *LazyHeaderWriter) Write(p []byte) (int, error) {
	w.WriteHeaderNow()
	return w.ResponseWriter.Write(p)
}

func (w *LazyHeaderWriter) WriteHeaderNow() {
	// Merge header overwrite
	for k, v := range w.header {
		w.ResponseWriter.Header()[k] = v
	}
	w.ResponseWriter.WriteHeader(w.sc)
}

func NewLazyHeaderWriter(w http.ResponseWriter) *LazyHeaderWriter {
	// make a copy of current header from wrapped writer
	header := make(http.Header)
	for k, v := range w.Header() {
		header[k] = v
	}
	return &LazyHeaderWriter{ResponseWriter: w, sc: http.StatusOK, header: header}
}

/*********************
	go-kit Endpoint
**********************/

// MakeEndpoint convert given mvcMetadata to kit/endpoint.Endpoint
func MakeEndpoint(m *mvcMetadata) endpoint.Endpoint {
	// Note: we assume given metadata is valid, so we don't do out-of-index or type check
	return func(c context.Context, request interface{}) (response interface{}, err error) {
		// prepare input params
		in := make([]reflect.Value, m.in.count)
		in[m.in.context.i] = reflect.ValueOf(c)
		if m.in.request.isValid() {
			in[m.in.request.i] = reflect.ValueOf(request)
		}

		out := m.function.Call(in)

		// post process output
		err, _ = out[m.out.error.i].Interface().(error)
		response = out[m.out.response.i].Interface()
		if !m.out.sc.isValid() && !m.out.header.isValid() {
			return response, err
		}

		// if necessary, wrap the response
		wrapper := &Response{B: response}
		if m.out.sc.isValid() {
			wrapper.SC = int(out[m.out.sc.i].Int())
		}

		if m.out.header.isValid() {
			wrapper.H, _ = out[m.out.header.i].Interface().(http.Header)
		}

		return wrapper, err
	}
}

/**********************************
	go-kit RequestDetails Decoder
***********************************/

// MakeGinBindingDecodeRequestFunc
// bindable requestType can only be struct or pointer of struct
func MakeGinBindingDecodeRequestFunc(s *mvcMetadata) httptransport.DecodeRequestFunc {
	// No need to decode
	if s.request == nil || isHttpRequestPtr(s.request) {
		return func(c context.Context, r *http.Request) (request interface{}, err error) {
			return r, nil
		}
	}
	// decode request using GinBinding
	return func(c context.Context, r *http.Request) (request interface{}, err error) {
		ginCtx := GinContext(c)
		if ginCtx == nil {
			return nil, NewHttpError(http.StatusInternalServerError, errors.New("context issue"))
		}

		toBind, toRet := instantiateByType(s.request)

		// We always try to bind H, Uri and Query. other bindings are determined by Content-Type (in ShouldBind)
		err = bind(toBind,
			ginCtx.ShouldBindHeader,
			ginCtx.ShouldBindUri,
			ginCtx.ShouldBindQuery)

		if err != nil {
			return nil, translateBindingError(err)
		}

		err = ginCtx.ShouldBind(toBind)

		if err != nil && !(errors.Is(err, io.EOF) && r.ContentLength <= 0) {
			return nil, translateBindingError(err)
		}
		return toRet.Interface(), validateBinding(c, toBind)
	}
}

type bindingFunc func(interface{}) error

func bind(obj interface{}, bindings ...bindingFunc) (err error) {
	for _, b := range bindings {
		if err = b(obj); err != nil {
			return
		}
	}
	return
}

func translateBindingError(err error) error {
	return NewBindingError(err)
}

func validateBinding(ctx context.Context, obj interface{}) error {
	if bindingValidator == nil {
		return nil
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		return bindingValidator.StructCtx(ctx, obj)
	default:
		return nil
	}
}

// returned ptr is the pointer regardless if given type is Ptr or other type
// returned value is actually the value with given type
func instantiateByType(t reflect.Type) (ptr interface{}, value *reflect.Value) {
	var p reflect.Value
	switch t.Kind() {
	case reflect.Ptr:
		t = t.Elem()
		p = reflect.New(t)
		return p.Interface(), &p
	default:
		p = reflect.New(t)
		v := p.Elem()
		return p.Interface(), &v
	}
}
