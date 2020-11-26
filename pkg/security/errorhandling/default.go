package errorhandling

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type DefaultAccessDeniedHandler struct {
	entryPoint security.AuthenticationEntryPoint
}

func (h *DefaultAccessDeniedHandler) HandleAccessDenied(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ctx, ok := c.(*gin.Context)
	if !ok {
		return
	}
	if isJson(r) {
		writeErrorAsJson(ctx, http.StatusForbidden, err)
	} else {
		writeErrorAsHtml(ctx, http.StatusForbidden, err)
	}
}

type DefaultSecurityErrorHandler struct {
	entryPoint security.AuthenticationEntryPoint
}


/**************************
	Helpers
***************************/
func isJson(r *http.Request) bool {
	// TODO should be more comprehensive than this
	accept := r.Header.Get("Accept")
	contentType := r.Header.Get("Content-Type")
	return strings.Contains(accept, "application/json") || strings.Contains(contentType, "application/json")
}

func writeErrorAsHtml(ctx *gin.Context, code int, err error) {

	if coder, ok := err.(security.ErrorCoder); ok {
		ctx.HTML(code, "error.tmpl", gin.H{
			template.ModelKeyError: err,
			template.ModelKeyErrorCode: coder.Code(),
			template.ModelKeyStatusCode: code,
			template.ModelKeyStatusText: http.StatusText(code),
		})
	} else {
		ctx.HTML(code, "error.tmpl", gin.H{
			template.ModelKeyError: err,
			template.ModelKeyStatusCode: code,
			template.ModelKeyStatusText: http.StatusText(code),
		})
	}
}

func writeErrorAsJson(ctx *gin.Context, code int, err error) {

	if coder, ok := err.(security.ErrorCoder); ok {
		ctx.JSON(code, gin.H{
			template.ModelKeyError: err,
			template.ModelKeyErrorCode: coder.Code(),
			template.ModelKeyStatusCode: code,
			template.ModelKeyStatusText: http.StatusText(code),
			template.ModelKeyMessage: err.Error(),
		})
	} else {
		ctx.JSON(code, gin.H{
			template.ModelKeyError: err,
			template.ModelKeyStatusCode: code,
			template.ModelKeyStatusText: http.StatusText(code),
			template.ModelKeyMessage: err.Error(),
		})
	}
}