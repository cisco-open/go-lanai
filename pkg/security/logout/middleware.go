package logout

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
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

func (mw *LogoutMiddleware) EmptyHandlerFunc() gin.HandlerFunc {
	return emptyHandlerFunc
}

func emptyHandlerFunc(*gin.Context) {

}