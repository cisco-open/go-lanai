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

package middleware

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MappingBuilder struct {
	name       string
	middleware web.Middleware
	matcher    web.RouteMatcher
	order      int
	// overrides
	condition   web.RequestMatcher
	handlerFunc interface{}
}

func NewBuilder(names ...string) *MappingBuilder {
	name := "unknown"
	if len(names) > 0 {
		name = names[0]
	}
	return &MappingBuilder{
		name: name,
		order: 0,
	}
}

/*****************************
	Public
******************************/

func (b *MappingBuilder) Name(name string) *MappingBuilder {
	b.name = name
	return b
}

func (b *MappingBuilder) Order(order int) *MappingBuilder {
	b.order = order
	return b
}

func (b *MappingBuilder) With(middleware web.Middleware) *MappingBuilder {
	b.middleware = middleware
	return b
}

func (b *MappingBuilder) ApplyTo(matcher web.RouteMatcher) *MappingBuilder {
	b.matcher = matcher
	return b
}

// Use set middleware handler. Support:
// - gin.HandlerFunc
// - http.HandlerFunc
// - web.HandlerFunc
func (b *MappingBuilder) Use(handlerFunc interface{}) *MappingBuilder {
	switch handlerFunc.(type) {
	case gin.HandlerFunc, http.HandlerFunc, web.HandlerFunc:
		b.handlerFunc = handlerFunc
	default:
		panic(fmt.Errorf("unsupported HandlerFunc type: %T", handlerFunc))
	}
	return b
}

func (b *MappingBuilder) WithCondition(condition web.RequestMatcher) *MappingBuilder {
	b.condition = condition
	return b
}

func (b *MappingBuilder) Build() web.MiddlewareMapping {
	var condition web.RequestMatcher
	var handlerFunc interface{}
	if b.middleware != nil {
		handlerFunc = b.middleware.HandlerFunc()
		if conditional,ok := b.middleware.(web.ConditionalMiddleware); ok {
			condition = conditional.Condition()
		}
	}

	if b.handlerFunc != nil {
		handlerFunc = b.handlerFunc
	}

	if b.condition != nil {
		condition = b.condition
	}

	switch v := handlerFunc.(type) {
	case gin.HandlerFunc:
		return web.NewMiddlewareGinMapping(b.name, b.order, b.matcher, condition, v)
	case http.HandlerFunc:
		return web.NewMiddlewareMapping(b.name, b.order, b.matcher, condition, web.HandlerFunc(v))
	case web.HandlerFunc:
		return web.NewMiddlewareMapping(b.name, b.order, b.matcher, condition, v)
	default:
		panic(fmt.Errorf("unable to build '%s' middleware mapping: unsupported HandlerFunc type %v. please use With(...) or Use(...)", b.name, handlerFunc))
	}
}

/*****************************
	Getters
******************************/

func (b *MappingBuilder) GetRouteMatcher() web.RouteMatcher {
	return b.matcher
}

func (b *MappingBuilder) GetCondition() web.RequestMatcher {
	return b.condition
}

func (b *MappingBuilder) GetName() string {
	return b.name
}

func (b *MappingBuilder) GetOrder() int {
	return b.order
}

/*****************************
	Helpers
******************************/



