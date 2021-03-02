package mapping

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*********************************
	SimpleMappingBuilder
 *********************************/
type MappingBuilder struct {
	name        string
	path        string
	method      string
	condition   web.RequestMatcher
	handlerFunc gin.HandlerFunc
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

func Put(path string) *MappingBuilder {
	return New().Put(path)
}

func Patch(path string) *MappingBuilder {
	return New().Patch(path)
}

func Delete(path string) *MappingBuilder {
	return New().Delete(path)
}

func Options(path string) *MappingBuilder {
	return New().Options(path)
}

func Head(path string) *MappingBuilder {
	return New().Head(path)
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

func (b *MappingBuilder) HandlerFunc(handlerFunc gin.HandlerFunc) *MappingBuilder {
	b.handlerFunc = handlerFunc
	return b
}

// Convenient setters
func (b *MappingBuilder) Get(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodGet)
}

func (b *MappingBuilder) Post(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodPost)
}

func (b *MappingBuilder) Put(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodPut)
}

func (b *MappingBuilder) Patch(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodPatch)
}

func (b *MappingBuilder) Delete(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodDelete)
}

func (b *MappingBuilder) Options(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodOptions)
}

func (b *MappingBuilder) Head(path string) *MappingBuilder {
	return b.Path(path).Method(http.MethodHead)
}

func (b *MappingBuilder) Build() web.SimpleMapping {
	if err := b.validate(); err != nil {
		panic(err)
	}
	return b.buildMapping()
}

/*****************************
	Getters
******************************/
func (b *MappingBuilder) GetPath() string {
	return b.path
}

func (b *MappingBuilder) GetMethod() string {
	return b.method
}

func (b *MappingBuilder) GetCondition() web.RequestMatcher {
	return b.condition
}

func (b *MappingBuilder) GetName() string {
	return b.name
}

/*****************************
	Private
******************************/
func (b *MappingBuilder) validate() (err error) {
	switch {
	case b.path == "":
		err = errors.New("empty path")
	case b.handlerFunc == nil:
		err = errors.New("handler func not specified")
	}
	return
}

func (b *MappingBuilder) buildMapping() web.SimpleMapping {
	if b.method == "" {
		b.method = web.MethodAny
	}

	if b.name == "" {
		b.name = fmt.Sprintf("%s %s", b.method, b.path)
	}
	return web.NewSimpleMapping(b.name, b.path, b.method, b.condition, b.handlerFunc)
}

