package logout

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

//goland:noinspection GoNameStartsWithPackageName
type LogoutMiddleware struct {
	successHandler security.AuthenticationSuccessHandler
	logoutHandlers []LogoutHandler
}

func NewLogoutMiddleware(successHandler security.AuthenticationSuccessHandler, logoutHandlers ...LogoutHandler) *LogoutMiddleware {
	return &LogoutMiddleware{
		successHandler: successHandler,
		logoutHandlers: logoutHandlers,
	}
}

func (mw *LogoutMiddleware) LogoutHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentAuth := security.Get(ctx)
		if currentAuth == nil || !currentAuth.Authenticated() {
			mw.handleSuccess(ctx)
			return
		}

		for _, handler := range mw.logoutHandlers {
			handler.HandleLogout(ctx, ctx.Request, ctx.Writer, currentAuth)
		}
		mw.handleSuccess(ctx)
	}
}

func (mw *LogoutMiddleware) EndpointHandlerFunc() gin.HandlerFunc {
	return notFoundHandlerFunc
}

func (mw *LogoutMiddleware) handleSuccess(c *gin.Context) {
	c.Set(gin.AuthUserKey, nil)
	c.Set(security.ContextKeySecurity, nil)
	mw.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, nil)
	if c.Writer.Written() {
		c.Abort()
	}
}

func notFoundHandlerFunc(c *gin.Context) {
	_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("page not found"))
}