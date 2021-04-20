package errorhandling

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
)

//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingMiddleware struct {
	entryPoint          security.AuthenticationEntryPoint
	accessDeniedHandler security.AccessDeniedHandler
	authErrorHandler    security.AuthenticationErrorHandler
	errorHandler 		security.ErrorHandler
}

func NewErrorHandlingMiddleware() *ErrorHandlingMiddleware {
	return &ErrorHandlingMiddleware{}
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
	eh.logError(c, err)

	if c.Writer.Written() {
		return
	}

	// we assume eh.authErrorHandler and eh.accessDeniedHandler is always not nil (guaranteed by ErrorHandlingConfigurer)
	switch {
	case errors.Is(err, security.ErrorTypeInternal):
		eh.authErrorHandler.HandleAuthenticationError(c, c.Request, c.Writer, err)

	case eh.entryPoint != nil && errors.Is(err, security.ErrorSubTypeInsufficientAuth):
		eh.entryPoint.Commence(c, c.Request, c.Writer, err)

	case errors.Is(err, security.ErrorTypeAuthentication):
		eh.authErrorHandler.HandleAuthenticationError(c, c.Request, c.Writer, err)

	case errors.Is(err, security.ErrorTypeAccessControl):
		eh.accessDeniedHandler.HandleAccessDenied(c, c.Request, c.Writer, err)

	default:
		eh.errorHandler.HandleError(c, c.Request, c.Writer, err)
	}
}

func (eh *ErrorHandlingMiddleware) logError(c *gin.Context, err error) {
	errMsgs := []string{err.Error()}
	for cause, isNested := err, true; isNested && cause != nil; {
		var nested errorutils.NestedError
		if nested, isNested = cause.(errorutils.NestedError); isNested {
			cause = nested.Cause()
			if cause != nil {
				errMsgs = append(errMsgs, cause.Error())
			}
		}
	}
	msg := strings.Join(errMsgs, " - [Caused By]: ")
	logger.WithContext(c.Request.Context()).Debugf("[Error]: %s", msg)
}

/**************************
	Helpers
***************************/
