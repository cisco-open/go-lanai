package scope

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	securityint "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope/internal_test"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	oauth22 "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	test "cto-github.cisco.com/NFV-BU/go-lanai/test/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/utils/testapp"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/utils/testscope"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/utils/testsuite"
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
	Revoker internal_test.MockedTokenRevoker
	Counter internal_test.InvocationCounter `optional:"true"`
}

func TestMain(m *testing.M) {
	testsuite.RunTests(m,
		testsuite.TestOptions(
			testapp.WithModules(Module),
		),
	)
}

/*************************
	Test Cases
 *************************/

func TestScopeManagerBasicBehavior(t *testing.T) {
	di := ManagerTestDI{}
	test.RunTest(context.Background(), t,
		testapp.Bootstrap(),
		testapp.WithFxOptions(
			appconfig.FxEmbeddedApplicationAdHoc(testAcctsFS),
			appconfig.FxEmbeddedApplicationAdHoc(testBasicFS),
			fx.Provide(securityint.BindSecurityIntegrationProperties),
			fx.Provide(internal_test.ProvideScopeMocks),
		),
		testapp.WithDI(&di),
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
		testapp.Bootstrap(),
		testapp.WithFxOptions(
			appconfig.FxEmbeddedApplicationAdHoc(testAcctsFS),
			appconfig.FxEmbeddedApplicationAdHoc(testAltFS),
			fx.Provide(securityint.BindSecurityIntegrationProperties),
			fx.Provide(internal_test.ProvideScopeMocksWithCounter),
		),
		testapp.WithDI(&di),
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

/*************************
	Sub-Test Cases
 *************************/

/* System Accounts */

func SubTestSysAcctLogin() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
		e := Do(ctx, func(ctx context.Context) {
			doAssertCurrentScope(ctx, g, "SysAcctLogin",
				assertAuthenticated(),
				assertWithUser(systemUsername, systemUserId),
				assertWithTenant(defaultTenantId, defaultTenantName),
				assertNotProxyAuth(),
				assertValidityGreaterThan(validity),
			)
		}, UseSystemAccount())
		g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
	}
}

func SubTestSysAcctWithTenant() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// use tenantId
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "SysAcct+TenantId",
					assertAuthenticated(),
					assertWithUser(systemUsername, systemUserId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, UseSystemAccount(), WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantName and existing auth
			ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "SysAcct+TenantName",
					assertAuthenticated(),
					assertWithUser(systemUsername, systemUserId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, UseSystemAccount(), WithTenantName(tenantName))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

/* Switch User using SysAcct */

func SubTestSwitchUserUsingSysAcct() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// use username
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+Username",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUsername(username), UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use user ID and existing auth
			ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+UserId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUserId(userId), UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

func SubTestSwitchUserWithTenantUsingSysAcct() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// use tenant ID
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+Username+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUsername(username), WithTenantId(tenantId), UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenent Name and existing auth
			ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SysAcct+Username+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(systemUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUserId(userId), WithTenantName(tenantName), UseSystemAccount())
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

/* Switch User */

func SubTestSwitchUser() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = testscope.WithMockedSecurity(ctx, securityMockAdmin())
		{
			// use username
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUsername(username))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use userId
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+UserId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUserId(userId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
	}
}

func SubTestSwitchUserWithTenant() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = testscope.WithMockedSecurity(ctx, securityMockAdmin())
		{
			// use tenantId
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUsername(username), WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantName
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+Username+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertProxyAuth(adminUsername),
					assertValidityGreaterThan(validity),
				)
			}, WithUsername(username), WithTenantName(tenantName))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}

	}
}

/* Switch Tenant */

func SubTestSwitchTenant() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
		{
			// use tenantId
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, WithTenantId(tenantId))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}
		{
			// use tenantName
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityGreaterThan(validity),
				)
			}, WithTenantName(tenantName))
			g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		}

	}
}

/* Timing */

func SubTestBackoffOnError(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
		{
			// first invocation
			e := Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should be invoked once")
		}

		{
			// immediate replay
			e := Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should not be invoked before backoff ")
		}
		{
			// wait and replay
			time.Sleep(200 * time.Millisecond)
			e := Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, WithTenantId(badTenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(2), "AuthenticationClient should not invoked again after backoff ")
		}
	}
}

func SubTestValidityNotGuaranteed(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
		{
			// invocation
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityLessThan(validity),
				)
			}, WithTenantId(tenantId))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient should be invoked once")
		}
		{
			// immediate replay
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Replay Switch+TenantName",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
					assertValidityLessThan(validity),
				)
			}, WithTenantId(tenantId))

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
			ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
			e := Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, WithTenantId(tenantId), UseSystemAccount())

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
	}
}

func SubTestWithoutCurrentAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		{
			// first invocation
			e := Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, WithTenantId(tenantId))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
	}
}

func SubTestInsufficientAccess() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
		{
			// cannot switch tenant
			e := Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, WithTenantId(badTenantName))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
		{
			// cannot switch tenant
			e := Do(ctx, func(ctx context.Context) {
				t.Errorf("scoped function should be be invoked in case of error")
			}, WithUsername(adminUsername))

			g.Expect(e).To(Not(Succeed()), "scope manager should returns error")
		}
	}
}

func SubTestSwitchToSameUser(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.Counter.ResetAll()
		{
			// first invocation
			ctx = testscope.WithMockedSecurity(ctx, securityMockAdmin())
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SameUsername",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(defaultTenantId, defaultTenantName),
					assertNotProxyAuth(),
				)
			}, WithUsername(adminUsername))

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
			ctx = testscope.WithMockedSecurity(ctx, securityMockRegular())
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+SameTenantId",
					assertAuthenticated(),
					assertWithUser(username, userId),
					assertWithTenant(defaultTenantId, defaultTenantName),
					assertNotProxyAuth(),
				)
			}, WithTenantId(defaultTenantId))

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
			ctx = testscope.WithMockedSecurity(ctx, securityMockAdmin())
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
				)
				toBeRevoked = security.Get(ctx).(oauth22.Authentication).AccessToken().Value()
			}, WithTenantId(tenantId))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(1), "AuthenticationClient.SwitchTenant should be invoked")
		}
		{
			// revoke and try again
			di.Revoker.Revoke(toBeRevoked)
			ctx = testscope.WithMockedSecurity(ctx, securityMockAdmin())
			e := Do(ctx, func(ctx context.Context) {
				doAssertCurrentScope(ctx, g, "Switch+TenantId",
					assertAuthenticated(),
					assertWithUser(adminUsername, adminUserId),
					assertWithTenant(tenantId, tenantName),
					assertNotProxyAuth(),
				)
			}, WithTenantId(tenantId))

			g.Expect(e).To(Succeed(), "scope manager should not returns error")
			g.Expect(di.Counter.Get(seclient.AuthenticationClient.SwitchTenant)).
				To(Equal(2), "AuthenticationClient.SwitchTenant should be invoked again after token revoke")
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
		oauth2, ok := auth.(oauth22.Authentication)
		g.Expect(ok).To(BeTrue(), "[%s] should oauth2.Authentication", msg)
		g.Expect(oauth2.AccessToken()).To(Not(BeNil()), "[%s] should contains access token", msg)
		expected := time.Now().Add(validity)
		g.Expect(oauth2.AccessToken().ExpiryTime().Before(expected)).To(BeTrue(), "[%s] should be valid less than %v", msg, validity)
	}
}

func assertValidityGreaterThan(validity time.Duration) assertion {
	return func(g *WithT, auth security.Authentication, msg string) {
		oauth2, ok := auth.(oauth22.Authentication)
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

func securityMockAdmin() testscope.SecurityMockOptions {
	return func(d *testscope.SecurityDetailsMock) {
		d.Username = adminUsername
		d.UserId = adminUserId
		d.TenantId = defaultTenantId
		d.TenantName = defaultTenantName
		d.Permissions = utils.NewStringSet(
			security.SpecialPermissionSwitchUser,
			security.SpecialPermissionSwitchTenant)
	}
}

func securityMockRegular() testscope.SecurityMockOptions {
	return func(d *testscope.SecurityDetailsMock) {
		d.Username = username
		d.UserId = userId
		d.TenantId = defaultTenantId
		d.TenantName = defaultTenantName
		d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
	}
}

//func (c *cache) loadFunc() loadFunc {
//	return func(ctx context.Context, k cKey) (v entryValue, exp time.Time, err error) {
//		fmt.Printf("loading key-%v...\n", k)
//		time.Sleep(1 * time.Second)
//		if k % 2 == 0 {
//			// happy path valid 5 seconds
//			valid := 5 * time.Second
//			exp = time.Now().Add(valid)
//			v = oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
//				opt.Token = oauth2.NewDefaultAccessToken("My test token")
//			})
//			fmt.Printf("loaded key-%v=%v with exp in %v\n", k, v.AccessToken().Value(), valid)
//		} else {
//			// unhappy path valid 2 seconds
//			valid := 2 * time.Second
//			exp = time.Now().Add(valid)
//			err = fmt.Errorf("oops")
//			fmt.Printf("loaded key-%v=%v with exp in %v\n", k, err, valid)
//		}
//		return
//	}
//}

//func (c *cache) evictFunc() gcache.EvictedFunc {
//	return func(k interface{}, v interface{}) {
//		fmt.Printf("evicted %v=%v\n", k, v)
//	}
//}
