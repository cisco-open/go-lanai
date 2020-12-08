package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
)

/**
	This is a high priority handler because it writes to the header.
	Therefore it must be before any other success handler that may write the response status (e.g. redirect handler)
 */
type ChangeSessionHandler struct{}

func (h *ChangeSessionHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, w http.ResponseWriter, a security.Authentication) {
	s := Get(c)
	if s == nil {
		return
	}

	err := s.ChangeId()

	if err == nil {
		http.SetCookie(w, NewCookie(s.Name(), s.id, s.options))
	} else {
		panic(security.NewAuthenticationInternalError("Failed to update session ID"))
	}
}

func (h *ChangeSessionHandler) PriorityOrder() int {
	return security.HandlerOrderChangeSession
}

type ConcurrentSessionHandler struct{}

func (h *ConcurrentSessionHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, w http.ResponseWriter, a security.Authentication) {
	//TODO:
	// 1. need to be able to count number of session per user
	// 2. if session under limit, do nothing
	// 3. if session is over limit, kickout the oldest session
	// This means the session store needs a getAllSessions(principal) method on the store
}
