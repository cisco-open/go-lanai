package csrf

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type manager struct {
	tokenStore TokenStore
	requireProtection web.RequestMatcher
	parameterName string
	headerName string

}

func newManager(tokenStore TokenStore, csrfProtectionMatcher web.RequestMatcher) *manager {
	if csrfProtectionMatcher == nil {
		csrfProtectionMatcher = matcher.NotRequest(matcher.RequestWithMethods("GET", "HEAD", "TRACE", "OPTIONS"))
	}

	return &manager{
		tokenStore: tokenStore,
		parameterName: "_csrf",
		headerName: "X-CSRF-TOKEN",
		requireProtection: csrfProtectionMatcher,
	}
}

//TODO CsrfAuthenticationStrategy - check if template rendering is done before or after middleware code.

func (m *manager) CsrfHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedToken, err := m.tokenStore.LoadToken(c)

		// this means there's something wrong with reading the csrf token from storage - e.g. can't deserialize or can't access storage
		// this means we can't recover, so abort
		if err != nil {
			_ = c.Error(security.NewInternalError(err.Error()))
			c.Abort()
		}

		if expectedToken == nil {
			expectedToken = m.tokenStore.Generate(c, m.parameterName, m.headerName)
			err = m.tokenStore.SaveToken(c, expectedToken)
			if err != nil {
				_ = c.Error(security.NewInternalError(err.Error()))
				c.Abort()
			}
		}

		//This so that the templates knows what to render to
		//we don't depend on the value being stored in session to decouple it from the store implementation.
		c.Set(web.ContextKeyCsrf, expectedToken)

		matches, err := m.requireProtection.MatchesWithContext(c, c.Request)
		if err != nil {
			_ = c.Error(security.NewInternalError(err.Error()))
			c.Abort()
		}

		if matches {
			actualToken := c.GetHeader(m.headerName)

			if actualToken == "" {
				actualToken, _ = c.GetPostForm(m.parameterName)
			}

			//both error case returns access denied, but the error message may be different
			if actualToken == "" {
				_ = c.Error(security.NewMissingCsrfTokenError("request is missing csrf token"))
				c.Abort()
			} else if actualToken != expectedToken.Value {
				_ = c.Error(security.NewInvalidCsrfTokenError("request has invalid csrf token"))
				c.Abort()
			}
		}
	}
}

type CsrfDeniedHandler struct {
	delegate security.AccessDeniedHandler
}

// implement order.Ordered
func (h *CsrfDeniedHandler) Order() int {
	return 0
}

// implement security.AccessDeniedHandler
func (h *CsrfDeniedHandler) HandleAccessDenied(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, security.ErrorSubTypeCsrf):
		h.delegate.HandleAccessDenied(c, r, rw, err)
	}
}

