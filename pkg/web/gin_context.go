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

	if ginCtx, ok := ctx.Value(kGinContextKey).(*gin.Context); ok {
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

// SetKV set a kv pair to given context. The value can be obtained via context.Context.Value(key)
func SetKV(ctx context.Context, key string, value interface{}) {
	switch c := ctx.(type) {
	case utils.MutableContext:
		c.Set(key, value)
	default:
		if gc := GinContext(ctx); gc != nil {
			gc.Set(key, value)
		}
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
	return r.AddGlobalMiddlewares(GinContextMerger())
}

/**************************
	Handler Funcs
 **************************/

func GinContextPathAware(props *ServerProperties) gin.HandlerFunc {
	return func(gc *gin.Context) {
		gc.Set(ContextKeyContextPath, props.ContextPath)
	}
}

func GinContextMerger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.MakeMutableContext(c.Request.Context(), ginContextValuer(c))
		ctx.Set(kGinContextKey, c)
		c.Request = c.Request.WithContext(ctx)
	}
}

// NewHttpGinHandlerFunc Integrate http.HandlerFunc with GIN handler
func NewHttpGinHandlerFunc(handlerFunc http.HandlerFunc) gin.HandlerFunc {
	if handlerFunc == nil {
		panic(fmt.Errorf("cannot wrap a nil hanlder"))
	}

	handler := func(c *gin.Context) {
		c = preProcessContext(c) //nolint:contextcheck // gin.Context is the context
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
		c = preProcessContext(c) //nolint:contextcheck // gin.Context is the context
		s.ServeHTTP(c.Writer, c.Request)
	}
	return handler
}

func preProcessContext(c *gin.Context) *gin.Context{
	// because of GinContextMerger is manditory middleware for all mappings, we are sure c.Request.Context() contains gin.Context.
	// So we only need to make sure it's also mutable
	ctx := utils.MakeMutableContext(c.Request.Context())
	if ctx != c.Request.Context() {
		c.Request = c.Request.WithContext(ctx)
	}
	// note, we could aslo make a copy of gin context in case we want to use it out of reqeust scope
	// but currently, we don't have such requirement
	return c
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
func ginContextValuer(ginCtx *gin.Context) func(key interface{}) interface{} {
	return func(key interface{}) interface{} {
		return ginCtx.Value(key)
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

func NewSimpleGinMapping(name, group, path, method string, condition RequestMatcher, handlerFunc gin.HandlerFunc) *simpleGinMapping {
	return &simpleGinMapping{
		simpleMapping: *NewSimpleMapping(name, group, path, method, condition, nil).(*simpleMapping),
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

// middlewareGinMapping implements MiddlewareGinMapping
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
