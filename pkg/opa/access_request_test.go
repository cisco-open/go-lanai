package opa

import (
	"context"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test Setup
 *************************/

/*************************
	Test
 *************************/

func TestAllowRequest(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(BundleServerProvider()),
			fx.Invoke(opatestserver.InitializeBundleServer),
		),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestRequestWithPermission(di), "TestRequestWithPermission"),
		test.GomegaSubTest(SubTestRequestWithoutPermission(di), "TestRequestWithoutPermission"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestRequestWithPermission(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, memberAdminOptions())
		// member admin - can read
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.ExtraData["debug"] = "test"
		})
		g.Expect(e).To(Succeed())
		// member admin - can write
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())
	}
}

func SubTestRequestWithoutPermission(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// owner - can read
		ctx = sectest.ContextWithSecurity(ctx, memberOwnerOptions())
		// member user - can read
		e = AllowResource(ctx, "poc", OpRead, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())

		// owner - can write
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())
	}
}
