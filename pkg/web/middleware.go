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

import "net/http"

// ConditionalMiddleware is an additional interface that a Middleware can implement to control when the middleware is applied.
// e.g. a middleware want to be applied if request's header contains "Authorization"
type ConditionalMiddleware interface {
	Condition() RequestMatcher
}

// Middleware defines a http.HandlerFunc to be used by MiddlewareMapping and middleware.MappingBuilder
type Middleware interface {
	HandlerFunc() http.HandlerFunc
}

type middlewareMapping struct {
	name        string
	order       int
	matcher     RouteMatcher
	condition   RequestMatcher
	handlerFunc http.HandlerFunc
}

// NewMiddlewareMapping create a MiddlewareMapping with http.HandlerFunc
// It's recommended to use middleware.MappingBuilder instead of this function:
// e.g.
// <code>
// middleware.NewBuilder("my-auth").Order(-10).Use(func...).Build()
// </code>
func NewMiddlewareMapping(name string, order int, matcher RouteMatcher, cond RequestMatcher, handlerFunc http.HandlerFunc) MiddlewareMapping {
	return &middlewareMapping {
		name: name,
		matcher: matcher,
		order: order,
		condition: cond,
		handlerFunc: handlerFunc,
	}
}

func (mm middlewareMapping) Name() string {
	return mm.name
}

func (mm middlewareMapping) Matcher() RouteMatcher {
	return mm.matcher
}

func (mm middlewareMapping) Order() int {
	return mm.order
}

func (mm middlewareMapping) Condition() RequestMatcher {
	return mm.condition
}

func (mm middlewareMapping) HandlerFunc() http.HandlerFunc {
	return mm.handlerFunc
}



