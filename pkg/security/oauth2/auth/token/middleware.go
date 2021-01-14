package token

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
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
		// first we double check if client is authenticated
		client := mw.authenticatedClient(ctx)
		if client == nil {
			mw.handleError(ctx, oauth2.NewClientNotFoundError("invalid client"))
			return
		}

		// parse request
		tokenReuqest, e := auth.ParseTokenRequest(ctx.Request)
		if e != nil {
			mw.handleError(ctx, oauth2.NewInvalidTokenRequestError("invalid token request", e))
			return
		}

		// see if client id matches
		if tokenReuqest.ClientId != "" && tokenReuqest.ClientId != client.ClientId() {
			mw.handleError(ctx, oauth2.NewInvalidTokenRequestError("given client ID does not match authenticated client"))
			return
		}

		// check scope
		if e := mw.validateScope(ctx, tokenReuqest, client); e != nil {
			mw.handleError(ctx, e)
			return
		}

		// check grant
		if e := mw.validateGrant(ctx, tokenReuqest, client); e != nil {
			mw.handleError(ctx, e)
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

func (mw *TokenEndpointMiddleware) authenticatedClient(c context.Context) auth.OAuth2Client {
	sec := security.Get(c)
	if sec.State() < security.StateAuthenticated {
		return nil
	}

	if client, ok := sec.Principal().(auth.OAuth2Client); ok {
		return client

	}
	return nil
}

func (mw *TokenEndpointMiddleware) validateScope(c context.Context, req *auth.TokenRequest, client auth.OAuth2Client) error {
	for scope, _ := range req.Scopes {
		if !client.Scopes().Has(scope) {
			return oauth2.NewInvalidScopeError("invalid scope: " + scope)
		}
	}
	return nil
}

func (mw *TokenEndpointMiddleware) validateGrant(c context.Context, req *auth.TokenRequest, client auth.OAuth2Client) error {
	if req.GrantType == "" {
		return oauth2.NewInvalidTokenRequestError("missing grant_type")
	}

	if !client.GrantTypes().Has(req.GrantType) {
		return oauth2.NewInvalidGrantError(fmt.Sprintf("grant type '%s' is not allowed by this client '%s'", req.GrantType, client.ClientId()))
	}

	if req.GrantType == oauth2.GrantTypeImplicit {
		return oauth2.NewInvalidGrantError("implicit grant type not supported from token endpoint")
	}
	return nil
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