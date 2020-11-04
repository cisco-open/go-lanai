package middleware

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"github.com/gin-gonic/gin"
)

type MappingBuilder struct {
	name       string
	middleware web.Middleware
	matcher    web.RouteMatcher
	order      int
	// overrides
	condition   web.ConditionalMiddlewareFunc
	handlerFunc gin.HandlerFunc
}

func NewBuilder(names ...string) *MappingBuilder {
	name := "unknown"
	if len(names) > 0 {
		name = names[0]
	}
	return &MappingBuilder{
		name: name,
		order: 0,
	}
}

/*****************************
	Public
******************************/
func (b *MappingBuilder) Name(name string) *MappingBuilder {
	b.name = name
	return b
}

func (b *MappingBuilder) Order(order int) *MappingBuilder {
	b.order = order
	return b
}

func (b *MappingBuilder) With(middleware web.Middleware) *MappingBuilder {
	b.middleware = middleware
	return b
}

func (b *MappingBuilder) ApplyTo(matcher web.RouteMatcher) *MappingBuilder {
	b.matcher = matcher
	return b
}

func (b *MappingBuilder) Use(handlerFunc gin.HandlerFunc) *MappingBuilder {
	b.handlerFunc = handlerFunc
	return b
}

func (b *MappingBuilder) WithCondition(condition web.ConditionalMiddlewareFunc) *MappingBuilder {
	b.condition = condition
	return b
}

func (b *MappingBuilder) Build() web.MiddlewareMapping {
	var conditionFunc web.ConditionalMiddlewareFunc
	var handlerFunc gin.HandlerFunc
	if b.middleware != nil {
		handlerFunc = b.middleware.HandlerFunc()
		if conditional,ok := b.middleware.(web.ConditionalMiddleware); ok {
			conditionFunc = conditional.ConditionFunc()
		}
	}

	if b.handlerFunc != nil {
		handlerFunc = b.handlerFunc
	}

	if b.condition != nil {
		conditionFunc = b.condition
	}

	if handlerFunc == nil {
		panic(fmt.Errorf("unable to build '%s' middleware mapping: missing HandlerFunc. please use With(...) or Use(...)", b.name))
	}

	return web.NewMiddlewareMapping(b.name, b.order, b.matcher, makeConditionalHandlerFunc(handlerFunc, conditionFunc))
}

func makeConditionalHandlerFunc(handlerFunc gin.HandlerFunc, conditionFunc web.ConditionalMiddlewareFunc) gin.HandlerFunc {
	if conditionFunc == nil {
		return handlerFunc
	}
	
	return func(c *gin.Context) {
		if conditionFunc(c.Request) {
			handlerFunc(c)
		}
	}
}

