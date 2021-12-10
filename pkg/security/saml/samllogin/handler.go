package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/crewjam/saml/samlsp"
	"net/http"
)

type TrackedRequestSuccessHandler struct {
	tracker samlsp.RequestTracker
}

func NewTrackedRequestSuccessHandler(tracker samlsp.RequestTracker) security.AuthenticationSuccessHandler{
	return &TrackedRequestSuccessHandler{
		tracker: tracker,
	}
}

func (t *TrackedRequestSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	redirectURI := "/"
	if trackedRequestIndex := r.Form.Get("RelayState"); trackedRequestIndex != "" {
		trackedRequest, err := t.tracker.GetTrackedRequest(r, trackedRequestIndex)
		if err == nil {
			redirectURI = trackedRequest.URI
		} else {
			logger.Errorf("error getting tracked request %v", err)
		}
		_ = t.tracker.StopTrackingRequest(rw, r, trackedRequestIndex)
	}
	http.Redirect(rw, r, redirectURI, 302)
	_,_ = rw.Write([]byte{})
}