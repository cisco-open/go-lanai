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

type ConditionalMiddleware interface {
	Condition() RequestMatcher
}

type Middleware interface {
	HandlerFunc() HandlerFunc
}

type middlewareMapping struct {
	name        string
	order       int
	matcher     RouteMatcher
	condition   RequestMatcher
	handlerFunc HandlerFunc
}

func NewMiddlewareMapping(name string, order int, matcher RouteMatcher, cond RequestMatcher, handlerFunc HandlerFunc) MiddlewareMapping {
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

func (mm middlewareMapping) HandlerFunc() HandlerFunc {
	return mm.handlerFunc
}



