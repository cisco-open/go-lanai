package manager_test

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	securityint "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"embed"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

const (
	defaultTenantId   = "id-tenant-1"
	defaultTenantName = "tenant-1"
	systemUsername    = "system"
	systemUserId      = "id-system"

	adminUsername = "admin"
	adminUserId   = "id-admin"
	username      = "regular"
	userId        = "id-regular"

	tenantId      = "id-tenant-2"
	tenantName    = "tenant-2"
	badTenantId   = "id-tenant-3"
	badTenantName = "tenant-3"

	validity = 10 * time.Second
)

/*************************
	Test Main Setup
 *************************/

//go:embed manager_accts_test.yml
var testAcctsFS embed.FS

//go:embed manager_basic_test.yml
var testBasicFS embed.FS

//go:embed manager_alt_test.yml
var testAltFS embed.FS

type ManagerTestDI struct {
	fx.In
	Revoker sectest.MockedTokenRevoker
	Counter InvocationCounter `optional:"true"`
}

func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		suitetest.TestOptions(
			apptest.WithModules(scope.Module),
		),
	)
}

/*************************
	Test Cases
 *************************/

func TestScopeManagerBasicBehavior(t *testing.T) {
	di := ManagerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di),
		sectest.WithMockedScopes(testAcctsFS, testBasicFS),
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
		apptest.WithFxOptions(
			appconfig.FxEmbeddedApplicationAdHoc(testAcctsFS),
			appconfig.FxEmbeddedApplicationAdHoc(testAltFS),
			fx.Provide(securityint.BindSecurityIntegrationProperties),
			fx.Provide(provideScopeMocksWithCounter),
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
		apptest.WithFxOptions(
			fx.Provide(securityint.BindSecurityIntegrationProperties),
			fx.Provide(provideNoopScopeManager),
		),
		test.GomegaSubTest(SubTestNoopScopeManager(), "VerifyNoopScopeManager"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

/* System Accounts */

func SubTestSysAcctLogin() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
		e := scope.Do(ctx, func(ctx context.Context) {
			doAssertCurrentScope(ctx, g, "SysAcctLogin",
				assertAuthenticated(),
				assertWithUser(systemUsername, systemUserId),
				assertWithTenant(defaultTenantId, defaultTenantName),
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
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.UseSystemAccount(), scope.WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantName and existing auth
			ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "SysAcct+TenantName",
					assertAuthenticated(),
					assertWithUser(systemUsername, systemUserId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.UseSystemAccount(), scope.WithTenantName(tenantName))
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
					assertWithTenant(defaultTenantId, defaultTenantName),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use user ID and existing auth
			ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+UserId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
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
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.WithTenantId(tenantId), scope.UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenent Name and existing auth
			ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+Username+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUserId(userId), scope.WithTenantName(tenantName), scope.UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

/* Switch User */

func SubTestSwitchUser() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.WithMockedSecurity(ctx, securityMockAdmin())
		{
			// use username
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
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
					assertWithTenant(defaultTenantId, defaultTenantName),
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
		ctx = sectest.WithMockedSecurity(ctx, securityMockAdmin())
		{
			// use tenantId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantName
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithUsername(username), scope.WithTenantName(tenantName))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}

	}
}

/* Switch Tenant */

func SubTestSwitchTenant() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
		{
			// use tenantId
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantName
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, scope.WithTenantName(tenantName))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}

	}
}

/* Timing */

func SubTestBackoffOnError(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
		{
			// first invocation
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should be invoked once")
		}

		{
			// immediate replay
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
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
		ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
		{
			// invocation
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
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
				doAssertCurrentScope(ctx, g, "Replay Switch+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
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
			ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
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
		ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
		{
			// cannot switch tenant
			e := scope.Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, scope.WithTenantId(badTenantName))

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
			ctx = sectest.WithMockedSecurity(ctx, securityMockAdmin())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SameUsername",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(defaultTenantId, defaultTenantName),
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
			ctx = sectest.WithMockedSecurity(ctx, securityMockRegular())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SameTenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
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
			ctx = sectest.WithMockedSecurity(ctx, securityMockAdmin())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(tenantId, tenantName),
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
			ctx = sectest.WithMockedSecurity(ctx, securityMockAdmin())
			e := scope.Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(tenantId, tenantName),
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
			g.Expect(details.TenantName()).To(Equal(name), "[%s] should authenticated as tenant name [%s]", msg, name)
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
//		d.TenantName = defaultTenantName
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
		d.TenantName = defaultTenantName
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
		d.TenantName = defaultTenantName
		d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
	}
}

