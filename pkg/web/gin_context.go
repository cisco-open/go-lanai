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

// SimpleGinMapping simple mapping of gin.HandlerFunc
type SimpleGinMapping interface {
	SimpleMapping
	GinHandlerFunc() gin.HandlerFunc
}

type MiddlewareGinMapping interface {
	MiddlewareMapping
	GinHandlerFunc() gin.HandlerFunc
}

/**************************
	Public
 **************************/

// GinContext returns *gin.Context which either contained in the context or is the given context itself
func GinContext(ctx context.Context) *gin.Context {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx
	}

	if ginCtx, ok := ctx.Value(gin.ContextKey).(*gin.Context); ok {
		return ginCtx
	}
	return nil
}

// MustGinContext returns *gin.Context like GinContext but panic if not found
func MustGinContext(ctx context.Context) *gin.Context {
	if gc := GinContext(ctx); gc != nil {
		return gc
	}
	panic(fmt.Sprintf("gin.Context is not found in given context %v", ctx))
}

// HttpRequest returns *http.Request associated with given context
func HttpRequest(ctx context.Context) *http.Request {
	if gc := GinContext(ctx); gc != nil {
		return gc.Request
	}
	return nil
}

// MustHttpRequest returns *http.Request associated with given context, panic if not found
func MustHttpRequest(ctx context.Context) *http.Request {
	return MustGinContext(ctx).Request
}

// SetKV set a kv pair to given context.
// The value can be obtained via context.Context.Value(key)
// - When Key is string, put it in gin.Context
// - When Key is not string, try MutableKVContext
// - Fallback to use fmt.Sprintf(`%v`, key) as key and put it in gin.Context
func SetKV(ctx context.Context, key interface{}, value interface{}) {
	if strKey, ok := key.(string); ok {
		switch c := ctx.(type) {
		case utils.MutableContext:
			c.Set(strKey, value)
		}
		return
	}
	if c, ok := ctx.(utils.ExtendedMutableContext); ok {
		c.SetKV(key, value)
		return
	}
	// fallback
	if gc := GinContext(ctx); gc != nil {
		gc.Set(fmt.Sprintf(`%v`, key), value)
	}
}

/**************************
	Customizers
 **************************/

// PriorityGinContextCustomizer implements Customizer and order.PriorityOrdered
type PriorityGinContextCustomizer struct {
	properties *ServerProperties
}

func NewPriorityGinContextCustomizer(properties *ServerProperties) *PriorityGinContextCustomizer {
	return &PriorityGinContextCustomizer{
		properties: properties,
	}
}

func (c PriorityGinContextCustomizer) PriorityOrder() int {
	// medium precedence, makes this customizer before any non-priority-ordered customizers
	return 0
}

func (c PriorityGinContextCustomizer) Customize(_ context.Context, r *Registrar) error {
	return r.AddGlobalMiddlewares(GinContextPathAware(c.properties))
}

// GinContextCustomizer implements Customizer and order.Ordered
type GinContextCustomizer struct {
	properties *ServerProperties
}

func NewGinContextCustomizer(properties *ServerProperties) *GinContextCustomizer {
	return &GinContextCustomizer{
		properties: properties,
	}
}

func (c GinContextCustomizer) Order() int {
	// medium precedence, makes this customizer before any non-ordered customizers
	return 0
}

func (c GinContextCustomizer) Customize(_ context.Context, r *Registrar) error {
	if e := r.AddGlobalMiddlewares(GinContextMerger()); e != nil {
		return e
	}
	return r.AddEngineOptions(func(engine *Engine) {
		engine.ContextWithFallback = true
	})
}

/**************************
	Handler Funcs
 **************************/

// GinContextPathAware is a Gin middleware mandatory for all mappings.
// It save the context path into context. The context path can be used in many components/utilities.
func GinContextPathAware(props *ServerProperties) gin.HandlerFunc {
	return func(gc *gin.Context) {
		gc.Set(ContextKeyContextPath, props.ContextPath)
	}
}

// GinContextMerger is a Gin middleware that merge Request.Context() with gin.Context,
// allowing values in gin.Context also available via Request.Context().Value().
// This middleware is mandatory for all mappings.
// Note:	as of Gin 1.8.0, if we set gin.Engine.ContextWithFallback to true. This makes gin.Context fully integrated
// 			with its underling Request.Context(). The side effect of this is gin.Context.Value() is also calling
// 			Request.Context().Value(), which cause stack overflow on non-existing keys.
//
//			To break this loop, we use different version of utils.ContextValuer to extract values from gin.Context(),
//			without using gin.Context.Value() function.
func GinContextMerger() gin.HandlerFunc {
	return func(gc *gin.Context) {
		ctx := utils.MakeMutableContext(gc.Request.Context(), ginContextValuer(gc))
		// Note, this is optional since Gin 1.8.0. We are doing this simply for performance
		ctx.Set(gin.ContextKey, gc)
		gc.Request = gc.Request.WithContext(ctx)
	}
}

// NewHttpGinHandlerFunc Integrate http.HandlerFunc with GIN handler
func NewHttpGinHandlerFunc(handlerFunc http.HandlerFunc) gin.HandlerFunc {
	if handlerFunc == nil {
		panic(fmt.Errorf("cannot wrap a nil hanlder"))
	}

	handler := func(c *gin.Context) {
		c = preProcessGinContext(c)
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
		c = preProcessGinContext(c)
		s.ServeHTTP(c.Writer, c.Request)
	}
	return handler
}

func preProcessGinContext(gc *gin.Context) *gin.Context {
	// because of GinContextMerger is mandatory middleware for all mappings, we are sure gc.Request.Context() contains gin.Context.
	// So we only need to make sure it's also mutable
	rc := gc.Request.Context()
	ctx := utils.MakeMutableContext(rc)
	if ctx != rc {
		gc.Request = gc.Request.WithContext(ctx)
	}
	// note, we could also make a copy of gin context in case we want to use it out of request scope
	// but currently, we don't have such requirement
	return gc
}

/**************************
	go-kit options
 **************************/
// integrateGinContextBefore Makes sure the context sent to go-kit's encoders/decoders/endpoints/errorHandlers
// contains values stored in gin.Context
func integrateGinContextBefore(ctx context.Context, _ *http.Request) (ret context.Context) {
	ret = utils.MakeMutableContext(ctx)
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

/**************************
	helpers
 **************************/

func ginContextValuer(gc *gin.Context) func(key interface{}) interface{} {
	return func(key interface{}) interface{} {
		switch strKey, ok := key.(string); ok {
		case strKey == gin.ContextKey:
			return gc
		default:
			v, _ := gc.Get(strKey)
			return v
		}
	}
}

/*********************************
	SimpleGinMapping
 *********************************/

// simpleGinMapping implements SimpleGinMapping
type simpleGinMapping struct {
	simpleMapping
	handlerFunc gin.HandlerFunc
}

func NewSimpleGinMapping(name, group, path, method string, condition RequestMatcher, handlerFunc gin.HandlerFunc) SimpleGinMapping {
	return &simpleGinMapping{
		simpleMapping: *NewSimpleMapping(name, group, path, method, condition, nil).(*simpleMapping),
		handlerFunc:   handlerFunc,
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

// middlewareGinMapping implements MiddlewareGinMapping
type middlewareGinMapping struct {
	middlewareMapping
	handlerFunc gin.HandlerFunc
}

func NewMiddlewareGinMapping(name string, order int, matcher RouteMatcher, cond RequestMatcher, handlerFunc gin.HandlerFunc) MiddlewareGinMapping {
	return &middlewareGinMapping{
		middlewareMapping: *NewMiddlewareMapping(name, order, matcher, cond, nil).(*middlewareMapping),
		handlerFunc:       handlerFunc,
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
