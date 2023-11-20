package opainput

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

func PopulateAuthenticationClause(ctx context.Context, input *opa.Input) error {
	auth := security.Get(ctx)
	if !security.IsFullyAuthenticated(auth) {
		input.Authentication = nil
		return nil
	}
	if input.Authentication == nil {
		input.Authentication = opa.NewAuthenticationClause()
	}
	return populateAuthenticationClause(auth, input.Authentication)
}

func populateAuthenticationClause(auth security.Authentication, clause *opa.AuthenticationClause) error {
	clause.Username = getUsernameOrEmpty(auth)
	clause.Permissions = make([]string, 0, len(auth.Permissions()))
	for k := range auth.Permissions() {
		clause.Permissions = append(clause.Permissions, k)
	}

	switch v := auth.(type) {
	case oauth2.Authentication:
		clause.Client = &opa.OAuthClientClause{
			ClientID:  v.OAuth2Request().ClientId(),
			GrantType: v.OAuth2Request().GrantType(),
			Scopes:    v.OAuth2Request().Scopes().Values(),
		}
	default:
	}

	details := auth.Details()
	if v, ok := details.(security.UserDetails); ok {
		clause.UserID = v.UserId()
		clause.AccessibleTenants = v.AssignedTenantIds().Values()
	}
	if v, ok := details.(security.TenantDetails); ok {
		clause.TenantID = v.TenantId()
	}
	if v, ok := details.(security.ProviderDetails); ok {
		clause.ProviderID = v.ProviderId()
	}
	if v, ok := details.(security.AuthenticationDetails); ok {
		clause.Roles = v.Roles().Values()
	}
	return nil
}

func getUsernameOrEmpty(auth security.Authentication) string {
	username, e := security.GetUsername(auth)
	if e != nil {
		return ""
	}
	return username
}
