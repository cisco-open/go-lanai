package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
)

type DebugAuthSuccessHandler struct {}

func (h *DebugAuthSuccessHandler) HandleAuthenticationSuccess(
	_ context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	logger.Debugf("session knows auth succeeded: from [%v] to [%v]", from, to)
}

type DebugAuthErrorHandler struct {}

func (h *DebugAuthErrorHandler) HandleAuthenticationError(_ context.Context, _ *http.Request, _ http.ResponseWriter, err error) {
	logger.Debugf("session knows auth failed with %v", err.Error())
}
