package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
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

		username := ctx.PostFormArray(mw.usernameParam)
		if len(username) == 0 {
			username = []string{""}
		}

		password := ctx.PostFormArray(mw.passwordParam)
		if len(password) == 0 {
			password = []string{""}
		}

		currentAuth, ok := security.Get(ctx).(passwd.UsernamePasswordAuthentication)
		if ok && currentAuth.Authenticated() && passwd.IsSamePrincipal(username[0], currentAuth) {
			// We currently allow re-authenticate without logout.
			// If we don't want to allow it, we need to figure out how to error out without clearing the authentication.
			// Note: currently, clearing authentication happens in error handling middleware on all SecurityAuthenticationError
		}

		candidate := passwd.UsernamePasswordPair{
			Username: username[0],
			Password: password[0],
			EnforceMFA: passwd.MFAModeOptional,
		}
		// Search auth in the slice of allowed credentials
		auth, err := mw.authenticator.Authenticate(&candidate)
		if err != nil {
			mw.handleError(ctx, err, &candidate)
			return
		}
		mw.handleSuccess(ctx, auth)
	}
}

func (mw *FormAuthenticationMiddleware) EndpointHandlerFunc() gin.HandlerFunc {
	return notFoundHandlerFunc
}

func (mw *FormAuthenticationMiddleware) handleSuccess(c *gin.Context, new security.Authentication) {
	if new != nil {
		c.Set(gin.AuthUserKey, new.Principal())
		c.Set(security.ContextKeySecurity, new)
	}
	mw.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, new)
	if c.Writer.Written() {
		c.Abort()
	}
}

func (mw *FormAuthenticationMiddleware) handleError(c *gin.Context, err error, candidate security.Candidate) {
	if candidate != nil {
		s := session.Get(c)
		if s != nil {
			s.AddFlash(candidate.Principal(), mw.usernameParam)
		}
	}
	_ = c.Error(err)
	c.Abort()
}

func notFoundHandlerFunc(c *gin.Context) {
	_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("page not found"))
}