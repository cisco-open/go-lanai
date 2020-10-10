package web

import (
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type MappingFunc func(registrar *Registrar)

type Mapping interface {
	Path() string
	Method() string
}

type EndpointMapping interface {
	Mapping
	Endpoint() endpoint.Endpoint
	DecodeRequestFunc() httptransport.DecodeRequestFunc
	EncodeRequestFunc() httptransport.EncodeRequestFunc
	DecodeResponseFunc() httptransport.DecodeResponseFunc
	EncodeResponseFunc() httptransport.EncodeResponseFunc
}

// TODO for static resource mapping
type StaticMapping interface {
	Mapping
	StaticFile() string
}
