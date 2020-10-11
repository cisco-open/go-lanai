package web

import (
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type MappingFunc func(registrar *Registrar)

type Controller interface {
	Endpoints() []Mapping
}

type Mapping interface {
	Name() string
	Path() string
	Method() string
}

// EndpointMapping defines REST API mapping.
// REST API is usually implemented by Controller and accept/produce JSON objects
type EndpointMapping interface {
	Mapping
	Endpoint() endpoint.Endpoint
	DecodeRequestFunc() httptransport.DecodeRequestFunc
	EncodeRequestFunc() httptransport.EncodeRequestFunc
	DecodeResponseFunc() httptransport.DecodeResponseFunc
	EncodeResponseFunc() httptransport.EncodeResponseFunc
}

// StaticMapping defines static assets mapping. e.g. javascripts, css, images, etc
type StaticMapping interface {
	Mapping
	StaticRoot() string
}

// TODO
// MvcMapping defines templated MVC mapping. e.g. html templates
// Templated MVC is usually implemented by Controller and produce a template and model for dynamic html generation
type MvcMapping interface {
	Mapping
	TemplateRoot() string
}
