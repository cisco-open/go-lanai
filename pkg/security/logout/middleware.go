package logout

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
)

//goland:noinspection GoNameStartsWithPackageName
type LogoutMiddleware struct {
	successHandler      security.AuthenticationSuccessHandler
	errorHandler        security.AuthenticationErrorHandler
	entryPoint          security.AuthenticationEntryPoint
	logoutHandlers      []LogoutHandler
	conditionalHandlers []ConditionalLogoutHandler
}

func NewLogoutMiddleware(successHandler security.AuthenticationSuccessHandler,
	errorHandler security.AuthenticationErrorHandler,
	entryPoint security.AuthenticationEntryPoint,
	logoutHandlers ...LogoutHandler) *LogoutMiddleware {

	conditionalHandlers := make([]ConditionalLogoutHandler, 0, len(logoutHandlers))
	for _, h := range logoutHandlers {
		if conditional, ok := h.(ConditionalLogoutHandler); ok {
			conditionalHandlers = append(conditionalHandlers, conditional)
		}
	}
	return &LogoutMiddleware{
		successHandler:      successHandler,
		errorHandler:        errorHandler,
		entryPoint:          entryPoint,
		logoutHandlers:      logoutHandlers,
		conditionalHandlers: conditionalHandlers,
	}
}

func (mw *LogoutMiddleware) LogoutHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		before := security.Get(gc)
		// pre-logout check
		for _, h := range mw.conditionalHandlers {
			switch e := h.ShouldLogout(gc, gc.Request, gc.Writer, before); {
			case e != nil && mw.entryPoint != nil:
				mw.handleCancelled(gc, e)
				return
			case e != nil:
				mw.handleError(gc, e)
				return
			}
		}

		// perform logout
		for _, handler := range mw.logoutHandlers {
			if e := handler.HandleLogout(gc, gc.Request, gc.Writer, before); e != nil {
				mw.handleError(gc, e)
				return
			}
		}
		mw.handleSuccess(gc, before)
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

func (mw *LogoutMiddleware) handleCancelled(ctx *gin.Context, err error) {
	mw.entryPoint.Commence(ctx, ctx.Request, ctx.Writer, err)
	if ctx.Writer.Written() {
		ctx.Abort()
	}
}
