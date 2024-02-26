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

package sectest

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

const (
	defaultTenantId   = "id-tenant-1"
	defaultTenantExternalId = "tenant-1"
	systemUsername    = "system"
	systemUserId      = "id-system"

	adminUsername = "admin"
	adminUserId   = "id-admin"
	username      = "regular"
	userId        = "id-regular"

	tenantId      = "id-tenant-2"
	tenantExternalId    = "tenant-2"
	badTenantId   = "id-tenant-3"
	badTenantExternalId = "tenant-3"

	validity = 10 * time.Second
)

/*************************
	Test Cases
 *************************/

func TestScopeManagerBasicBehavior(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedScopes(),
		test.GomegaSubTest(SubTestSysAcctLogin(), "SystemAccountLogin"),
		test.GomegaSubTest(SubTestSysAcctWithTenant(), "SystemAccountWithTenant"),
		test.GomegaSubTest(SubTestSwitchUserUsingSysAcct(), "SwitchUserUsingSysAcct"),
		test.GomegaSubTest(SubTestSwitchUserWithTenantUsingSysAcct(), "SwitchUserWithTenantUsingSysAcct"),
		test.GomegaSubTest(SubTestSwitchUser(), "SwitchUser"),
		test.GomegaSubTest(SubTestSwitchUserWithTenant(), "SwitchUserWithTenant"),
		test.GomegaSubTest(SubTestSwitchTenant(), "SwitchTenant"),
		test.GomegaSubTest(SubTestSwitchWithError(), "SwitchWithError"),
	)
}

/*************************
	Sub Tests
 *************************/

/* System Accounts */

func SubTestSysAcctLogin() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = WithMockedSecurity(ctx, securityMockRegular())
		e := scope.Do(ctx, func(ctx context.Context) {
			doAssertCurrentScope(ctx, g, "SysAcctLogin",
				assertAuthenticated(),
				assertWithUser(systemUsername, systemUserId),
				assertWithTenant(defaultTenantId, defaultTenantExternalId),
				assertNotProxyAuth(),
				assertValidityGreaterThan(validity),
			)
		}, scope.UseSystemAccount())
		g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
	}
}

func SubTestSysAcctWithTenant() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// use tenantId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "SysAcct+TenantId",
					assertAuthenticated(),
					assertWithUser(systemUsername, systemUserId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.UseSystemAccount(), scope.WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantExternalId and existing auth
			ctx = WithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "SysAcct+TenantExternalId",
					assertAuthenticated(),
					assertWithUser(systemUsername, systemUserId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.UseSystemAccount(), scope.WithTenantExternalId(tenantExternalId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

/* Switch User using SysAcct */

func SubTestSwitchUserUsingSysAcct() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// use username
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+Username",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantExternalId),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use user ID and existing auth
			ctx = WithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+UserId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantExternalId),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUserId(userId), scope.UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

func SubTestSwitchUserWithTenantUsingSysAcct() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// use tenant ID
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+Username+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.WithTenantId(tenantId), scope.UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenent Name and existing auth
			ctx = WithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+Username+TenantExternalId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUserId(userId), scope.WithTenantExternalId(tenantExternalId), scope.UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

/* Switch User */

func SubTestSwitchUser() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = WithMockedSecurity(ctx, securityMockAdmin())
		{
			// use username
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantExternalId),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use userId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+UserId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantExternalId),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUserId(userId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

func SubTestSwitchUserWithTenant() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = WithMockedSecurity(ctx, securityMockAdmin())
		{
			// use tenantId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantExternalId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username+TenantExternalId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.WithTenantExternalId(tenantExternalId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}

	}
}

/* Switch Tenant */

func SubTestSwitchTenant() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = WithMockedSecurity(ctx, securityMockRegular())
		{
			// use tenantId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantExternalId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantExternalId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithTenantExternalId(tenantExternalId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

/* Error Case */

func SubTestSwitchWithError() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = WithMockedSecurity(ctx, securityMockRegular())
		{
			// use tenantId
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}

		{
			// use tenantExternalId
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantExternalId(badTenantExternalId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
	}
}

/*************************
	Helpers
 *************************/

type assertion func(g *gomega.WithT, auth security.Authentication, msg string)

func doAssertCurrentScope(ctx context.Context, g *gomega.WithT, msg string, assertions ...assertion) {
	auth := security.Get(ctx)
	for _, fn := range assertions {
		fn(g, auth, msg)
	}
}

func assertAuthenticated() assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		g.Expect(auth.State()).To(Equal(security.StateAuthenticated), "[%s] security state should be authenticated within scope", msg)
	}
}

func assertWithUser(username, uid string) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		if username != "" {
			g.Expect(auth.Principal()).To(Equal(username), "[%s] should authenticated as username [%s]", msg, username)
		}

		details, ok := auth.Details().(security.UserDetails)
		g.Expect(ok).To(BeTrue(), "[%s] auth details should be UserDetails", msg)
		if uid != "" {
			g.Expect(details.UserId()).To(Equal(uid), "[%s] should authenticated as user ID [%s]", msg, uid)
		}
	}
}

func assertWithTenant(id, name string) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		details, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeTrue(), "[%s] auth details should be TenantDetails", msg)

		if id != "" {
			g.Expect(details.TenantId()).To(Equal(id), "[%s] should authenticated as tenant ID [%s]", msg, id)
		}

		if name != "" {
			g.Expect(details.TenantExternalId()).To(Equal(name), "[%s] should authenticated as tenant name [%s]", msg, name)
		}
	}
}

func assertNotProxyAuth() assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		details, ok := auth.Details().(security.ProxiedUserDetails)
		g.Expect(ok).To(BeTrue(), "[%s] auth details should be ProxiedUserDetails", msg)
		g.Expect(details.Proxied()).To(BeFalse(), "[%s] should not be proxy auth", msg)
	}
}

func assertProxyAuth(origName string) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		details, ok := auth.Details().(security.ProxiedUserDetails)
		g.Expect(ok).To(BeTrue(), "[%s] auth details should be ProxiedUserDetails", msg)

		g.Expect(details.Proxied()).To(BeTrue(), "[%s] should be proxy auth", msg)
		if origName != "" {
			g.Expect(details.OriginalUsername()).To(Equal(origName), "[%s] should be proxy auth with original username", msg, origName)
		}
	}
}

func assertValidityGreaterThan(validity time.Duration) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		oauth, ok := auth.(oauth2.Authentication)
		g.Expect(ok).To(BeTrue(), "[%s] should oauth2.Authentication", msg)
		g.Expect(oauth.AccessToken()).To(Not(BeNil()), "[%s] should contains access token", msg)
		expected := time.Now().Add(validity)
		g.Expect(oauth.AccessToken().ExpiryTime().After(expected)).To(BeTrue(), "[%s] should be valid greater than %v", msg, validity)
	}
}

func securityMockAdmin() SecurityMockOptions {
	return func(d *SecurityDetailsMock) {
		d.Username = adminUsername
		d.UserId = adminUserId
		d.TenantId = defaultTenantId
		d.TenantExternalId = defaultTenantExternalId
		d.Permissions = utils.NewStringSet(
			security.SpecialPermissionSwitchUser,
			security.SpecialPermissionSwitchTenant)
	}
}

func securityMockRegular() SecurityMockOptions {
	return func(d *SecurityDetailsMock) {
		d.Username = username
		d.UserId = userId
		d.TenantId = defaultTenantId
		d.TenantExternalId = defaultTenantExternalId
		d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
	}
}

