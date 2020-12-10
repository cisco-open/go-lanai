package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
	"sort"
)

/**
	This is a high priority handler because it writes to the header.
	Therefore it must be before any other success handler that may write the response status (e.g. redirect handler)
 */
type ChangeSessionHandler struct{}

func (h *ChangeSessionHandler) HandleAuthenticationSuccess(c context.Context, _ *http.Request, w http.ResponseWriter, _ security.Authentication) {
	s := Get(c)
	if s == nil {
		return
	}

	err := s.ChangeId()

	if err == nil {
		http.SetCookie(w, NewCookie(s.Name(), s.id, s.options))
	} else {
		panic(security.NewInternalError("Failed to update session ID"))
	}
}

func (h *ChangeSessionHandler) PriorityOrder() int {
	return security.HandlerOrderChangeSession
}

type GetMaximumSessions func() int

//This handler runs after ChangeSessionHandler so that the updated session id is indexed to the principal
type ConcurrentSessionHandler struct{
	sessionStore Store
	getMaxSessions GetMaximumSessions
}

func (h *ConcurrentSessionHandler) HandleAuthenticationSuccess(c context.Context, _ *http.Request, w http.ResponseWriter, auth security.Authentication) {
	s := Get(c)
	if s == nil {
		return
	}

	p, err := getPrincipalName(auth)
	if err != nil {
		//Auth is something we don't recognize, this indicates a program error
		panic(security.NewInternalError(err.Error()))
	}

	//Adding to the index before checking the limit.
	//If done other way around, concurrent logins may be doing the check before the other request added to the index
	//thus making it possible to exceed the limit
	//By doing the check at the end, we can end up with the right number of sessions when all requests finishes.
	err = h.sessionStore.AddToPrincipalIndex(p, s)
	if err != nil {
		panic(security.NewInternalError(err.Error()))
	}

	sessionName := s.Name()

	//This will also clean the expired sessions from the index, so we do it regardless if max sessions is set or not
	existing, err := h.sessionStore.FindByPrincipalName(p, sessionName)

	if err != nil {
		panic(security.NewInternalError(err.Error()))
	}

	max := h.getMaxSessions()

	if len(existing) <= max || max <= 0 {
		return
	}

	sort.SliceStable(existing, func(i, j int) bool {
		return existing[i].createdOn().Before(existing[j].createdOn())
	})

	for i := 0; i < len(existing) - max; i++ {
		err = h.sessionStore.Delete(existing[i])

		if err != nil {
			panic(security.NewInternalError("Cannot delete session that exceeded max concurrent session limit"))
		}

		err = h.sessionStore.RemoveFromPrincipalIndex(p, existing[i])
	}
}

func (h *ConcurrentSessionHandler) PriorityOrder() int {
	return security.HandlerOrderConcurrentSession
}

func getPrincipalName(auth security.Authentication) (string, error) {
	if account, ok := auth.Principal().(security.Account); ok {
		return account.Username(), nil
	} else if principal, ok := auth.Principal().(string); ok {
		return principal, nil
	} else {
		return "", security.NewInternalError("unrecognized principal type")
	}
}


//TODO:
type LogoutSuccessHandler struct {

}

func (h *LogoutSuccessHandler) HandleAuthenticationSuccess(c context.Context, _ *http.Request, w http.ResponseWriter, auth security.Authentication) {

}