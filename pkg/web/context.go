package web

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
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
	Mappings
 *********************************/
type Controller interface {
	Mappings() []Mapping
}

type HandlerFunc http.HandlerFunc

// MvcHandlerFunc is a function with following signature
// 	- two input parameters with 1st as context.Context and 2nd as <request>
// 	- two output parameters with 1st as <response> and 2nd as error
// See rest.EndpointFunc, template.ModelViewHandlerFunc
type MvcHandlerFunc interface{}

// Mapping
type Mapping interface {
	Name() string
}

// StaticMapping defines static assets mapping. e.g. javascripts, css, images, etc
type StaticMapping interface {
	Mapping
	Path() string
	StaticRoot() string
}

// RoutedMapping
type RoutedMapping interface {
	Mapping
	Path() string
	Method() string
	Condition() RequestMatcher
}

// SimpleMapping
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

/*********************************
	Routing Matchers
 *********************************/
// Route contains information needed for registering handler func in gin.Engine
type Route struct {
	Method string
	Path string
	Group string
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
// implmenets SimpleMapping
type simpleMapping struct {
	name        string
	path        string
	method      string
	condition   RequestMatcher
	handlerFunc HandlerFunc
}

func NewSimpleMapping(name, path, method string, condition RequestMatcher, handlerFunc HandlerFunc) SimpleMapping {
	return &simpleMapping{
		name: name,
		path: path,
		method: method,
		condition: condition,
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


