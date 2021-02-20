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
func RetrieveAuthenticatedClient(c context.Context) oauth2.OAuth2Client {
	if client, ok := c.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client); ok {
		return client
	}

	sec := security.Get(c)
	if sec.State() < security.StatePrincipalKnown {
		return nil
	}

	if client, ok := sec.Principal().(oauth2.OAuth2Client); ok {
		return client
	}
	return nil
}

func RetrieveFullyAuthenticatedClient(c context.Context) (oauth2.OAuth2Client, error) {
	sec := security.Get(c)
	if sec.State() < security.StateAuthenticated {
		return nil, oauth2.NewInvalidGrantError("client is not fully authenticated")
	}

	if client, ok := sec.Principal().(oauth2.OAuth2Client); ok {
		return client, nil
	}
	return nil, oauth2.NewInvalidGrantError("client is not fully authenticated")
}

func ValidateGrant(c context.Context, client oauth2.OAuth2Client, grantType string) error {
	if grantType == "" {
		return oauth2.NewInvalidTokenRequestError("missing grant_type")
	}

	if !client.GrantTypes().Has(grantType) {
		return oauth2.NewUnauthorizedClientError(fmt.Sprintf("grant type '%s' is not allowed by this client '%s'", grantType, client.ClientId()))
	}

	return nil
}

func ValidateScope(c context.Context, client oauth2.OAuth2Client, scopes...string) error {
	for _, scope := range scopes {
		if !client.Scopes().Has(scope) {
			return oauth2.NewInvalidScopeError("invalid scope: " + scope)
		}
	}
	return nil
}

func ValidateAllScopes(c context.Context, client oauth2.OAuth2Client, scopes utils.StringSet) error {
	if ok, invalid := IsSubSet(c, client.Scopes(), scopes); !ok {
		return oauth2.NewInvalidScopeError("invalid scope: " + invalid)
	}
	return nil
}

func ValidateAllAutoApprovalScopes(c context.Context, client oauth2.OAuth2Client, scopes utils.StringSet) error {
	if ok, invalid := IsSubSet(c, client.AutoApproveScopes(), scopes); !ok {
		return oauth2.NewAccessRejectedError("scope not auto approved: " + invalid)
	}
	return nil
}

func IsSubSet(c context.Context, superset utils.StringSet, subset utils.StringSet) (ok bool, invalid string) {
	for scope, _ := range subset {
		if !superset.Has(scope) {
			return false, scope
		}
	}
	return true, ""
}

// approval param is a map with scope as keys and approval status as values
func ValidateApproval(c context.Context, approval map[string]bool, client oauth2.OAuth2Client, scopes utils.StringSet) error {
	if e := ValidateAllScopes(c, client, scopes); e != nil {
		return e
	}

	for scope, _ := range scopes {
		if approved, ok := approval[scope]; !ok || !approved {
			return oauth2.NewAccessRejectedError(fmt.Sprintf("user disapproved scope [%s]", scope))
		}
	}
	return nil
}