package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ConditionalMiddlewareFunc func(*http.Request) bool

type ConditionalMiddleware interface {
	ConditionFunc() ConditionalMiddlewareFunc
}

type Middleware interface {
	HandlerFunc() gin.HandlerFunc
}

type middlewareMapping struct {
	name               string
	order              int
	matcher            RouteMatcher
	handlerFunc        gin.HandlerFunc
}

func NewMiddlewareMapping(name string, order int, matcher RouteMatcher, handlerFunc gin.HandlerFunc) MiddlewareMapping {
	return &middlewareMapping {
		name: name,
		matcher: matcher,
		order: order,
		handlerFunc: handlerFunc,
	}
}

func (mm *middlewareMapping) Name() string {
	return mm.name
}

func (mm *middlewareMapping) Matcher() RouteMatcher {
	return mm.matcher
}

func (mm *middlewareMapping) Order() int {
	return mm.order
}

func (mm *middlewareMapping) HandlerFunc() gin.HandlerFunc {
	return mm.handlerFunc
}

