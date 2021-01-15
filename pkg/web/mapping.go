package web

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"regexp"
)

// Validation reference: https://godoc.org/github.com/go-playground/validator#hdr-Baked_In_Validators_and_Tags

var (
	pathParamPattern, _ = regexp.Compile(`\/:[^\/]*`)
)

/*********************************
	Mappings
 *********************************/
type Controller interface {
	Mappings() []Mapping
}

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
	HandlerFunc() gin.HandlerFunc
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
	HandlerFunc() gin.HandlerFunc
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
	handlerFunc gin.HandlerFunc
}

func NewSimpleMapping(name, path, method string, condition RequestMatcher, handlerFunc gin.HandlerFunc) SimpleMapping {
	return &simpleMapping{
		name: name,
		path: path,
		method: method,
		handlerFunc: handlerFunc,
	}
}

func (g simpleMapping) HandlerFunc() gin.HandlerFunc {
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