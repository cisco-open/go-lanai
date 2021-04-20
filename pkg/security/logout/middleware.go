package logout

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
)

//goland:noinspection GoNameStartsWithPackageName
type LogoutMiddleware struct {
	successHandler security.AuthenticationSuccessHandler
	errorHandler   security.AuthenticationErrorHandler
	logoutHandlers []LogoutHandler
}

func NewLogoutMiddleware(successHandler security.AuthenticationSuccessHandler, errorHandler security.AuthenticationErrorHandler, logoutHandlers ...LogoutHandler) *LogoutMiddleware {
	return &LogoutMiddleware{
		successHandler: successHandler,
		errorHandler:   errorHandler,
		logoutHandlers: logoutHandlers,
	}
}

func (mw *LogoutMiddleware) LogoutHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		before := security.Get(ctx)
		var err error
		for _, handler := range mw.logoutHandlers {
			if e := handler.HandleLogout(ctx, ctx.Request, ctx.Writer, before); e != nil {
				err = e
			}
		}

		if err != nil {
			mw.handleError(ctx, err)
		} else {
			mw.handleSuccess(ctx, before)
		}
	}
}

func (mw *LogoutMiddleware) handleSuccess(c *gin.Context, before security.Authentication) {
	mw.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, security.Get(c))
	if c.Writer.Written() {
		c.Abort()
	}
}

func (mw *LogoutMiddleware) handleError(ctx *gin.Context, err error) {
	mw.errorHandler.HandleAuthenticationError(ctx, ctx.Request, ctx.Writer,
		security.NewInternalAuthenticationError(err))
	if ctx.Writer.Written() {
		ctx.Abort()
	}
}