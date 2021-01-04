package formlogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"net/http"
)

const (
	detailsKeyShouldRemember = "RememberMe"
)

type RememberMeSuccessHandler struct {
	rememberMeParam string
}

func NewRememberMeSuccessHandler(rememberMeParam string) *RememberMeSuccessHandler {
	return &RememberMeSuccessHandler{
		rememberMeParam: rememberMeParam,
	}
}

func (h *RememberMeSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, _ http.ResponseWriter, _, to security.Authentication) {
	details, ok := to.Details().(map[interface{}]interface{})
	if !ok {
		details = map[interface{}]interface{}{}
	}

	// set remember-me decision to auth's details if request has such parameter
	remember := r.PostForm.Get(h.rememberMeParam)
	if remember != "" {
		details[detailsKeyShouldRemember] = true
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
	if doRemember, ok := details[detailsKeyShouldRemember].(bool); !ok || !doRemember {
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