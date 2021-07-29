package token

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"errors"
	"github.com/gin-gonic/gin"
)

/***********************
	Token Endpoint
 ***********************/

//goland:noinspection GoNameStartsWithPackageName
type TokenEndpointMiddleware struct {
	granter     auth.TokenGranter
}

//goland:noinspection GoNameStartsWithPackageName
type TokenEndpointOptionsFunc func(*TokenEndpointOptions)

//goland:noinspection GoNameStartsWithPackageName
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
		// first we double check if client is authenticated
		client := auth.RetrieveAuthenticatedClient(ctx)
		if client == nil {
			mw.handleError(ctx, oauth2.NewClientNotFoundError("invalid client"))
			return
		}

		// parse request
		tokenRequest, e := auth.ParseTokenRequest(ctx.Request)
		if e != nil {
			mw.handleError(ctx, oauth2.NewInvalidTokenRequestError("invalid token request", e))
			return
		}

		// see if client id matches
		if tokenRequest.ClientId != "" && tokenRequest.ClientId != client.ClientId() {
			mw.handleError(ctx, oauth2.NewInvalidTokenRequestError("given client Domain does not match authenticated client"))
			return
		}
		tokenRequest.Extensions[oauth2.ExtUseSessionTimeout] = client.UseSessionTimeout()

		// check grant
		if e := auth.ValidateGrant(ctx, client, tokenRequest.GrantType); e != nil {
			mw.handleError(ctx, e)
			return
		}

		// check if supported
		if tokenRequest.GrantType == oauth2.GrantTypeImplicit {
			mw.handleError(ctx, oauth2.NewInvalidGrantError("implicit grant type not supported from token endpoint"))
			return
		}

		token, e := mw.granter.Grant(ctx, tokenRequest)
		if e != nil {
			mw.handleError(ctx, e)
			return
		}

		mw.handleSuccess(ctx, token)
	}
}

func (mw *TokenEndpointMiddleware) handleSuccess(c *gin.Context, v interface{}) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(200, v)
	c.Abort()
}

func (mw *TokenEndpointMiddleware) handleError(c *gin.Context, err error) {
	if errors.Is(err, oauth2.ErrorTypeOAuth2) {
		err = oauth2.NewInvalidGrantError(err)
	}

	_ = c.Error(err)
	c.Abort()
}