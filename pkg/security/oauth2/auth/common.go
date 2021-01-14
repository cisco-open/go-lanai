package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
)

/***********************
	Common Functions
 ***********************/
func RetrieveAuthenticatedClient(c context.Context) OAuth2Client {
	sec := security.Get(c)
	if sec.State() < security.StateAuthenticated {
		return nil
	}

	if client, ok := sec.Principal().(OAuth2Client); ok {
		return client

	}
	return nil
}

func ValidateGrant(c context.Context, req *TokenRequest, client OAuth2Client) error {
	if req.GrantType == "" {
		return oauth2.NewInvalidTokenRequestError("missing grant_type")
	}

	if !client.GrantTypes().Has(req.GrantType) {
		return oauth2.NewInvalidGrantError(fmt.Sprintf("grant type '%s' is not allowed by this client '%s'", req.GrantType, client.ClientId()))
	}

	return nil
}

func ValidateScope(c context.Context, req *TokenRequest, client OAuth2Client) error {
	for scope, _ := range req.Scopes {
		if !client.Scopes().Has(scope) {
			return oauth2.NewInvalidScopeError("invalid scope: " + scope)
		}
	}
	return nil
}