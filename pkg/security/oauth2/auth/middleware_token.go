package auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

/***********************
	Token Endpoint
 ***********************/
type TokenEndpointMiddleware struct {
	granter TokenGranter
	//TODO
}

type TokenEndpointOptionsFunc func(*TokenEndpointOptions)

type TokenEndpointOptions struct {
	Granter *CompositeTokenGranter
}

func NewTokenEndpointMiddleware(optionFuncs...TokenEndpointOptionsFunc) *TokenEndpointMiddleware {
	opts := TokenEndpointOptions{
		Granter: NewCompositeTokenGranter(),
	}
	for _, optFunc := range optionFuncs {
		if optFunc != nil {
			optFunc(&opts)
		}
	}
	return &TokenEndpointMiddleware{
		granter: opts.Granter,
	}
}

func (mw *TokenEndpointMiddleware) TokenHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO
		tokenReuqest, e := parseTokenRequest(ctx.Request)
		if e != nil {
			mw.handleError(ctx, oauth2.NewInvalidTokenRequestError("invalid token request", e))
			return
		}

		token, e := mw.granter.Grant(ctx, tokenReuqest)
		if e != nil {
			mw.handleError(ctx, e)
			return
		}

		mw.handleSuccess(ctx, token)
	}
}

func (mw *TokenEndpointMiddleware) EndpointHandlerFunc() gin.HandlerFunc {
	return notFoundHandlerFunc
}

// TODO
func (mw *TokenEndpointMiddleware) handleSuccess(c *gin.Context, v interface{}) {
	c.JSON(200, v)
	c.Abort()
}

// TODO
func (mw *TokenEndpointMiddleware) handleError(c *gin.Context, err error) {
	security.Clear(c)
	_ = c.Error(err)
	c.Abort()
}

func notFoundHandlerFunc(c *gin.Context) {
	_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("page not found"))
}