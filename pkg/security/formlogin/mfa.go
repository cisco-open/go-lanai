package formlogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"net/http"
)

type MfaAwareAuthenticationEntryPoint struct {
	delegate security.AuthenticationEntryPoint
	mfaPendingDelegate security.AuthenticationEntryPoint
}

func (h *MfaAwareAuthenticationEntryPoint) Commence(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	auth,ok := security.Get(c).(passwd.UsernamePasswordAuthentication)
	if ok && auth.IsMFAPending() {
		h.mfaPendingDelegate.Commence(c, r, rw, err)
	} else {
		h.delegate.Commence(c, r, rw, err)
	}
}

type MfaAwareSuccessHandler struct {
	delegate security.AuthenticationSuccessHandler
	mfaPendingDelegate security.AuthenticationSuccessHandler
}

func (h *MfaAwareSuccessHandler) HandleAuthenticationSuccess(
	c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	userAuth,ok := to.(passwd.UsernamePasswordAuthentication)
	if ok && userAuth.IsMFAPending() {
		h.mfaPendingDelegate.HandleAuthenticationSuccess(c, r, rw, from, to)
	} else {
		h.delegate.HandleAuthenticationSuccess(c, r, rw, from, to)
	}
}

type MfaAwareAuthenticationErrorHandler struct {
	delegate security.AuthenticationErrorHandler
	mfaPendingDelegate security.AuthenticationErrorHandler
}

func (h *MfaAwareAuthenticationErrorHandler) HandleAuthenticationError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	auth,ok := security.Get(c).(passwd.UsernamePasswordAuthentication)
	if ok && auth.IsMFAPending() {
		h.mfaPendingDelegate.HandleAuthenticationError(c, r, rw, err)
	} else {
		h.delegate.HandleAuthenticationError(c, r, rw, err)
	}
}
