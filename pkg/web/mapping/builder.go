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

package mapping

import (
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/gin-gonic/gin"
    "net/http"
)

/*********************************
	SimpleMappingBuilder
 *********************************/

//goland:noinspection GoNameStartsWithPackageName
type MappingBuilder struct {
	name        string
	group       string
	path        string
	method      string
	condition   web.RequestMatcher
	handlerFunc interface{}
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

// HandlerFunc support
// - gin.HandlerFunc
// - http.HandlerFunc
// - web.HandlerFunc
func (b *MappingBuilder) HandlerFunc(handlerFunc interface{}) *MappingBuilder {
	switch handlerFunc.(type) {
	case gin.HandlerFunc, http.HandlerFunc, web.HandlerFunc:
		b.handlerFunc = handlerFunc
	default:
		panic(fmt.Errorf("unsupported HandlerFunc type: %T", handlerFunc))
	}
	b.handlerFunc = handlerFunc
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

func (b *MappingBuilder) Build() web.SimpleMapping {
	if err := b.validate(); err != nil {
		panic(err)
	}
	return b.buildMapping()
}

/*****************************
	Getters
******************************/

func (b *MappingBuilder) GetPath() string {
	return b.path
}

func (b *MappingBuilder) GetMethod() string {
	return b.method
}

func (b *MappingBuilder) GetCondition() web.RequestMatcher {
	return b.condition
}

func (b *MappingBuilder) GetName() string {
	return b.name
}

/*****************************
	Private
******************************/
func (b *MappingBuilder) validate() (err error) {
	switch {
	case b.path == "" && (b.group == "" || b.group == "/"):
		err = errors.New("empty path")
	case b.handlerFunc == nil:
		err = errors.New("handler func not specified")
	}
	return
}

func (b *MappingBuilder) buildMapping() web.SimpleMapping {
	if b.method == "" {
		b.method = web.MethodAny
	}

	if b.name == "" {
		b.name = fmt.Sprintf("%s %s%s", b.method, b.group, b.path)
	}

	switch b.handlerFunc.(type) {
	case gin.HandlerFunc:
		return web.NewSimpleGinMapping(b.name, b.group, b.path, b.method, b.condition, b.handlerFunc.(gin.HandlerFunc))
	case http.HandlerFunc, web.HandlerFunc:
		return web.NewSimpleMapping(b.name, b.group, b.path, b.method, b.condition, b.handlerFunc.(web.HandlerFunc))
	default:
		panic(fmt.Errorf("unsupported HandlerFunc type: %T", b.handlerFunc))
	}
}
