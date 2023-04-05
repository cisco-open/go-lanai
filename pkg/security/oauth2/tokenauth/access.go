package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
)

/************************
	Access Control
************************/

func ScopesApproved(scopes...string) access.ControlFunc {
	if len(scopes) == 0 {
		return func(_ security.Authentication) (bool, error) {
			return true, nil
		}
	}

	return func(auth security.Authentication) (decision bool, reason error) {
		err := security.NewAccessDeniedError("required scope was not approved by user")
		switch oauth := auth.(type) {
		case oauth2.Authentication:
			if oauth.OAuth2Request() == nil || !oauth.OAuth2Request().Approved() {
				return false, err
			}

			approved := oauth.OAuth2Request().Scopes()
			if approved == nil || !approved.HasAll(scopes...) {
				return false, err
			}
		default:
			return false, err
		}
		return true, nil
	}
}

/******************************
	Access Control Conditions
*******************************/

// RequireScopes returns ControlCondition using ScopesApproved
func RequireScopes(scopes ...string) access.ControlCondition {
	return &access.ConditionWithControlFunc{
		Description:   fmt.Sprintf("client has scopes [%s] approved", scopes),
		ControlFunc:   ScopesApproved(scopes...),
	}
}