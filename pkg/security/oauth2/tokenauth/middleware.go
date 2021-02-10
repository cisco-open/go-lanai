package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
)

const (
	bearerTokenPrefix = "Bearer "
	scopeParamPrefix = "scope."
)
/***********************
	Authorize Endpoint
 ***********************/
type TokenAuthMiddleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
	//TODO
}

type TokenAuthMWOptions func(opt *TokenAuthMWOption)

type TokenAuthMWOption struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
	//TODO
}

func NewTokenAuthMiddleware(opts...TokenAuthMWOptions) *TokenAuthMiddleware {
	opt := TokenAuthMWOption{
		//TODO
	}
	for _, optFunc := range opts {
		if optFunc != nil {
			optFunc(&opt)
		}
	}
	return &TokenAuthMiddleware{
		authenticator: opt.Authenticator,
		successHandler: opt.SuccessHandler,
		// TODO
	}
}

func (mw *TokenAuthMiddleware) AuthenticateHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		before := security.Get(ctx)
		if before != nil &&  before.State() >= security.StatePrincipalKnown {
			// this should not happen if the configuration is correct.
			// if session is enabled, this could happen. We always re-authenticate by clearing current auth
			security.Clear(ctx)
		}

		// grab bearer token and create candidate
		header := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerTokenPrefix) {
			mw.handleError(ctx, oauth2.NewInvalidAccessTokenError("missing bearer token"))
			return
		}

		tokenValue := strings.TrimPrefix(header, bearerTokenPrefix)
		candidate := BearerToken{
			Token:      tokenValue,
			DetailsMap: map[string]interface{}{},
		}

		// Authenticate
		auth, err := mw.authenticator.Authenticate(ctx, &candidate)
		if err != nil {
			mw.handleError(ctx, err)
			return
		}
		mw.handleSuccess(ctx, before, auth)
	}
}

func (mw *TokenAuthMiddleware) handleSuccess(c *gin.Context, before, new security.Authentication) {
	if new != nil {
		c.Set(gin.AuthUserKey, new.Principal())
		c.Set(security.ContextKeySecurity, new)
	}

	mw.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, new)
	// we don't explicitly write any thig on success
}

func (mw *TokenAuthMiddleware) handleError(c *gin.Context, err error) {
	if !errors.Is(err, oauth2.ErrorTypeOAuth2) {
		err = oauth2.NewInvalidAccessTokenError(err.Error(), err)
	}

	security.Clear(c)
	_ = c.Error(err)
	c.Abort()
}

