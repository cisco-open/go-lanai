package web

import (
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type Controller interface {
	Endpoints() []Mapping
}

// MvcHandlerFunc is a function with following signature
// 	- two input parameters with 1st as context.Context and 2nd as <request>
// 	- two output parameters with 1st as <response> and 2nd as error
// See rest.EndpointFunc, template.ModelViewHandlerFunc
type MvcHandlerFunc interface{}

type Mapping interface {
	Name() string
	Path() string
	Method() string
}

// StaticMapping defines static assets mapping. e.g. javascripts, css, images, etc
type StaticMapping interface {
	Mapping
	StaticRoot() string
}

// MvcMapping defines HTTP handling that follows MVC pattern
// could be either EndpointMapping or TemplateMapping
type MvcMapping interface {
	Mapping
	Endpoint() endpoint.Endpoint
	DecodeRequestFunc() httptransport.DecodeRequestFunc
	EncodeRequestFunc() httptransport.EncodeRequestFunc
	DecodeResponseFunc() httptransport.DecodeResponseFunc
	EncodeResponseFunc() httptransport.EncodeResponseFunc
}

// EndpointMapping defines REST API mapping.
// REST API is usually implemented by Controller and accept/produce JSON objects
type EndpointMapping MvcMapping

// TODO
// TemplateMapping defines templated MVC mapping. e.g. html templates
// Templated MVC is usually implemented by Controller and produce a template and model for dynamic html generation
type TemplateMapping MvcMapping


