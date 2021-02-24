package template

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"fmt"
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
	name        string
	path        string
	method      string
	condition   web.RequestMatcher
	handlerFunc ModelViewHandlerFunc
}

func New(names ...string) *MappingBuilder {
	var name string
	if len(names) > 0 {
		name = names[0]
	}
	return &MappingBuilder{
		name: name,
		method: web.MethodAny,
	}
}

// Convenient Constructors
func Any(path string) *MappingBuilder {
	return New().Path(path).Method(web.MethodAny)
}

func Get(path string) *MappingBuilder {
	return New().Get(path)
}

func Post(path string) *MappingBuilder {
	return New().Post(path)
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

func (b *MappingBuilder) Condition(condition web.RequestMatcher) *MappingBuilder {
	b.condition = condition
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
	if b.path == "" {
		err = errors.New("empty Path")
	}

	if b.handlerFunc == nil {
		err = errors.New("handler func is required for template mapping")
	}
	return
}

func (b *MappingBuilder) buildMapping() web.MvcMapping {
	if b.method == "" {
		b.method = web.MethodAny
	}

	if b.name == "" {
		b.name = fmt.Sprintf("%s %s", b.method, b.path)
	}

	// For templated HTMLs, it's usually browser-to-service communication
	// Since we don't usually need to do service-to-service communication,
	//we don't need to apply init request encoder and response decoder
	var e           endpoint.Endpoint
	var decodeRequestFunc = httptransport.NopRequestDecoder
	var encodeResponseFunc = TemplateEncodeResponseFunc

	if b.handlerFunc != nil {
		metadata := web.MakeFuncMetadata(b.handlerFunc, validateHandlerFunc)
		e = web.MakeEndpoint(metadata)
		decodeRequestFunc = web.MakeGinBindingDecodeRequestFunc(metadata)
	}

	return web.NewMvcMapping(b.name, b.path, b.method, b.condition,
		e, decodeRequestFunc, nil,
		nil, encodeResponseFunc,
		TemplateErrorEncoder)
}

// this is an additional validator, we assume basic validation is done (meaning given value web.MvcHandlerFunc)
func validateHandlerFunc(f *reflect.Value) error {
	t := f.Type()
	// check response type
	foundMV := false
	for i := t.NumOut() - 1; i >= 0; i-- {
		if t.Out(i).ConvertibleTo(reflect.TypeOf(&ModelView{})) {
			foundMV = true
			break
		}
	}

	switch {
	case !foundMV:
		return errors.New("ModelViewHandlerFunc need return ModelView or *ModelView")
	//more checks if needed
	}

	return nil
}
