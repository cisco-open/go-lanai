package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"net/http"
)

// OAuth2ErrorHanlder implements security.AccessDeniedHandler and security.AuthenticationErrorHandler
type OAuth2ErrorHanlder struct {}

func NewOAuth2ErrorHanlder() *OAuth2ErrorHanlder {
	return &OAuth2ErrorHanlder{}
}

// security.AuthenticationErrorHandler
func (h *OAuth2ErrorHanlder) HandleAuthenticationError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	// TODO translate error
	writeErrorAsJson(c, http.StatusBadRequest, err, rw)
}

// security.AccessDeniedHandler
func (h *OAuth2ErrorHanlder) HandleAccessDenied(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	// TODO translate error
	writeErrorAsJson(c, http.StatusBadRequest, err, rw)
}

func writeErrorAsJson(ctx context.Context, code int, err error, rw http.ResponseWriter) {
	httpError := web.NewHttpError(code, err)
	rest.JsonErrorEncoder(ctx, httpError, rw)
}

