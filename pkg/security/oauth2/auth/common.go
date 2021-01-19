package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
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

func ValidateGrant(c context.Context, client OAuth2Client, grantType string) error {
	if grantType == "" {
		return oauth2.NewInvalidTokenRequestError("missing grant_type")
	}

	if !client.GrantTypes().Has(grantType) {
		return oauth2.NewInvalidGrantError(fmt.Sprintf("grant type '%s' is not allowed by this client '%s'", grantType, client.ClientId()))
	}

	return nil
}

func ValidateScope(c context.Context, client OAuth2Client, scopes...string) error {
	for _, scope := range scopes {
		if !client.Scopes().Has(scope) {
			return oauth2.NewInvalidScopeError("unauthorized scope: " + scope)
		}
	}
	return nil
}

func ValidateAllScopes(c context.Context, client OAuth2Client, scopes utils.StringSet) error {
	for scope, _ := range scopes {
		if !client.Scopes().Has(scope) {
			return oauth2.NewInvalidScopeError("unauthorized scope: " + scope)
		}
	}
	return nil
}