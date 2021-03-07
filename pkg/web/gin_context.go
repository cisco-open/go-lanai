package web

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

// Functions, HandlerFuncs and go-kit ServerOptions that make sure *gin.Context is availalble in endpoints and
// context is properly propagated in Request

// SimpleGinMapping
type SimpleGinMapping interface {
	SimpleMapping
	GinHandlerFunc() gin.HandlerFunc
}

type MiddlewareGinMapping interface {
	MiddlewareMapping
	GinHandlerFunc() gin.HandlerFunc
}

// GinContext returns *gin.Context which either contained in the context or is the given context itself
func GinContext(ctx context.Context) *gin.Context {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx
	}

	if ginCtx, ok := ctx.Value(kGinContextKey).(*gin.Context); ok {
		return ginCtx
	}

	return nil
}

// HttpGinHandlerFunc Integrate http.HandlerFunc with GIN handler
func NewHttpGinHandlerFunc(handlerFunc http.HandlerFunc) gin.HandlerFunc {
	if handlerFunc == nil {
		panic(fmt.Errorf("cannot wrap a nil hanlder"))
	}

	handler := func(c *gin.Context) {
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), kGinContextKey, c),
		)
		handlerFunc(c.Writer, c.Request)
	}
	return handler
}

// NewKitGinHandlerFunc Integrate go-kit Server with GIN handler
func NewKitGinHandlerFunc(s *httptransport.Server) gin.HandlerFunc {
	if s == nil {
		panic(fmt.Errorf("cannot wrap a nil hanlder"))
	}

	handler := func(c *gin.Context) {
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), kGinContextKey, c),
		)
		s.ServeHTTP(c.Writer, c.Request)
	}
	return handler
}

// makeGinConditionalHandlerFunc wraps given handler with a request matcher
func makeGinConditionalHandlerFunc(handler gin.HandlerFunc, rm RequestMatcher) gin.HandlerFunc {
	if rm == nil {
		return handler
	}
	return func(c *gin.Context) {
		if matches, e := rm.MatchesWithContext(c, c.Request); e == nil && matches {
			handler(c)
		} else if e != nil {
			c.Error(e)
			c.Abort()
		}
	}
}

// integrateGinContextBefore Makes sure the context sent to go-kit's encoders/decoders/endpoints/errorHandlers
// contains values stored in gin.Context
func integrateGinContextBefore(ctx context.Context, r *http.Request) (ret context.Context) {
	if ginCtx := GinContext(ctx); ginCtx != nil {
		ret = utils.MakeMutableContext(ctx, func(key interface{}) interface{} {
			return ginCtx.Value(key)
		})
	} else {
		ret = utils.MakeMutableContext(ctx)
	}

	return
}

// integrateGinContextFinalizer Makes sure the context processed by go-kit is set back to request,
// whose value would becomes accessible outside the go-kit realm
func integrateGinContextFinalizer(ctx context.Context, _ int, r *http.Request) {
	gc := GinContext(ctx)
	if gc == nil {
		return
	}
	// updates Request with final ctx
	// Note:
	// 	this update is important, because when the execution flow exit go-kit realm, all information stored in ctx
	//	would be lost if we don't set it to Request
	gc.Request = r.WithContext(ctx)
}

/*********************************
	SimpleGinMapping
 *********************************/
// implmenets SimpleGinMapping
type simpleGinMapping struct {
	simpleMapping
	handlerFunc gin.HandlerFunc
}

func NewSimpleGinMapping(name, path, method string, condition RequestMatcher, handlerFunc gin.HandlerFunc) *simpleGinMapping {
	return &simpleGinMapping{
		simpleMapping: *NewSimpleMapping(name, path, method, condition, nil).(*simpleMapping),
		handlerFunc: handlerFunc,
	}
}

func (m simpleGinMapping) GinHandlerFunc() gin.HandlerFunc {
	if m.handlerFunc != nil {
		return m.handlerFunc
	}

	if m.simpleMapping.handlerFunc != nil {
		return NewHttpGinHandlerFunc(http.HandlerFunc(m.simpleMapping.handlerFunc))
	}
	return nil
}

/*********************************
	MiddlewareGinMapping
 *********************************/
// implmenets MiddlewareGinMapping
type middlewareGinMapping struct {
	middlewareMapping
	handlerFunc        gin.HandlerFunc
}

func NewMiddlewareGinMapping(name string, order int, matcher RouteMatcher, cond RequestMatcher, handlerFunc gin.HandlerFunc) *middlewareGinMapping {
	return &middlewareGinMapping{
		middlewareMapping: *NewMiddlewareMapping(name, order, matcher, cond, nil).(*middlewareMapping),
		handlerFunc: handlerFunc,
	}
}

func (m middlewareGinMapping) GinHandlerFunc() gin.HandlerFunc {
	if m.handlerFunc != nil {
		return m.handlerFunc
	}

	if m.middlewareMapping.handlerFunc != nil {
		return NewHttpGinHandlerFunc(http.HandlerFunc(m.middlewareMapping.handlerFunc))
	}
	return nil
}
