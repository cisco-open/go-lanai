package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
)

type FormAuthenticationMiddleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
	usernameParam  string
	passwordParam  string
	requestMatcher web.RequestMatcher
}

type FormAuthOptions struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
	UsernameParam  string
	PasswordParam  string
	RequestMatcher web.RequestMatcher
}

func NewFormAuthenticationMiddleware(options... FormAuthOptions) *FormAuthenticationMiddleware {
	mw :=  &FormAuthenticationMiddleware{}
	for _,ops := range options {
		mw.authenticator = ops.Authenticator
		mw.successHandler = ops.SuccessHandler
		mw.usernameParam = ops.UsernameParam
		mw.passwordParam = ops.PasswordParam
		mw.requestMatcher = ops.RequestMatcher
	}
	return mw
}

func (mw *FormAuthenticationMiddleware) LoginProcessHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if match, err := mw.requestMatcher.MatchesWithContext(ctx, ctx.Request); !match {
			if err == nil {
				return
			} else {
				//TODO
				_ = ctx.Error(security.NewAuthenticationError(err.Error()))
				ctx.Abort()
			}
		}

		// TODO
		_ = ctx.Error(security.NewUsernameNotFoundError("Username and Password mismatched"))
		ctx.Abort()
	}
}

func (mw *FormAuthenticationMiddleware) LogoutHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO
	}
}

func (mw *FormAuthenticationMiddleware) EmptyHandlerFunc() gin.HandlerFunc {
	return emptyHandlerFunc
}

func emptyHandlerFunc(*gin.Context) {

}