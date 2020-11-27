package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"net/http"
	"strings"
)

// AccessDeniedHandler handles ErrorSubTypeAccessDenied
type AccessDeniedHandler interface {
	HandleAccessDenied(context.Context, *http.Request, http.ResponseWriter, error)
}

// AuthenticationErrorHandler handles ErrorTypeAuthentication
type AuthenticationErrorHandler interface {
	HandleAuthenticationError(context.Context, *http.Request, http.ResponseWriter, error)
}

// AuthenticationEntryPoint kicks off authentication process
type AuthenticationEntryPoint interface {
	Commence(context.Context, *http.Request, http.ResponseWriter, error)
}

/**************************
	Default Impls
***************************/
type DefaultAccessDeniedHandler struct {
}

func (h *DefaultAccessDeniedHandler) HandleAccessDenied(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	if isJson(r) {
		writeErrorAsJson(ctx, http.StatusForbidden, err, rw)
	} else {
		writeErrorAsHtml(ctx, http.StatusForbidden, err, rw)
	}
}

type DefaultAuthenticationErrorHandler struct {
}

func (h *DefaultAuthenticationErrorHandler) HandleAuthenticationError(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	if isJson(r) {
		writeErrorAsJson(ctx, http.StatusForbidden, err, rw)
	} else {
		writeErrorAsHtml(ctx, http.StatusForbidden, err, rw)
	}
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

func writeErrorAsHtml(ctx context.Context, code int, err error, rw http.ResponseWriter) {
	httpError := web.NewHttpError(code, err)
	template.TemplateErrorEncoder(ctx, httpError, rw)
}

func writeErrorAsJson(ctx context.Context, code int, err error, rw http.ResponseWriter) {
	httpError := web.NewHttpError(code, err)
	rest.JsonErrorEncoder(ctx, httpError, rw)
}