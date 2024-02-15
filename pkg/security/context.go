// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/internal"
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

type GlobalSettingReader interface {
	// Read setting of given key into "dest". Should support types:
	// 	- *[]byte
	// 	- *string
	// 	- *bool
	//	- *int
	Read(ctx context.Context, key string, dest interface{}) error
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
		secCtx = EmptyAuthentication("not authenticated")
	}
	return secCtx
}

// MustSet is the panicking version of Set.
func MustSet(ctx context.Context, auth Authentication) {
	if e := Set(ctx, auth); e != nil {
		panic(e)
	}
}

// Set security context, return error if the given context is not backed by utils.MutableContext.
func Set(ctx context.Context, auth Authentication) error {
	mc := utils.FindMutableContext(ctx)
	if mc == nil {
		return fmt.Errorf(`unable to set security into context: given context [%T] is not mutable`, ctx)
	}
	mc.Set(ContextKeySecurity, auth)

	// optionally, set AuthUserKey into gin context if available
	if gc := web.GinContext(ctx); gc != nil {
		if auth == nil {
			gc.Set(gin.AuthUserKey, nil)
		} else {
			gc.Set(gin.AuthUserKey, auth.Principal())
		}
	}
	return nil
}

// MustClear set security context as "unauthenticated".
func MustClear(ctx context.Context) {
	if e := Clear(ctx); e != nil {
		panic(e)
	}
}

// Clear attempt to set security context as "unauthenticated". Return error if not possible
func Clear(ctx context.Context) error {
	return Set(ctx, EmptyAuthentication("not authenticated"))
}

func HasPermissions(auth Authentication, permissions ...string) bool {
	for _, p := range permissions {
		if !auth.Permissions().Has(p) {
			return false
		}
	}
	return true
}

// IsTenantValid In most cases, the HasAccessToTenant should be used instead. It checks both the tenant's validity and whether the user has access to it
func IsTenantValid(ctx context.Context, tenantId string) bool {
	parentId, err := tenancy.GetParent(ctx, tenantId)
	//if we find a parent, that means we have this tenantId in tenant hierarchy, so it's valid
	if err == nil && parentId != "" {
		return true
	}

	//it's also possible that the tenantId is the root tenant (root tenant doesn't have a parent so it won't appear in the call above)
	rootId, err := tenancy.GetRoot(ctx)
	if err == nil && rootId != "" && rootId == tenantId {
		return true
	}

	return false
}

// HasAccessToTenant if no error return true, otherwise return false
func HasAccessToTenant(ctx context.Context, tenantId string) bool {
	err := HasErrorAccessingTenant(ctx, tenantId)
	return err == nil
}

// HasErrorAccessingTenant
/*
	if the tenantId is not valid, this method will return false, otherwise the following checks are applied in order

	1. If the user has ACCESS_ALL_TENANT permission, this method will return true

	2. If the user's designated tenants include the give tenant, this method will return true

	3. If the tenant hierarchy is loaded, this method will also check if any of the given tenant's ancestor
	is in the user's designated tenant. If yes, this method will return true.

	otherwise, this method return false.
*/
func HasErrorAccessingTenant(ctx context.Context, tenantId string) error {
	if !IsTenantValid(ctx, tenantId) {
		return ErrorInvalidTenantId
	}

	auth := Get(ctx)
	if ud, ok := auth.Details().(internal.TenantAccessDetails); ok {
		if ud.EffectiveAssignedTenantIds().Has(SpecialTenantIdWildcard) {
			return nil
		}

		hasDesc := tenancy.AnyHasDescendant(ctx, ud.EffectiveAssignedTenantIds(), tenantId)
		if hasDesc {
			return nil
		}
	}
	return ErrorTenantAccessDenied
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
