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
	"net/http"
)

type mvcMapping struct {
	name               string
	group              string
	path               string
	method             string
	condition          RequestMatcher
	decodeRequestFunc  DecodeRequestFunc
	encodeResponseFunc EncodeResponseFunc
	encodeErrorFunc EncodeErrorFunc
	endpoint        MvcHandlerFunc
}

// NewMvcMapping create a MvcMapping
// It's recommended to use rest.MappingBuilder or template.MappingBuilder instead of this function:
// e.g.
// <code>
// rest.Put("/path/to/api").EndpointFunc(func...).Build()
// template.Post("/path/to/page").HandlerFunc(func...).Build()
// </code>
func NewMvcMapping(name, group, path, method string, condition RequestMatcher,
	mvcHandlerFunc MvcHandlerFunc,
	decodeRequestFunc DecodeRequestFunc,
	encodeResponseFunc EncodeResponseFunc,
	errorEncoder EncodeErrorFunc) MvcMapping {
	return &mvcMapping{
		name:               name,
		group:              group,
		path:               path,
		method:             method,
		condition:          condition,
		endpoint:           mvcHandlerFunc,
		decodeRequestFunc:  decodeRequestFunc,
		encodeResponseFunc: encodeResponseFunc,
		encodeErrorFunc:    errorEncoder,
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

func (m *mvcMapping) DecodeRequestFunc() DecodeRequestFunc {
	return m.decodeRequestFunc
}

func (m *mvcMapping) EncodeResponseFunc() EncodeResponseFunc {
	return m.encodeResponseFunc
}

func (m *mvcMapping) EncodeErrorFunc() EncodeErrorFunc {
	return m.encodeErrorFunc
}

func (m *mvcMapping) HandlerFunc() MvcHandlerFunc {
	return m.endpoint
}

/*********************
	Response
**********************/

type Response struct {
	SC int
	H  http.Header
	B  interface{}
}

// StatusCode implements StatusCoder
func (r Response) StatusCode() int {
	if i, ok := r.B.(StatusCoder); ok {
		return i.StatusCode()
	}
	return r.SC
}

// Headers implements Headerer
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
	MVC Handler
**********************/

type mvcHandler struct {
	reqDecoder     DecodeRequestFunc
	respEncoder    EncodeResponseFunc
	errEncoder  EncodeErrorFunc
	handlerFunc MvcHandlerFunc
}

func makeMvcHttpHandlerFunc(m MvcMapping, opts ...func(h *mvcHandler)) http.HandlerFunc {
	handler := &mvcHandler{
		reqDecoder:     m.DecodeRequestFunc(),
		respEncoder:    m.EncodeResponseFunc(),
		errEncoder:     m.EncodeErrorFunc(),
		handlerFunc:    m.HandlerFunc(),
	}
	for _, fn := range opts {
		fn(handler)
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		request, e := handler.reqDecoder(ctx, r)
		if e != nil {
			handler.errEncoder(ctx, e, rw)
			return
		}

		response, e := handler.handlerFunc(ctx, request)
		if e != nil {
			handler.errEncoder(ctx, e, rw)
			return
		}

		if e := handler.respEncoder(ctx, rw, response); e != nil {
			handler.errEncoder(ctx, e, rw)
			return
		}
	}
}

