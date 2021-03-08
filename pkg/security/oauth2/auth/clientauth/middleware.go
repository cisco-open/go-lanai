package clientauth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ClientAuthMiddleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
}

type ClientAuthMWOptions func(*ClientAuthMWOption)

type ClientAuthMWOption struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
}

func NewClientAuthMiddleware(opts...ClientAuthMWOptions) *ClientAuthMiddleware {
	opt := ClientAuthMWOption{}

	for _, optFunc := range opts {
		if optFunc != nil {
			optFunc(&opt)
		}
	}
	return &ClientAuthMiddleware{
		authenticator:  opt.Authenticator,
		successHandler: opt.SuccessHandler,
	}
}

func (mw *ClientAuthMiddleware) ClientAuthFormHandlerFunc() web.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if e := r.ParseForm(); e != nil {
			return
		}

		_, hasClientId := r.PostForm[oauth2.ParameterClientId]
		if !hasClientId {
			return
		}

		clientId := r.PostForm.Get(oauth2.ParameterClientId)

		// form client auth should be placed after basic auth.
		// if already authenticated by basic auth and pricipal matches, we don't need to do anything here
		// if authenticated but pricipal doesn't match, it's an error
		before := security.Get(r.Context())
		currentAuth, ok := before.(passwd.UsernamePasswordAuthentication)
		switch {
		case ok && passwd.IsSamePrincipal(clientId, currentAuth):
			return
		case ok:
			mw.handleError(r.Context(), oauth2.NewInvalidClientError("client_id parameter and Authorization header doesn't match"))
		}

		secret := r.PostForm.Get(oauth2.ParameterClientSecret)
		candidate := passwd.UsernamePasswordPair{
			Username: clientId,
			Password: secret,
			EnforceMFA: passwd.MFAModeSkip,
		}
		// Authenticate
		auth, err := mw.authenticator.Authenticate(r.Context(), &candidate)
		if err != nil {
			mw.handleError(r.Context(), err)
			return
		}
		mw.handleSuccess(r.Context(), r, rw, before, auth)
	}
}

func (mw *ClientAuthMiddleware) ErrorTranslationHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// find first authentication error and translate it
		for _, e := range c.Errors {
			switch {
			case errors.Is(e.Err, security.ErrorTypeAuthentication):
				e.Err = oauth2.NewInvalidClientError("client authentication failed", e.Err)
			}
		}
	}
}

func (mw *ClientAuthMiddleware) handleSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, before, new security.Authentication) {
	gc := web.GinContext(c)
	if new != nil {
		gc.Set(gin.AuthUserKey, new.Principal())
		gc.Set(security.ContextKeySecurity, new)
		mw.successHandler.HandleAuthenticationSuccess(c, r, rw, before, new)
	}
	gc.Next()
}

func (mw *ClientAuthMiddleware) handleError(c context.Context, err error) {
	gc := web.GinContext(c)
	security.Clear(gc)
	_ = gc.Error(err)
	gc.Abort()
}