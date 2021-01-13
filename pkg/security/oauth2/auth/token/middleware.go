package token

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"github.com/gin-gonic/gin"
)

/***********************
	Token Endpoint
 ***********************/
type TokenEndpointMiddleware struct {
	granter     auth.TokenGranter
	//TODO
}

type TokenEndpointOptionsFunc func(*TokenEndpointOptions)

type TokenEndpointOptions struct {
	Granter     *auth.CompositeTokenGranter
}

func NewTokenEndpointMiddleware(optionFuncs...TokenEndpointOptionsFunc) *TokenEndpointMiddleware {
	opts := TokenEndpointOptions{
		Granter: auth.NewCompositeTokenGranter(),
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
		tokenReuqest, e := auth.ParseTokenRequest(ctx.Request)
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