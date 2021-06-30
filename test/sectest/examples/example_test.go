package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"embed"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

/*************************
	Examples
 *************************/

type TestTarget struct{}

func (t *TestTarget) DoSomethingWithinSecurityScope(ctx context.Context) error {
	e := scope.Do(ctx, func(scopedCtx context.Context) {
		// scopedCtx contains switched security context
		// do something with scopedCtx...
		_ = t.DoSomethingRequiringSecurity(scopedCtx)
	}, scope.UseSystemAccount())
	return e
}

func (t *TestTarget) DoSomethingRequiringSecurity(ctx context.Context) error {
	auth := security.Get(ctx)
	if !security.IsFullyAuthenticated(auth) {
		return fmt.Errorf("not authenticated")
	}
	return nil
}

// TestUseDefaultSecurityScopeMocking
// apptest.Bootstrap and sectest.WithMockedScopes are required for usage of scope package
func TestUseDefaultSecurityScopeMocking(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		sectest.WithMockedScopes(),
		test.GomegaSubTest(SubTestExampleUseScope(), "UseScope"),
	)
	// Any sub tests can use "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope" as usual
}

//go:embed example-custom-scopes.yml
var customMockingConfigFS embed.FS

// TestUseCustomSecurityScopeMocking
// apptest.Bootstrap and sectest.WithMockedScopes are required for usage of scope package
func TestUseCustomSecurityScopeMocking(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		sectest.WithMockedScopes(customMockingConfigFS), // use custom config as embeded configuration
		test.GomegaSubTest(SubTestExampleUseScope(), "UseScope"),
	)
	// Any sub tests can use "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope" as usual
}

// TestCurrentSecurityContextMocking
// apptest.Bootstrap and sectest.WithMockedScopes are NOT required for usage of sectest.WithMockedSecurity
func TestCurrentSecurityContextMocking(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestExampleMockCurrentSecurity(), "MockCurrentSecurity"),
	)
}

// TestMockBothCurrentSecurityAndScope
// apptest.Bootstrap and sectest.WithMockedScopes are NOT required for usage of sectest.WithMockedSecurity
func TestMockBothCurrentSecurityAndScope(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestExampleMockBoth(), "MockBoth"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleUseScope() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		toTest := &TestTarget{}
		e := toTest.DoSomethingWithinSecurityScope(ctx)
		g.Expect(e).To(Succeed(), "scoped operation shouldn't returns error")
	}
}

func SubTestExampleMockCurrentSecurity() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.WithMockedSecurity(ctx, func(d *sectest.SecurityDetailsMock) {
			d.Username = "any-username"
			d.UserId = "any-user-id"
			d.TenantId = "any-tenant-id"
			d.TenantName = "any-tenant-name"
			d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
			// see sectest.SecurityDetailsMock for more options
		})

		toTest := &TestTarget{}
		e := toTest.DoSomethingRequiringSecurity(ctx)
		g.Expect(e).To(Succeed(), "methods requiring security shouldn't returns error")
	}
}

func SubTestExampleMockBoth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// combined usage
		ctx = sectest.WithMockedSecurity(ctx, func(d *sectest.SecurityDetailsMock) {
			d.Username = "any-username"
			d.UserId = "any-user-id"
			d.TenantId = "any-tenant-id"
			d.TenantName = "any-tenant-name"
			d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
			// see sectest.SecurityDetailsMock for more options
		})

		toTest := &TestTarget{}
		e := toTest.DoSomethingWithinSecurityScope(ctx)
		g.Expect(e).To(Succeed(), "security-aware methods shouldn't returns error")
	}
}
