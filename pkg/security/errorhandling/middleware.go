package errorhandling

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/requestcache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"errors"
	"github.com/gin-gonic/gin"
)

//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingMiddleware struct {
	entryPoint security.AuthenticationEntryPoint
	accessDeniedHandler security.AccessDeniedHandler
	authErrorHandler security.AuthenticationErrorHandler
	saveRequestMatcher web.RequestMatcher
}

func NewErrorHandlingMiddleware() *ErrorHandlingMiddleware {
	notFavicon := matcher.NotRequest(matcher.RequestWithPattern("/**/favicon.*"))
	notXMLHttpRequest := matcher.NotRequest(matcher.RequestWithHeader("X-Requested-With", "XMLHttpRequest", false))
	notTrailer := matcher.NotRequest(matcher.RequestHasHeader("Trailer"))
	notMultiPart := matcher.NotRequest(matcher.RequestWithHeader("Content-Type", "multipart/form-data", true))
	notCsrf := matcher.NotRequest(matcher.RequestHasHeader(csrf.CsrfHeaderName).Or(matcher.RequestHasPostParameter(csrf.CsrfParamName)))

	savedRequestMatcher := notFavicon.And(notXMLHttpRequest).And(notTrailer).And(notMultiPart).And(notCsrf)
	return &ErrorHandlingMiddleware{
		saveRequestMatcher: savedRequestMatcher,
	}
}

func (eh *ErrorHandlingMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer eh.tryRecover(ctx)
		ctx.Next()
		eh.tryHandleErrors(ctx)
	}
}

func (eh *ErrorHandlingMiddleware) tryRecover(c *gin.Context) {
	r := recover()
	if r == nil {
		// nothing to recover
		return
	}

	err,ok := r.(error)
	if !ok || !errors.Is(err, security.ErrorTypeSecurity) {
		// not something we can handle
		panic(r)
	}

	eh.handleError(c, err)
}

func (eh *ErrorHandlingMiddleware) tryHandleErrors(c *gin.Context) {
	// find first security error and handle it
	for _,e := range c.Errors {
		if errors.Is(e.Err, security.ErrorTypeSecurity) {
			eh.handleError(c, e.Err)
			break
		}
	}
}

func (eh *ErrorHandlingMiddleware) handleError(c *gin.Context, err error) {
	if c.Writer.Written() {
		return
	}

	// we assume eh.authErrorHandler and eh.accessDeniedHandler is always not nil (guaranteed by ErrorHandlingConfigurer)
	switch {
	case errors.Is(err, security.ErrorTypeInternal):
		eh.authErrorHandler.HandleAuthenticationError(c, c.Request, c.Writer, err)

	case eh.entryPoint != nil && errors.Is(err, security.ErrorSubTypeInsufficientAuth):
		if match, err := eh.saveRequestMatcher.MatchesWithContext(c, c.Request); match && err == nil{
			requestcache.SaveRequest(c)
		}

		eh.entryPoint.Commence(c, c.Request, c.Writer, err)

	case errors.Is(err, security.ErrorTypeAuthentication):
		eh.authErrorHandler.HandleAuthenticationError(c, c.Request, c.Writer, err)

	default:
		eh.accessDeniedHandler.HandleAccessDenied(c, c.Request, c.Writer, err)
	}
}

/**************************
	Helpers
***************************/
