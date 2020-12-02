package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
)

type FormAuthenticationMiddleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
	usernameParam  string
	passwordParam  string
}

type FormAuthOptions struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
	UsernameParam  string
	PasswordParam  string
}

func NewFormAuthenticationMiddleware(options... FormAuthOptions) *FormAuthenticationMiddleware {
	mw :=  &FormAuthenticationMiddleware{}
	for _,ops := range options {
		mw.authenticator = ops.Authenticator
		mw.successHandler = ops.SuccessHandler
		mw.usernameParam = ops.UsernameParam
		mw.passwordParam = ops.PasswordParam
	}
	return mw
}

func (mw *FormAuthenticationMiddleware) LoginProcessHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
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