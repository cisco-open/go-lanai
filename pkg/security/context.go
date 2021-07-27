package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

const (
	HighestMiddlewareOrder = int(-1<<18 + 1)                 // -0x3ffff = -262143
	LowestMiddlewareOrder  = HighestMiddlewareOrder + 0xffff // -0x30000 = -196608
)

type AuthenticationState int

const (
	StateAnonymous = AuthenticationState(iota)
	StatePrincipalKnown
	StateAuthenticated
)

type Permissions map[string]interface{}

func (p Permissions) Has(permission string) bool {
	_, ok := p[permission]
	return ok
}

type Authentication interface {
	Principal() interface{}
	Permissions() Permissions
	State() AuthenticationState
	Details() interface{}
}

// EmptyAuthentication represent unauthenticated user.
// Note: anonymous user is considered authenticated
type EmptyAuthentication string

func (EmptyAuthentication) Principal() interface{} {
	return nil
}

func (EmptyAuthentication) Details() interface{} {
	return nil
}

func (EmptyAuthentication) State() AuthenticationState {
	return StateAnonymous
}

func (EmptyAuthentication) Permissions() Permissions {
	return map[string]interface{}{}
}

func GobRegister() {
	gob.Register(EmptyAuthentication(""))
	gob.Register((*AnonymousAuthentication)(nil))
	gob.Register((*CodedError)(nil))
	gob.Register(errors.New(""))
	gob.Register((*DefaultAccount)(nil))
	gob.Register((*AcctDetails)(nil))
	gob.Register((*AcctLockingRule)(nil))
	gob.Register((*AcctPasswordPolicy)(nil))
	gob.Register((*AccountMetadata)(nil))
}

/**********************************
	Common Functions
 **********************************/
func Get(ctx context.Context) Authentication {
	secCtx, ok := ctx.Value(ContextKeySecurity).(Authentication)
	if !ok {
		secCtx = EmptyAuthentication("EmptyAuthentication")
	}
	return secCtx
}

func Clear(ctx context.Context) {
	if mc, ok := ctx.(utils.MutableContext); ok {
		mc.Set(gin.AuthUserKey, nil)
		mc.Set(ContextKeySecurity, nil)
	}

	if gc := web.GinContext(ctx); gc != nil {
		gc.Set(gin.AuthUserKey, nil)
		gc.Set(ContextKeySecurity, nil)
	}
}

// TryClear attempt to clear security context. Return true if succeeded
func TryClear(ctx context.Context) bool {
	switch ctx.(type) {
	case utils.MutableContext:
		Clear(ctx.(utils.MutableContext))
	default:
		return false
	}
	return true
}

func HasPermissions(auth Authentication, permissions ...string) bool {
	for _, p := range permissions {
		if !auth.Permissions().Has(p) {
			return false
		}
	}
	return true
}

//In most cases, the HasAccessToTenant should be used instead. It checks both the tenant's validity and whether the user has access to it
func IsTenantValid(ctx context.Context, tenantId string) bool {
	parentId, err := tenancy.GetParent(ctx, tenantId)
	//if we find a parent, that means we have this tenantId in tenant hierarchy, so it's valid
	if err == nil && parentId != ""  {
		return true
	}

	//it's also possible that the tenantId is the root tenant (root tenant doesn't have a parent so it won't appear in the call above)
	rootId, err := tenancy.GetRoot(ctx)
	if err == nil && rootId != "" && rootId == tenantId {
		return true
	}

	return false
}

/*
	if the tenantId is not valid, this method will return false, otherwise the following checks are applied in order

	1. If the user has ACCESS_ALL_TENANT permission, this method will return true

	2. If the user's designated tenants include the give tenant, this method will return true

	3. If the tenant hierarchy is loaded, this method will also check if any of the given tenant's ancestor
	is in the user's designated tenant. If yes, this method will return true.

	otherwise, this method return false.
 */
func HasAccessToTenant(ctx context.Context, tenantId string) bool {
	if !IsTenantValid(ctx, tenantId) {
		return false
	}

	auth := Get(ctx)

	if HasPermissions(auth, SpecialPermissionAccessAllTenant) {
		return true
	}

	if ud, ok := auth.Details().(UserDetails); ok {
		if ud.AssignedTenantIds().Has(tenantId) {
			return true
		}

		if !tenancy.IsLoaded(ctx) {
			logger.Warnf("Tenant hierarchy is not loaded by the auth server, hasAccessToTenant will not consider child tenants in the tenant hierarchy")
			return false
		}

		ancestors, err := tenancy.GetAncestors(ctx, tenantId)
		if err != nil {
			return false
		}

		for _, ancestor := range ancestors {
			if ud.AssignedTenantIds().Has(ancestor) {
				return true
			}
		}
	}
	return false
}

func IsFullyAuthenticated(auth Authentication) bool {
	return auth.State() >= StateAuthenticated
}

func IsBeingAuthenticated(from, to Authentication) bool {
	fromUnauthenticatedState := from == nil || from.State() < StateAuthenticated
	toAuthenticatedState := to != nil && to.State() > StatePrincipalKnown
	return fromUnauthenticatedState && toAuthenticatedState
}

func IsBeingUnAuthenticated(from, to Authentication) bool {
	fromAuthenticated := from != nil && from.State() > StateAnonymous
	toUnAuthenticatedState := to == nil || to.State() <= StateAnonymous
	return fromAuthenticated && toUnAuthenticatedState
}

func DetermineAuthenticationTime(_ context.Context, userAuth Authentication) (authTime time.Time) {
	if userAuth == nil {
		return
	}

	details, ok := userAuth.Details().(map[string]interface{})
	if !ok {
		return
	}

	v, ok := details[DetailsKeyAuthTime]
	if !ok {
		return
	}

	switch t := v.(type) {
	case time.Time:
		authTime = t
	case string:
		authTime = utils.ParseTime(time.RFC3339, t)
	}
	return
}

func GetUsername(userAuth Authentication) (string, error) {
	if userAuth == nil {
		return "", fmt.Errorf("unsupported authentication is nil")
	}

	principal := userAuth.Principal()
	var username string
	switch principal.(type) {
	case Account:
		username = principal.(Account).Username()
	case string:
		username = principal.(string)
	case fmt.Stringer:
		username = principal.(fmt.Stringer).String()
	default:
		return "", fmt.Errorf("unsupported principal type %T", principal)
	}
	return username, nil
}