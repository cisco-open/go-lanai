package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
	"strings"
)

var (
	switchTenantPermissions = []string{
		security.SpecialPermissionSwitchTenant,
	}
)

// SwitchTenantGranter implements auth.TokenGranter
type SwitchTenantGranter struct {
	PermissionBasedGranter
	accountStore oauth2.OAuth2AccountStore
	authService  auth.AuthorizationService
}

func NewSwitchTenantGranter(
	authService auth.AuthorizationService,
	authenticator security.Authenticator,
	accountStore oauth2.OAuth2AccountStore,
) *SwitchTenantGranter {
	if authenticator == nil {
		panic(fmt.Errorf("cannot create SwitcTenantGranter without authenticator."))
	}

	if authService == nil {
		panic(fmt.Errorf("cannot create SwitchTenantGranter without authorization service."))
	}

	if accountStore == nil {
		panic(fmt.Errorf("cannot create SwitchTenantGranter without account store."))
	}

	return &SwitchTenantGranter{
		PermissionBasedGranter: PermissionBasedGranter{
			authenticator: authenticator,
		},
		authService:  authService,
		accountStore: accountStore,
	}
}

func (g *SwitchTenantGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypeSwitchTenant != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// common check
	if e := auth.ValidateGrant(ctx, client, request.GrantType); e != nil {
		return nil, e
	}

	// additional request params check
	if e := g.validateRequest(ctx, request); e != nil {
		return nil, e
	}

	// extract existing auth
	stored, e := g.authenticateToken(ctx, request)
	if e != nil {
		return nil, e
	}

	// check permissions
	if e := g.validate(ctx, request, stored); e != nil {
		return nil, e
	}

	// additional check
	// check client details (if client ID matches, if all requested scope is allowed, etc)
	if e := g.validateStoredClient(ctx, client, stored.OAuth2Request()); e != nil {
		return nil, e
	}

	// get merged request with reduced scope
	req, e := g.reduceScope(ctx, client, stored.OAuth2Request(), request)
	if e != nil {
		return nil, e
	}

	oauth, e := g.authService.SwitchAuthentication(ctx, req, stored.UserAuthentication(), stored)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}

	// create token
	token, e := g.authService.CreateAccessToken(ctx, oauth)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}
	return token, nil
}

func (g *SwitchTenantGranter) validateRequest(ctx context.Context, request *auth.TokenRequest) error {
	tenantId, idOk := request.Extensions[oauth2.ParameterTenantId].(string)
	tenantExternalid, nameOk := request.Extensions[oauth2.ParameterTenantExternalId].(string)
	if !nameOk && !idOk {
		return oauth2.NewInvalidTokenRequestError(fmt.Sprintf("both [%s] and [%s] are missing", oauth2.ParameterTenantId, oauth2.ParameterTenantExternalId))
	}

	if strings.TrimSpace(tenantId) == "" && strings.TrimSpace(tenantExternalid) == "" {
		return oauth2.NewInvalidTokenRequestError(fmt.Sprintf("both [%s] and [%s] are empty", oauth2.ParameterTenantId, oauth2.ParameterTenantExternalId))
	}
	return nil
}

func (g *SwitchTenantGranter) validate(ctx context.Context, request *auth.TokenRequest, stored security.Authentication) error {
	if e := g.PermissionBasedGranter.validateStoredPermissions(ctx, stored, switchTenantPermissions...); e != nil {
		return e
	}

	if proxy, ok := stored.Details().(security.ProxiedUserDetails); ok && proxy.Proxied() {
		return oauth2.NewInvalidGrantError("the access token represents a masqueraded context. need original token to switch tenant")
	}

	srcTenant, ok := stored.Details().(security.TenantDetails)
	if !ok {
		// there wasn't any tenant. shouldn't happen, but we allow it because it won't cause any trouble
		return nil
	}

	var tenantHasChanged bool

	tenantId, _ := request.Extensions[oauth2.ParameterTenantId].(string)
	if strings.TrimSpace(tenantId) != srcTenant.TenantId() {
		tenantHasChanged = true
	}
	tenantExternalId, _ := request.Extensions[oauth2.ParameterTenantExternalId].(string)
	if strings.TrimSpace(tenantExternalId) == srcTenant.TenantExternalId() {
		tenantHasChanged = true
	}

	if !tenantHasChanged {
		return oauth2.NewInvalidGrantError("cannot switch to same tenant")
	}

	return nil
}
