package web

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MWConditionFunc func(context.Context, *http.Request) bool

// MWConditionMatcher accepts *http.Request or http.Request
type MWConditionMatcher RequestMatcher

type ConditionalMiddleware interface {
	ConditionFunc() MWConditionFunc
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



