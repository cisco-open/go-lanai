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

package scope

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "time"
)

type ManagerOptions func(opt *managerOption)

type managerOption struct {
	Client               seclient.AuthenticationClient
	TokenStoreReader     oauth2.TokenStoreReader
	BackOffPeriod        time.Duration
	GuaranteedValidity   time.Duration
	KnownCredentials     map[string]string
	SystemAccounts       utils.StringSet
	DefaultSystemAccount string
	BeforeStartHooks     []ScopeOperationHook
	AfterEndHooks        []ScopeOperationHook
}

// defaultScopeManager always first attempt to login as system account and then switch to destination security context
type defaultScopeManager struct {
	managerBase
	client            seclient.AuthenticationClient
	knownCredentials  map[string]string
	systemAccounts    utils.StringSet
	defaultSysAcct    string
	defaultSysAcctKey cKey
}

func newDefaultScopeManager(opts ...ManagerOptions) *defaultScopeManager {
	opt := managerOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	return &defaultScopeManager{
		managerBase:          managerBase{
			cache:              newCache(),
			tokenStoreReader:   opt.TokenStoreReader,
			failureBackOff:     opt.BackOffPeriod,
			guaranteedValidity: opt.GuaranteedValidity,
			beforeStartHooks:   opt.BeforeStartHooks,
			afterEndHooks:      opt.AfterEndHooks,
		},
		client:           opt.Client,
		knownCredentials: opt.KnownCredentials,
		systemAccounts:   opt.SystemAccounts,
		defaultSysAcct: opt.DefaultSystemAccount,
		defaultSysAcctKey: cKey{
			username:   opt.DefaultSystemAccount,
		},
	}
}

func (m *defaultScopeManager) StartScope(ctx context.Context, scope *Scope) (context.Context, error) {
	if e := m.prepareScope(ctx, scope); e != nil {
		return nil, e
	}

	switch {
	case scope.useSysAcct:
		return m.managerBase.DoStartScope(ctx, scope, m.authWithSysAcct)
	default:
		return m.managerBase.DoStartScope(ctx, scope, m.authWithoutSysAcct)
	}
}

func (m *defaultScopeManager) Start(ctx context.Context, opts...Options) (context.Context, error) {
	scope := New(opts...)
	return m.StartScope(ctx, scope)
}

// prepareScope perform some validation and prepare scope with proper cache key and other attributes
func (m *defaultScopeManager) prepareScope(ctx context.Context, scope *Scope) error {
	if e := scope.validate(ctx); e != nil {
		return e
	}
	switch {
	case scope.useSysAcct:
		return m.prepareScopeWithSysAcct(ctx, scope)
	default:
		return m.prepareScopeWithoutSysAcct(ctx, scope)
	}
}

// prepareScopeWithSysAcct prepare scope with proper cache key and fill other default attributes.
// This mode ignores current authenticated user,
// and, if not specified, the target username is set to default system account
func (m *defaultScopeManager) prepareScopeWithSysAcct(ctx context.Context, scope *Scope) error {
	if scope.username == "" && scope.userId == "" {
		// user not specified
		if m.defaultSysAcct == "" {
			return ErrMissingDefaultSysAccount
		}
		scope.username = m.defaultSysAcct
	}

	currAuth := security.Get(ctx)
	m.normalizeTargetUser(currAuth, scope)
	m.prepareCacheKey(scope, "")
	return nil
}

// prepareScopeWithoutSysAcct prepare scope with proper cache key and fill other default attributes.
// This mode requires given context bears fully authenticated user, and the target username/userId is mandatory
func (m *defaultScopeManager) prepareScopeWithoutSysAcct(ctx context.Context, scope *Scope) error {
	currAuth := security.Get(ctx)
	currUsername, _, e := m.resolveUser(currAuth)
	if e != nil {
		return ErrNotCurrentlyAuthenticated
	}

	if scope.username == "" && scope.userId == "" {
		scope.username = currUsername
	}
	m.normalizeTargetUser(currAuth, scope)
	m.prepareCacheKey(scope, currUsername)
	return nil
}

// authWithSysAcct is an authenticateFunc which is invoked by loadFunc in a separate goroutine
// therefore it's safe to call managerBase.GetOrAuthenticate again without deadlocking.
// This auth method would try direct password login (if password is known),
// then fallback to 2 stepped context switching:
// 		1. try switch to default system account (may involve password login using system accoutn credentials)
// 		2. call switch user/tenant API with system account's access token
func (m *defaultScopeManager) authWithSysAcct(ctx context.Context, pKey *cKey) (security.Authentication, error) {
	if pKey == nil {
		return nil, fmt.Errorf("[Internal Error] cache key is nil")
	}

	// first, attempt password login
	if r, e := m.passwordLogin(ctx, pKey); e != nil {
		return nil, e
	} else if r != nil && r.Token != nil {
		return m.convertToAuthentication(ctx, r)
	}

	// then attempt to do switch context using system account
	// note that at this point, it's guaranteed that the given pKey is not default sys account key
	auth, e := m.GetOrAuthenticate(ctx, &m.defaultSysAcctKey, time.Now().UTC(), m.authWithSysAcct)
	if e != nil {
		return nil, e
	}

	r, e := m.switchContext(ctx, pKey, auth)
	if e != nil {
		return nil, e
	} else if r != nil && r.Token != nil {
		return m.convertToAuthentication(ctx, r)
	}
	return auth, nil
}

// authWithoutSysAcct is an authenticateFunc which is invoked by loadFunc in a separate goroutine
// therefore it's safe to call managerBase.GetOrAuthenticate again without deadlocking
// context switching by calling switch user/tenant API with current access token
func (m *defaultScopeManager) authWithoutSysAcct(ctx context.Context, pKey *cKey) (security.Authentication, error) {
	if pKey == nil {
		return nil, fmt.Errorf("[Internal Error] cache key is nil")
	} else if m.systemAccounts.Has(pKey.username) {
		return nil, fmt.Errorf("[Internal Error] cannot switch to system account without UseSystemAccount() option")
	}

	// attempt to use switch context with current auth
	auth := security.Get(ctx)
	r, e := m.switchContext(ctx, pKey, auth)
	if e != nil {
		return nil, e
	} else if r != nil && r.Token != nil {
		return m.convertToAuthentication(ctx, r)
	}
	//
	return auth, nil
}

func (m *defaultScopeManager) credentialsLookup(pKey *cKey) (password string, found bool) {
	if pKey.username == "" {
		return
	}
	password, found = m.knownCredentials[pKey.username]
	return
}

// passwordLogin perform password login if credentials is available.
// it returns nil, nil if no credentials is found
func (m *defaultScopeManager) passwordLogin(ctx context.Context, pKey *cKey) (*seclient.Result, error) {
	p, ok := m.credentialsLookup(pKey)
	if !ok {
		// password not available
		return nil, nil
	}

	authOpts := []seclient.AuthOptions{
		seclient.WithCredentials(pKey.username, p),
	}
	if pKey.tenantExternalId != "" || pKey.tenantId != "" {
		authOpts = append(authOpts, seclient.WithTenant(pKey.tenantId, pKey.tenantExternalId))
	}
	return m.client.PasswordLogin(ctx, authOpts...)
}

// switchContext perform switch user or switch tenant
// it returns nil, nil if target context is identical as given auth (same user and same tenant)
func (m *defaultScopeManager) switchContext(ctx context.Context, pKey *cKey, auth security.Authentication) (*seclient.Result, error) {
	if _, ok := m.credentialsLookup(pKey); ok {
		return nil, fmt.Errorf("user [%s] is configured to use password login only", pKey.username)
	}

	authOpts := []seclient.AuthOptions{
		seclient.WithAuthentication(auth),
	}
	if pKey.tenantExternalId != "" || pKey.tenantId != ""{
		authOpts = append(authOpts, seclient.WithTenant(pKey.tenantId, pKey.tenantExternalId))
	}

	if m.isSameUser(pKey.username, pKey.userId, auth) {
		// switch tenant
		if m.isSameTenant(pKey.tenantExternalId, pKey.tenantId, auth) {
			return nil, nil
		} else {
			return m.client.SwitchTenant(ctx, authOpts...)
		}
	} else {
		// switch user
		authOpts = append(authOpts, seclient.WithUser(pKey.userId, pKey.username))
		return m.client.SwitchUser(ctx, authOpts...)
	}
}