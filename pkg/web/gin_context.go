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

/**************************
	Customizer
 **************************/
// RecoveryCustomizer implements Customizer and order.Ordered
type GinContextCustomizer struct {}

func NewGinContextCustomizer() *GinContextCustomizer {
	return &GinContextCustomizer{}
}

func (c GinContextCustomizer) Order() int {
	// medium precedence, makes this customizer before any non-ordered customizers
	return 0
}

func (c GinContextCustomizer) Customize(ctx context.Context, r *Registrar) error {
	r.AddGlobalMiddlewares(GinContextMerger())
	return nil
}

/**************************
	Handler Funcs
 **************************/
func GinContextMerger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.MakeMutableContext(c.Request.Context(), ginContextValuer(c))
		ctx.Set(kGinContextKey, c)
		c.Request = c.Request.WithContext(ctx)
	}
}

// HttpGinHandlerFunc Integrate http.HandlerFunc with GIN handler
func NewHttpGinHandlerFunc(handlerFunc http.HandlerFunc) gin.HandlerFunc {
	if handlerFunc == nil {
		panic(fmt.Errorf("cannot wrap a nil hanlder"))
	}

	handler := func(c *gin.Context) {
		c = preProcessContext(c)
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
		c = preProcessContext(c)
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
func integrateGinContextBefore(ctx context.Context, r *http.Request) (ret context.Context) {
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
