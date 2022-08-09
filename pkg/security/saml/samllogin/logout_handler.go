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
	// TODO check if SAML
	if !h.requiresSamlSLO(ctx, auth) {
		return nil
	}
	return security.NewAuthenticationError("SAML SLO required")
}

func (h *SingleLogoutHandler) HandleLogout(ctx context.Context, r *http.Request, rw http.ResponseWriter, auth security.Authentication) error {
	// TODO validate again if SLO performed
	return nil
}

func (h *SingleLogoutHandler) requiresSamlSLO(ctx context.Context, auth security.Authentication) bool {
	logger.WithContext(ctx).Infof("Logging out with auth: %v", auth)
	var isSaml, sloCompleted bool

	switch m := auth.Details().(type) {
	case map[string]interface{}:
		// check if it's saml
		if method, ok := m[security.DetailsKeyAuthMethod].(string); ok && method == security.AuthMethodExternalSaml {
			isSaml = true
		}
		// check if SLO already completed
		if state, ok := m[kDetailsSLOState].(SLOState); ok && state.Is(SLOCompleted) {
			sloCompleted = true
		}
	default:
		switch auth.(type) {
		case *samlAssertionAuthentication:
			isSaml = true
		}
	}
	return isSaml && !sloCompleted
}

