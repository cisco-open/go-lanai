package web

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/gin-gonic/gin"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

// Functions, HandlerFuncs and go-kit ServerOptions that make sure *gin.Context is availalble in endpoints and
// context is properly propagated in Request

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

// MakeGinHandlerFunc Integrate go-kit Server with GIN handler
func MakeGinHandlerFunc(s *httptransport.Server, rm RequestMatcher) gin.HandlerFunc {
	handler := func(c *gin.Context) {
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), kGinContextKey, c),
		)
		s.ServeHTTP(c.Writer, c.Request)
	}
	return MakeConditionalHandlerFunc(handler, rm)
}

// MakeConditionalHandlerFunc wraps given handler with a request matcher
func MakeConditionalHandlerFunc(handler gin.HandlerFunc, rm RequestMatcher) gin.HandlerFunc {
	if rm == nil {
		return handler
	}
	return func(c *gin.Context) {
		if matches, e := rm.MatchesWithContext(c, c.Request); e == nil && matches {
			handler(c)
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