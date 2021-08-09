package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

/************************
	security.Candidate
************************/

// BearerToken is the supported security.Candidate of resource server authenticator
type BearerToken struct {
	Token string
	DetailsMap map[string]interface{}
}

func (t *BearerToken) Principal() interface{} {
	return ""
}

func (t *BearerToken) Credentials() interface{} {
	return t.Token
}

func (t *BearerToken) Details() interface{} {
	return t.DetailsMap
}

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