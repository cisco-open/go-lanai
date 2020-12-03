package logout

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LogoutMiddleware struct {
	successHandler security.AuthenticationSuccessHandler
	logoutHandlers []LogoutHandler
}

type LogoutOptions struct {
	SuccessHandler security.AuthenticationSuccessHandler
}

func NewLogoutMiddleware(successHandler security.AuthenticationSuccessHandler, logoutHandlers ...LogoutHandler) *LogoutMiddleware {
	return &LogoutMiddleware{
		successHandler: successHandler,
		logoutHandlers: logoutHandlers,
	}
}

func (mw *LogoutMiddleware) LogoutHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO
	}
}

func (mw *LogoutMiddleware) EndpointHandlerFunc() gin.HandlerFunc {
	return notFoundHandlerFunc
}

func notFoundHandlerFunc(c *gin.Context) {
	_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("page not found"))
}