package formlogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"net/http"
)

const (
	detailsKeyShouldRememberUsername = "RememberUsername"
)

type RememberUsernameSuccessHandler struct {
	rememberParam string
}

func NewRememberUsernameSuccessHandler(rememberParam string) *RememberUsernameSuccessHandler {
	return &RememberUsernameSuccessHandler{
		rememberParam: rememberParam,
	}
}

func (h *RememberUsernameSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, _ http.ResponseWriter, _, to security.Authentication) {
	details, ok := to.Details().(map[interface{}]interface{})
	if !ok {
		details = map[interface{}]interface{}{}
	}

	// set remember-me decision to auth's details if request has such parameter
	remember := r.PostForm.Get(h.rememberParam)
	if remember != "" {
		details[detailsKeyShouldRememberUsername] = true
	}

	// auth process not finished yet, bail
	if to.State() < security.StateAuthenticated {
		return
	}

	s := session.Get(c)
	if s == nil {
		return
	}

	// read remember-me decision from auth
	if doRemember, ok := details[detailsKeyShouldRememberUsername].(bool); !ok || !doRemember {
		// cleanup session
		s.Set(SessionKeyRememberedUsername, nil)
		return
	}

	// remember username
	switch to.Principal().(type) {
	case security.Account:
		s.Set(SessionKeyRememberedUsername, to.Principal().(security.Account).Username())
	case string:
		s.Set(SessionKeyRememberedUsername, to.Principal().(string))
	}
}