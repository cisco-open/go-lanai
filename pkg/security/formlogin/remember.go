package formlogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"net/http"
	"net/url"
	"time"
)

const (
	detailsKeyShouldRememberUsername = "RememberUsername"
)

type RememberUsernameSuccessHandler struct {
	cookieProps   security.CookieProperties
	serverProps   web.ServerProperties
	rememberParam string
}

func NewRememberUsernameSuccessHandler(cookieProps security.CookieProperties, serverProps web.ServerProperties, rememberParam string) *RememberUsernameSuccessHandler {
	return &RememberUsernameSuccessHandler{
		cookieProps: cookieProps,
		serverProps: serverProps,
		rememberParam: rememberParam,
	}
}

func (h *RememberUsernameSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, _, to security.Authentication) {
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

	// read remember-me decision from auth
	if doRemember, ok := details[detailsKeyShouldRememberUsername].(bool); !ok || !doRemember {
		// cleanup session
		h.clear(c, rw)
		return
	}

	// remember username
	switch to.Principal().(type) {
	case security.Account:
		h.save(to.Principal().(security.Account).Username(), c, rw)
	case string:
		h.save(to.Principal().(string), c, rw)
	}
}

func (h *RememberUsernameSuccessHandler) save(username string, _ context.Context, rw http.ResponseWriter) {
	cookie := h.newCookie(CookieKeyRememberedUsername, username, 0)
	http.SetCookie(rw, cookie)
}

func (h *RememberUsernameSuccessHandler) clear(_ context.Context, rw http.ResponseWriter) {
	cookie := h.newCookie(CookieKeyRememberedUsername, "", -1)
	http.SetCookie(rw, cookie)
}

func (h *RememberUsernameSuccessHandler) newCookie(name, value string, maxAge int) *http.Cookie {

	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     h.serverProps.ContextPath,
		Domain:   h.cookieProps.Domain,
		MaxAge:   maxAge,
		Expires:  calculateCookieExpires(maxAge),
		Secure:   h.cookieProps.Secure,
		HttpOnly: h.cookieProps.HttpOnly,
		SameSite: h.cookieProps.SameSite(),
	}

	return cookie
}

func calculateCookieExpires(maxAge int) time.Time {
	if maxAge == 0 {
		return time.Time{}
	}

	d := time.Duration(maxAge) * time.Second
	return time.Now().Add(d)
}