package clientauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"errors"
	"github.com/gin-gonic/gin"
)

type ClientAuthMiddleware struct {}

type ClientAuthMWOptions func(*ClientAuthMWOption)

type ClientAuthMWOption struct {}

func NewClientAuthMiddleware(opts...ClientAuthMWOptions) *ClientAuthMiddleware {
	opt := ClientAuthMWOption{}

	for _, optFunc := range opts {
		if optFunc != nil {
			optFunc(&opt)
		}
	}
	return &ClientAuthMiddleware{}
}

//func (mw *ClientAuthMiddleware) ClientAuthFormHandlerFunc() http.HandlerFunc {
//	return func(rw http.ResponseWriter, r *http.Request) {
//
//	}
//}

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
