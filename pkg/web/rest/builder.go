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

package rest

import (
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/go-kit/kit/endpoint"
    httptransport "github.com/go-kit/kit/transport/http"
    "net/http"
)

// EndpointFunc is a function with following signature
// 	- one or two input parameters with the 1st as context.Context and the 2nd as <request>
// 	- at least two output parameters with the 2nd last as <response> and the last as error
// where
// <request>:   a struct or a pointer to a struct whose fields are properly tagged
// <response>:  supported types are (will support more in the future):
//				- a struct or a pointer to a struct whose fields are properly tagged.
//				- interface{}, if decoding is not supported (rest not used by any go client)
//				- map[string]interface{}
//				- string
//				- []byte
// e.g.: func(context.Context, request *AnyStructWithTag) (response *AnyStructWithTag, error) {...}
type EndpointFunc web.MvcHandlerFunc

type MappingBuilder struct {
	name               string
	group              string
	path               string
	method             string
	condition          web.RequestMatcher
	endpointFunc       EndpointFunc
	endpoint           endpoint.Endpoint
	decodeRequestFunc  httptransport.DecodeRequestFunc
	encodeRequestFunc  httptransport.EncodeRequestFunc
	decodeResponseFunc httptransport.DecodeResponseFunc
	encodeResponseFunc httptransport.EncodeResponseFunc
}

func New(names ...string) *MappingBuilder {
	var name string
	if len(names) > 0 {
		name = names[0]
	}
	return &MappingBuilder{
		name:   name,
		method: web.MethodAny,
	}
}

// Convenient Constructors

func Any(path string) *MappingBuilder {
	return New().Path(path).Method(web.MethodAny)
}

func Get(path string) *MappingBuilder {
	return New().Get(path)
}

func Post(path string) *MappingBuilder {
	return New().Post(path)
}

func Put(path string) *MappingBuilder {
	return New().Put(path)
}

func Patch(path string) *MappingBuilder {
	return New().Patch(path)
}

func Delete(path string) *MappingBuilder {
	return New().Delete(path)
}

func Options(path string) *MappingBuilder {
	return New().Options(path)
}

func Head(path string) *MappingBuilder {
	return New().Head(path)
}

/*****************************
	Public
******************************/
func (b *MappingBuilder) Name(name string) *MappingBuilder {
	b.name = name
	return b
}

func (b *MappingBuilder) Group(group string) *MappingBuilder {
	b.group = group
	return b
}

func (b *MappingBuilder) Path(path string) *MappingBuilder {
	b.path = path
	return b
}

func (b *MappingBuilder) Method(method string) *MappingBuilder {
	b.method = method
	return b
}

func (b *MappingBuilder) Condition(condition web.RequestMatcher) *MappingBuilder {
	b.condition = condition
	return b
}

func (b *MappingBuilder) EndpointFunc(endpointFunc EndpointFunc) *MappingBuilder {
	b.endpointFunc = endpointFunc
	return b
}

// Convenient setters
func (b *MappingBuilder) Get(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodGet)
}

func (b *MappingBuilder) Post(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodPost)
}

func (b *MappingBuilder) Put(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodPut)
}

func (b *MappingBuilder) Patch(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodPatch)
}

func (b *MappingBuilder) Delete(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodDelete)
}

func (b *MappingBuilder) Options(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodOptions)
}

func (b *MappingBuilder) Head(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodHead)
}

// Overrides
func (b *MappingBuilder) Endpoint(endpoint endpoint.Endpoint) *MappingBuilder {
	b.endpoint = endpoint
	return b
}

func (b *MappingBuilder) DecodeRequestFunc(f httptransport.DecodeRequestFunc) *MappingBuilder {
	b.decodeRequestFunc = f
	return b
}

func (b *MappingBuilder) EncodeRequestFunc(f httptransport.EncodeRequestFunc) *MappingBuilder {
	b.encodeRequestFunc = f
	return b
}

func (b *MappingBuilder) DecodeResponseFunc(f httptransport.DecodeResponseFunc) *MappingBuilder {
	b.decodeResponseFunc = f
	return b
}

func (b *MappingBuilder) EncodeResponseFunc(f httptransport.EncodeResponseFunc) *MappingBuilder {
	b.encodeResponseFunc = f
	return b
}

func (b *MappingBuilder) Build() web.EndpointMapping {
	if err := b.validate(); err != nil {
		panic(err)
	}
	return b.buildMapping()
}

/*****************************
	Private
******************************/
type mapping struct {
	endpoint           endpoint.Endpoint
	decodeRequestFunc  httptransport.DecodeRequestFunc
	encodeRequestFunc  httptransport.EncodeRequestFunc
	decodeResponseFunc httptransport.DecodeResponseFunc
	encodeResponseFunc httptransport.EncodeResponseFunc
}

func (b *MappingBuilder) validate() (err error) {
	if b.path == "" && (b.group == "" || b.group == "/") {
		err = errors.New("empty path")
	}
	return
}

func (b *MappingBuilder) buildMapping() web.MvcMapping {
	if b.method == "" {
		b.method = web.MethodAny
	}

	if b.name == "" {
		b.name = fmt.Sprintf("%s %s%s", b.method, b.group, b.path)
	}

	m := &mapping{
		decodeRequestFunc:  httptransport.NopRequestDecoder,
		encodeRequestFunc:  jsonEncodeRequestFunc,
		decodeResponseFunc: nil, // TODO
		encodeResponseFunc: web.JsonResponseEncoder(),
	}

	if b.endpointFunc != nil {
		metadata := web.MakeFuncMetadata(b.endpointFunc, nil)
		m.endpoint = web.MakeEndpoint(metadata)
		m.decodeRequestFunc = web.MakeGinBindingDecodeRequestFunc(metadata)
	}

	b.customize(m)
	return web.NewMvcMapping(b.name, b.group, b.path, b.method, b.condition,
		m.endpoint, m.decodeRequestFunc, m.encodeRequestFunc,
		m.decodeResponseFunc, m.encodeResponseFunc,
		web.JsonErrorEncoder())
}

func (b *MappingBuilder) customize(m *mapping) {
	if b.endpoint != nil {
		m.endpoint = b.endpoint
	}

	if b.encodeRequestFunc != nil {
		m.encodeRequestFunc = b.encodeRequestFunc
	}

	if b.decodeRequestFunc != nil {
		m.decodeRequestFunc = b.decodeRequestFunc
	}

	if b.encodeResponseFunc != nil {
		m.encodeResponseFunc = b.encodeResponseFunc
	}

	if b.decodeResponseFunc != nil {
		m.decodeResponseFunc = b.decodeResponseFunc
	}
}
