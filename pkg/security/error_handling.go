package security

import (
	"context"
	"net/http"
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