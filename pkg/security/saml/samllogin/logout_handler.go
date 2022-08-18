package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
)

type SingleLogoutHandler struct {

}

func NewSingleLogoutHandler() *SingleLogoutHandler {
	return &SingleLogoutHandler{
	}
}

// ShouldLogout is a logout.ConditionalLogoutHandler method that interrupt logout process by returning authentication error,
// which would trigger authentication entry point and initiate SLO
func (h *SingleLogoutHandler) ShouldLogout(ctx context.Context, _ *http.Request, _ http.ResponseWriter, auth security.Authentication) error {
	if !h.requiresSamlSLO(ctx, auth) {
		return nil
	}
	return security.NewAuthenticationError("SAML SLO required")
}

func (h *SingleLogoutHandler) HandleLogout(ctx context.Context, _ *http.Request, _ http.ResponseWriter, auth security.Authentication) error {
	if !h.wasSLOFailed(ctx, auth) {
		return nil
	}
	return security.NewAuthenticationWarningError("SAML Single Logout failed")
}

func (h *SingleLogoutHandler) samlDetails(_ context.Context, auth security.Authentication) (map[string]interface{}, bool) {
	switch v := auth.(type) {
	case *samlAssertionAuthentication:
		return v.DetailsMap, true
	default:
		m, _ := auth.Details().(map[string]interface{})
		return m, false
	}
}

func (h *SingleLogoutHandler) requiresSamlSLO(ctx context.Context, auth security.Authentication) bool {
	var isSaml, sloCompleted bool
	var details map[string]interface{}
	// check if it's saml
	details, isSaml = h.samlDetails(ctx, auth)

	// check if SLO already completed
	state, ok := details[kDetailsSLOState].(SLOState)
	sloCompleted = ok && state.Is(SLOCompleted)

	return isSaml && !sloCompleted
}

func (h *SingleLogoutHandler) wasSLOFailed(ctx context.Context, auth security.Authentication) bool {
	var isSaml, sloFailed bool
	var details map[string]interface{}
	details, isSaml = h.samlDetails(ctx, auth)

	// check if SLO already completed
	state, ok := details[kDetailsSLOState].(SLOState)
	sloFailed = ok && state.Is(SLOFailed)

	return isSaml && sloFailed
}
