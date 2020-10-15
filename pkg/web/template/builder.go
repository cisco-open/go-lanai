package template

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"errors"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	"reflect"
)

// ModelViewHandlerFunc is a function with following signature
// 	- two input parameters with 1st as context.Context and 2nd as <request>
// 	- two output parameters with 1st as <response> and 2nd as error
// where
// <request>:   a struct or a pointer to a struct whose fields are properly tagged
// <response>:  a pointer to a ModelView.
// e.g.: func(context.Context, request *AnyStructWithTag) (response *ModelView, error) {...}
type ModelViewHandlerFunc web.MvcHandlerFunc

type MappingBuilder struct {
	name               string
	path               string
	method             string
	handlerFunc       ModelViewHandlerFunc
	template 		   string
}

func NewBuilder(names ...string) *MappingBuilder {
	name := "unknown"
	if len(names) > 0 {
		name = names[0]
	}
	return &MappingBuilder{
		name: name,
	}
}

/*****************************
	Public
******************************/
func (b *MappingBuilder) Name(name string) *MappingBuilder {
	b.name = name
	return b
}
func (b *MappingBuilder) Path(path string) *MappingBuilder {
	b.path = path
	return b
}

func (b *MappingBuilder) Method(method string) *MappingBuilder {
	b.method = method
	return b
}

func (b *MappingBuilder) HandlerFunc(endpointFunc ModelViewHandlerFunc) *MappingBuilder {
	b.handlerFunc = endpointFunc
	return b
}

// Convenient setters
func (b *MappingBuilder) Get(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodGet)
}

func (b *MappingBuilder) Post(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodPost)
}

func (b *MappingBuilder) Build() web.TemplateMapping {
	if err := b.validate(); err != nil {
		panic(err)
	}
	return b.buildMapping()
}

/*****************************
	Private
******************************/
// TODO more validation and better error handling
func (b *MappingBuilder) validate() (err error) {
	if b.path == "" || b.method == "" {
		err = errors.New("empty Path")
	}

	if b.handlerFunc == nil {
		err = errors.New("handler func is required for template mapping")
	}
	return
}

func (b *MappingBuilder) buildMapping() web.MvcMapping {
	// For templated HTMLs, it's usually browser-to-service communication
	// Since we don't usually need to do service-to-service communication,
	//we don't need to config request encoder and response decoder
	var e           endpoint.Endpoint
	var decodeRequestFunc = httptransport.NopRequestDecoder
	var encodeResponseFunc = ginTemplateEncodeResponseFunc

	if b.handlerFunc != nil {
		metadata := web.MakeFuncMetadata(b.handlerFunc, validateHandlerFunc)
		e = web.MakeEndpoint(metadata)
		decodeRequestFunc = web.MakeGinBindingDecodeRequestFunc(metadata)
	}

	return web.NewMvcMapping(b.name, b.path, b.method,
		e, decodeRequestFunc, nil, nil, encodeResponseFunc)
}

// this is an additional validator, we assume basic validation is done (meaning given value web.MvcHandlerFunc)
func validateHandlerFunc(f *reflect.Value) error {
	t := f.Type().Out(0)
	// check response type
	if !t.ConvertibleTo(reflect.TypeOf(&ModelView{})) {
		return errors.New("ModelViewHandlerFunc need return ModelView or *ModelView as first parameter")
	}
	return nil
}
