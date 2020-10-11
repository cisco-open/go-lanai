package rest

import (
	"errors"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

type MappingBuilder interface {
	Name(string) MappingBuilder
	Path(string) MappingBuilder
	Method(string) MappingBuilder
	EndpointFunc(EndpointFunc) MappingBuilder

	// convenient functions
	Get(string) MappingBuilder
	Post(string) MappingBuilder
	Put(string) MappingBuilder
	Patch(string) MappingBuilder
	Delete(string) MappingBuilder
	Options(string) MappingBuilder
	Head(string) MappingBuilder

	// Overrides
	Endpoint(endpoint.Endpoint) MappingBuilder
	DecodeRequestFunc(httptransport.DecodeRequestFunc) MappingBuilder
	EncodeRequestFunc(httptransport.EncodeRequestFunc) MappingBuilder
	DecodeResponseFunc(httptransport.DecodeResponseFunc) MappingBuilder
	EncodeResponseFunc(httptransport.EncodeResponseFunc) MappingBuilder

	// Builder
	Build() *endpointMapping
}

type mappingBuilder struct {
	name 		 string
	path         string
	method       string
	endpointFunc EndpointFunc
	endpoint           endpoint.Endpoint
	decodeRequestFunc  httptransport.DecodeRequestFunc
	encodeRequestFunc  httptransport.EncodeRequestFunc
	decodeResponseFunc httptransport.DecodeResponseFunc
	encodeResponseFunc httptransport.EncodeResponseFunc
}

func NewBuilder(names ...string) MappingBuilder {
	name := "unknown"
	if len(names) > 0 {
		name = names[0]
	}
	return &mappingBuilder{
		name: name,
	}
}

/*****************************
	MappingBuilder Impl
******************************/
func (b *mappingBuilder) Name(name string) MappingBuilder {
	b.name = name
	return b
}
func (b *mappingBuilder) Path(path string) MappingBuilder {
	b.path = path
	return b
}

func (b *mappingBuilder) Method(method string) MappingBuilder {
	b.method = method
	return b
}

func (b *mappingBuilder) EndpointFunc(endpointFunc EndpointFunc) MappingBuilder {
	b.endpointFunc = endpointFunc
	return b
}

// Convenient setters
func (b *mappingBuilder) Get(path string) MappingBuilder {
	return b.Path(path).Method(http.MethodGet)
}

func (b *mappingBuilder) Post(path string) MappingBuilder {
	return b.Path(path).Method(http.MethodPost)
}

func (b *mappingBuilder) Put(path string) MappingBuilder {
	return b.Path(path).Method(http.MethodPut)
}

func (b *mappingBuilder) Patch(path string) MappingBuilder {
	return b.Path(path).Method(http.MethodPatch)
}

func (b *mappingBuilder) Delete(path string) MappingBuilder {
	return b.Path(path).Method(http.MethodDelete)
}

func (b *mappingBuilder) Options(path string) MappingBuilder {
	return b.Path(path).Method(http.MethodOptions)
}

func (b *mappingBuilder) Head(path string) MappingBuilder {
	return b.Path(path).Method(http.MethodHead)
}

// Overrides
func (b *mappingBuilder) Endpoint(endpoint endpoint.Endpoint) MappingBuilder {
	b.endpoint = endpoint
	return b
}

func (b *mappingBuilder) DecodeRequestFunc(f httptransport.DecodeRequestFunc) MappingBuilder {
	b.decodeRequestFunc = f
	return b
}

func (b *mappingBuilder) EncodeRequestFunc(f httptransport.EncodeRequestFunc) MappingBuilder {
	b.encodeRequestFunc = f
	return b
}

func (b *mappingBuilder) DecodeResponseFunc(f httptransport.DecodeResponseFunc) MappingBuilder {
	b.decodeResponseFunc = f
	return b
}

func (b *mappingBuilder) EncodeResponseFunc(f httptransport.EncodeResponseFunc) MappingBuilder {
	b.encodeResponseFunc = f
	return b
}

func (b *mappingBuilder) Build() *endpointMapping {
	if err := b.validate(); err != nil {
		panic(err)
	}
	return b.buildMapping()
}

// TODO more validation and better error handling
func (b *mappingBuilder) validate() (err error) {
	if b.path == "" || b.method == "" {
		err = errors.New("empty Path")
	}
	return
}

func (b *mappingBuilder) buildMapping() *endpointMapping {
	m := &endpointMapping {
		name: b.name,
		path: b.path,
		method: b.method,
		endpointFunc: b.endpointFunc,
		endpoint: nil,
		decodeRequestFunc: httptransport.NopRequestDecoder,
		encodeRequestFunc: GenericEncodeRequestFunc,
		decodeResponseFunc: nil, // TODO
		encodeResponseFunc: GenericEncodeResponseFunc,
	}

	if b.endpointFunc != nil {
		metadata := MakeEndpointFuncMetadata(b.endpointFunc)
		m.endpoint = MakeEndpoint(metadata)
		m.decodeRequestFunc = MakeGinBindingDecodeRequestFunc(metadata)
		m.decodeResponseFunc = nil // TODO
	}

	b.customize(m)
	return m
}

func (b *mappingBuilder) customize(m *endpointMapping) {
	if b.endpoint != nil {
		m.endpoint = b.endpoint
	}

	if b.encodeRequestFunc != nil {
		m.encodeRequestFunc = b.encodeRequestFunc
	}

	if b.decodeRequestFunc != nil {
		m.decodeRequestFunc = b.decodeRequestFunc
	}

	if b.encodeResponseFunc != nil {
		m.encodeResponseFunc = b.encodeResponseFunc
	}

	if b.decodeResponseFunc != nil {
		m.decodeResponseFunc = b.decodeResponseFunc
	}
}
