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

type endpoints struct {

}

//func (c *endpoints) Mappings() {
//	t := reflect.TypeOf(c)
//}