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

package scope_test

import (
	"context"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	securityint "github.com/cisco-open/go-lanai/pkg/integrate/security"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope/testdata"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

const (
	defaultTenantId         = "id-tenant-1"
	defaultTenantExternalId = "tenant-1"
	systemUsername          = "system"
	systemUserId            = "id-system"

	adminUsername = "admin"
	adminUserId   = "id-admin"
	username      = "regular"
	userId        = "id-regular"

	tenantId            = "id-tenant-2"
	tenantExternalId    = "tenant-2"
	badTenantId         = "id-tenant-3"
	badTenantExternalId = "tenant-3"

	validity = 10 * time.Second
)

/*************************
	Test Main Setup
 *************************/

type ManagerTestDI struct {
	fx.In
	Revoker sectest.MockedTokenRevoker
	Counter InvocationCounter `optional:"true"`
}

/*************************
	Test Cases
 *************************/

func TestScopeManagerBasicBehavior(t *testing.T) {
	di := ManagerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(scope.Module),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(NewCounter, scope.FxManagerCustomizer(NewCustomizer)),
		),
		sectest.WithMockedScopes(testdata.TestAcctsFS, testdata.TestBasicFS),
		test.GomegaSubTest(SubTestHookInvocation(&di), "HookInvocation"),
		test.GomegaSubTest(SubTestSysAcctLogin(), "SystemAccountLogin"),
		test.GomegaSubTest(SubTestSysAcctWithTenant(), "SystemAccountWithTenant"),
		test.GomegaSubTest(SubTestSwitchUserUsingSysAcct(), "SwitchUserUsingSysAcct"),
		test.GomegaSubTest(SubTestSwitchUserWithTenantUsingSysAcct(), "SwitchUserWithTenantUsingSysAcct"),
		test.GomegaSubTest(SubTestSwitchUser(), "SwitchUser"),
		test.GomegaSubTest(SubTestSwitchUserWithTenant(), "SwitchUserWithTenant"),
		test.GomegaSubTest(SubTestSwitchTenant(), "SwitchTenant"),
	)
}

func TestScopeManagerWithAltSettings(t *testing.T) {
	di := ManagerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(scope.Module),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedApplicationAdHoc(testdata.TestAcctsFS),
			appconfig.FxEmbeddedApplicationAdHoc(testdata.TestAltFS),
			fx.Provide(securityint.BindSecurityIntegrationProperties),
			fx.Provide(ProvideScopeMocksWithCounter),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestBackoffOnError(&di), "BackoffOnError"),
		test.GomegaSubTest(SubTestValidityNotGuaranteed(&di), "ValidityNotGuaranteed"),
		test.GomegaSubTest(SubTestWithoutSysAcctConfig(), "WithoutSysAcctConfig"),
		test.GomegaSubTest(SubTestWithoutCurrentAuth(), "WithoutCurrentAuth"),
		test.GomegaSubTest(SubTestInsufficientAccess(), "InsufficientAccess"),
		test.GomegaSubTest(SubTestSwitchToSameUser(&di), "SwitchToSameUser"),
		test.GomegaSubTest(SubTestSwitchToSameTenant(&di), "SwitchToSameTenant"),
		test.GomegaSubTest(SubTestRevokedToken(&di), "RevokedToken"),
	)
}

func TestOverridingDefaultScopeManager(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(scope.Module),
		apptest.WithFxOptions(
			fx.Decorate(ProvideNoopScopeManager),
		),
		test.GomegaSubTest(SubTestNoopScopeManager(), "VerifyNoopScopeManager"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

/* System Accounts */

func SubTestHookInvocation(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(securityMockRegular()))
		di.Counter.ResetAll()
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
		e := scope.Do(ctx, func(ctx context.Context) {
			doAssertCurrentScope(ctx, g, "SysAcctLogin",
				assertAuthenticated(),
				assertWithUser(systemUsername, systemUserId),
				assertWithTenant(defaultTenantId, defaultTenantExternalId),
				assertNotProxyAuth(),
				assertValidityGreaterThan(validity),
			)
			g.Expect(di.Counter.Get(TestScopeManagerHook.Before)).To(BeNumerically(">", 0), "before hook should be executed")
		}, scope.UseSystemAccount())
		g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		g.Expect(di.Counter.Get(TestScopeManagerHook.After)).To(BeNumerically(">", 0), "before hook should be executed")
	}
}

func SubTestSysAcctLogin() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(securityMockRegular()))
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
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
			ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
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
			ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
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
			ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
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
		ctx = ContextWithMockedSecurity(ctx, securityMockAdmin())
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
		ctx = ContextWithMockedSecurity(ctx, securityMockAdmin())
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
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
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

/* Timing */

func SubTestBackoffOnError(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
		{
			// first invocation
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should not be invoked in case of error")
			}, scope.WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should be invoked once")
		}

		{
			// immediate replay
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should not be invoked in case of error")
			}, scope.WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should not be invoked before backoff ")
		}
		{
			// wait and replay
			time.Sleep(200 * time.Millisecond)
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(2), "AuthenticationClient should not invoked again after backoff ")
		}
	}
}

func SubTestValidityNotGuaranteed(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
		{
			// invocation
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantExternalId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
					assertValidityLessThan(validity),
				)
			}, scope.WithTenantId(tenantId))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should be invoked once")
		}
		{
			// immediate replay
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Replay Switch+TenantExternalId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
					assertValidityLessThan(validity),
				)
			}, scope.WithTenantId(tenantId))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should be invoked once")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should not be invoked repeatedly")
		}
	}
}

/* Error */

func SubTestWithoutSysAcctConfig() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// first invocation
			ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantId(tenantId), scope.UseSystemAccount())

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
	}
}

func SubTestWithoutCurrentAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// first invocation
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantId(tenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
	}
}

func SubTestInsufficientAccess() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
		{
			// cannot switch tenant
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantId(badTenantExternalId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
		{
			// cannot switch tenant
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithUsername(adminUsername))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
	}
}

func SubTestSwitchToSameUser(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		{
			// first invocation
			ctx = ContextWithMockedSecurity(ctx, securityMockAdmin())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SameUsername",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(defaultTenantId, defaultTenantExternalId),
					assertNotProxyAuth(),
				)
			}, scope.WithUsername(adminUsername))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchUser)).
				To(Equal(0), "AuthenticationClient.SwitchUser should not be invoked")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(0), "AuthenticationClient.SwitchTenant should not be invoked")
		}
	}
}

func SubTestSwitchToSameTenant(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		{
			// first invocation
			ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SameTenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantExternalId),
					assertNotProxyAuth(),
				)
			}, scope.WithTenantId(defaultTenantId))

			g.Expect(e).To(Succeed(), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(0), "AuthenticationClient should not be invoked")

		}
	}
}

func SubTestRevokedToken(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		var toBeRevoked string
		{
			// first invocation
			ctx = ContextWithMockedSecurity(ctx, securityMockAdmin())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
				)
				toBeRevoked = security.Get(ctx).(oauth2.Authentication).AccessToken().Value()
			}, scope.WithTenantId(tenantId))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient.SwitchTenant should be invoked")
		}
		{
			// revoke and try again
			di.Revoker.Revoke(toBeRevoked)
			ctx = ContextWithMockedSecurity(ctx, securityMockAdmin())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(tenantId, tenantExternalId),
					assertNotProxyAuth(),
				)
			}, scope.WithTenantId(tenantId))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(2), "AuthenticationClient.SwitchTenant should be invoked again after token revoke")
		}
	}
}

func SubTestNoopScopeManager() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := scope.Do(ctx, func(scoped context.Context) {
			g.Expect(scoped).To(BeIdenticalTo(ctx), "noop scope manager shouldn't do anything")
		})
		g.Expect(e).To(Succeed(), "noop scope manager shouldn't returns error")
	}
}

/*************************
	Helpers
 *************************/

type assertion func(g *gomega.WithT, auth security.Authentication, msg string)

func doAssertCurrentScope(ctx context.Context, g *gomega.WithT, msg string, assertions ...assertion) {
	g.Expect(scope.Describe(ctx)).ToNot(Equal("no scope"), "current scope should be available")
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

func assertWithTenant(id, externalId string) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		details, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeTrue(), "[%s] auth details should be TenantDetails", msg)

		if id != "" {
			g.Expect(details.TenantId()).To(Equal(id), "[%s] should authenticated as tenant ID [%s]", msg, id)
		}

		if externalId != "" {
			g.Expect(details.TenantExternalId()).To(Equal(externalId), "[%s] should authenticated as tenant externalId [%s]", msg, externalId)
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

func assertValidityLessThan(validity time.Duration) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		oauth2, ok := auth.(oauth2.Authentication)
		g.Expect(ok).To(BeTrue(), "[%s] should oauth2.Authentication", msg)
		g.Expect(oauth2.AccessToken()).To(Not(BeNil()), "[%s] should contains access token", msg)
		expected := time.Now().Add(validity)
		g.Expect(oauth2.AccessToken().ExpiryTime().Before(expected)).To(BeTrue(), "[%s] should be valid less than %v", msg, validity)
	}
}

func assertValidityGreaterThan(validity time.Duration) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		oauth2, ok := auth.(oauth2.Authentication)
		g.Expect(ok).To(BeTrue(), "[%s] should oauth2.Authentication", msg)
		g.Expect(oauth2.AccessToken()).To(Not(BeNil()), "[%s] should contains access token", msg)
		expected := time.Now().Add(validity)
		g.Expect(oauth2.AccessToken().ExpiryTime().After(expected)).To(BeTrue(), "[%s] should be valid greater than %v", msg, validity)
	}
}

//func securityMockSysAcct() testscope.SecurityMockOptions {
//	return func(d *testscope.SecurityDetailsMock) {
//		d.Username = systemUsername
//		d.UserId = systemUserId
//		d.TenantId = defaultTenantId
//		d.TenantExternalId = defaultTenantExternalId
//		d.Permissions = utils.NewStringSet(
//			security.SpecialPermissionAccessAllTenant,
//			security.SpecialPermissionSwitchUser,
//			security.SpecialPermissionSwitchTenant)
//	}
//}

func securityMockAdmin() sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Username = adminUsername
		d.UserId = adminUserId
		d.TenantId = defaultTenantId
		d.TenantExternalId = defaultTenantExternalId
		d.Permissions = utils.NewStringSet(
			security.SpecialPermissionSwitchUser,
			security.SpecialPermissionSwitchTenant)
	}
}

func securityMockRegular() sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Username = username
		d.UserId = userId
		d.TenantId = defaultTenantId
		d.TenantExternalId = defaultTenantExternalId
		d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
	}
}

func ContextWithMockedSecurity(ctx context.Context, opts ...sectest.SecurityMockOptions) context.Context {
	return sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(opts...))
}
