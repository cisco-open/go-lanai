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
	"github.com/cisco-open/go-lanai/pkg/utils/matcher"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	"regexp"
)

// Validation reference: https://godoc.org/github.com/go-playground/validator#hdr-Baked_In_Validators_and_Tags

var (
	pathParamPattern, _ = regexp.Compile(`\/:[^\/]*`)
)

/*********************************
	Customization
 *********************************/

// Customizer is invoked by Registrar at the beginning of initialization,
// customizers can register anything except for additional customizers
// If a customizer retains the given context in anyway, it should also implement PostInitCustomizer to release it
type Customizer interface {
	Customize(ctx context.Context, r *Registrar) error
}

// PostInitCustomizer is invoked by Registrar after initialization, register anything in PostInitCustomizer.PostInit
// would cause error or takes no effect
type PostInitCustomizer interface {
	Customizer
	PostInit(ctx context.Context, r *Registrar) error
}

type EngineOptions func(*Engine)

/*********************************
	Request
 *********************************/

// RequestRewriter handles request rewrite. e.g. rewrite http.Request.URL.Path
type RequestRewriter interface {
	// HandleRewrite take the rewritten request and put it through the entire handling cycle.
	// The http.Request.Context() is carried over
	// Note: if no error is returned, caller should stop processing the original request and discard the original request
	HandleRewrite(rewritten *http.Request) error
}

/*********************************
	Response
 *********************************/

// StatusCoder is same interface defined in "github.com/go-kit/kit/transport/http"
// this interface is majorly used internally with error handling
type StatusCoder interface {
	StatusCode() int
}

// Headerer is same interface defined in "github.com/go-kit/kit/transport/http"
// this interface is majorly used internally with error handling
// If an error value implements Headerer, the provided headers will be applied to the response writer, after
// the Content-Type is set.
type Headerer interface {
	Headers() http.Header
}

// BodyContainer is a reponse body wrapping interface.
// this interface is majorly used internally for mapping
type BodyContainer interface {
	Body() interface{}
}

/*********************************
	Error Translator
 *********************************/

// ErrorTranslator can be registered via web.Registrar
// it will contribute our MvcMapping's error handling process.
// Note: it won't contribute Middleware's error handling
//
// Implementing Notes:
// 	1. if it doesn't handle the error, return same error
//  2. if custom StatusCode is required, make the returned error implement StatusCoder
//  3. if custom Header is required, make the returned error implement Headerer
//  4. we have HttpError to help with custom Headerer and StatusCoder implementation
type ErrorTranslator interface {
	Translate(ctx context.Context, err error) error
}

// ErrorTranslateFunc is similar to ErrorTranslator in function format. Mostly used for selective error translation
// registration (ErrorHandlerMapping). Same implementing rules applies
type ErrorTranslateFunc func(ctx context.Context, err error) error
func (fn ErrorTranslateFunc) Translate(ctx context.Context, err error) error {
	return fn(ctx, err)
}

/*********************************
	Mappings
 *********************************/

type Controller interface {
	Mappings() []Mapping
}

// HandlerFunc have same signature as http.HandlerFunc with additional assurance:
// - the http.Request used on this HandlerFunc version contains a mutable context utils.MutableContext
type HandlerFunc http.HandlerFunc

// MvcHandlerFunc is a function with following signature
// 	- one or two input parameters with 1st as context.Context and 2nd as <request>
// 	- at least two output parameters with 2nd last as <response> and last as error
// See rest.EndpointFunc, template.ModelViewHandlerFunc
type MvcHandlerFunc interface{}

// Mapping generic interface for all kind of endpoint mappings
type Mapping interface {
	Name() string
}

// StaticMapping defines static assets mapping. e.g. javascripts, css, images, etc
type StaticMapping interface {
	Mapping
	Path() string
	StaticRoot() string
	Aliases() map[string]string
	AddAlias(path, filePath string) StaticMapping
}

// RoutedMapping for endpoints that matches specific path, method and optionally a RequestMatcher as condition
type RoutedMapping interface {
	Mapping
	Group() string
	Path() string
	Method() string
	Condition() RequestMatcher
}

// SimpleMapping endpoints that are directly implemented as HandlerFunc
type SimpleMapping interface {
	RoutedMapping
	HandlerFunc() HandlerFunc
}

// MvcMapping defines HTTP handling that follows MVC pattern
// could be EndpointMapping or TemplateMapping
type MvcMapping interface {
	RoutedMapping
	Endpoint() endpoint.Endpoint
	DecodeRequestFunc() httptransport.DecodeRequestFunc
	EncodeRequestFunc() httptransport.EncodeRequestFunc
	DecodeResponseFunc() httptransport.DecodeResponseFunc
	EncodeResponseFunc() httptransport.EncodeResponseFunc
	ErrorEncoder() httptransport.ErrorEncoder
}

// EndpointMapping defines REST API mapping.
// REST API is usually implemented by Controller and accept/produce JSON objects
type EndpointMapping MvcMapping

// TemplateMapping defines templated MVC mapping. e.g. html templates
// Templated MVC is usually implemented by Controller and produce a template and model for dynamic html generation
type TemplateMapping MvcMapping

type MiddlewareMapping interface {
	Mapping
	Matcher() RouteMatcher
	Order() int
	Condition() RequestMatcher
	HandlerFunc() HandlerFunc
}

type ErrorTranslateMapping interface {
	Mapping
	Matcher() RouteMatcher
	Order() int
	Condition() RequestMatcher
	TranslateFunc() ErrorTranslateFunc
}

/*********************************
	Routing Matchers
 *********************************/

// Route contains information needed for registering handler func in gin.Engine
type Route struct {
	Method string
	Path   string
	Group  string
}

// RouteMatcher is a typed ChainableMatcher that accept *Route or Route
type RouteMatcher interface {
	matcher.ChainableMatcher
}

// RequestMatcher is a typed ChainableMatcher that accept *http.Request or http.Request
type RequestMatcher interface {
	matcher.ChainableMatcher
}

// NormalizedPath removes path parameter name from path.
// path "/path/with/:param" is effectively same as "path/with/:other_param_name"
func NormalizedPath(path string) string {
	return pathParamPattern.ReplaceAllString(path, "/:var")
}

/*********************************
	SimpleMapping
 *********************************/

// simpleMapping implements SimpleMapping
type simpleMapping struct {
	name        string
	group       string
	path        string
	method      string
	condition   RequestMatcher
	handlerFunc HandlerFunc
}

func NewSimpleMapping(name, group, path, method string, condition RequestMatcher, handlerFunc HandlerFunc) SimpleMapping {
	return &simpleMapping{
		name:        name,
		group:       group,
		path:        path,
		method:      method,
		condition:   condition,
		handlerFunc: handlerFunc,
	}
}

func (g simpleMapping) HandlerFunc() HandlerFunc {
	return g.handlerFunc
}

func (g simpleMapping) Condition() RequestMatcher {
	return g.condition
}

func (g simpleMapping) Method() string {
	return g.method
}

func (g simpleMapping) Group() string {
	return g.group
}

func (g simpleMapping) Path() string {
	return g.path
}

func (g simpleMapping) Name() string {
	return g.name
}

/*********************************
	orderedServerOption
 *********************************/

// orderedServerOption wraps go-kit's httptransport.ServerOption and provide ordering
type orderedServerOption struct {
	httptransport.ServerOption
	order int
}

func (o orderedServerOption) Order() int {
	return o.order
}

func newOrderedServerOption(opt httptransport.ServerOption, order int) *orderedServerOption {
	return &orderedServerOption{ServerOption: opt, order: order}
}
