package opa

import (
	"context"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test
 *************************/

type testDI struct {
	fx.In
}

func TestAllowResource(t *testing.T) {
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
		test.GomegaSubTest(SubTestResourceBaseline(di), "TestResourceBaseline"),
		test.GomegaSubTest(SubTestMemberAdmin(di), "TestMemberAdmin"),
		test.GomegaSubTest(SubTestMemberOwner(di), "TestMemberOwner"),
		test.GomegaSubTest(SubTestMemberNonOwner(di), "TestMemberNonOwner"),
		test.GomegaSubTest(SubTestNonMember(di), "TestNonMember"),
		test.GomegaSubTest(SubTestSharedUser(di), "TestSharedUser"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestResourceBaseline(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = AllowResource(ctx, "doesn't matter", "whatever", func(res *Resource) {
			res.Policy = "baseline/allow"
			res.RawInput = map[string]interface{}{
				"just_data": "data",
			}
		})
		g.Expect(e).To(Succeed())
	}
}

func SubTestMemberAdmin(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
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

func SubTestMemberOwner(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// owner - can read
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
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

func SubTestMemberNonOwner(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// member user
		ctx = sectest.ContextWithSecurity(ctx, MemberNonOwnerOptions())
		// member user - can read
		e = AllowResource(ctx, "poc", OpRead, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())

		// member user - cannot write
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(HaveOccurred())
	}
}

func SubTestNonMember(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// non-member admin - can't read
		ctx = sectest.ContextWithSecurity(ctx, NonMemberAdminOptions())
		e = AllowResource(ctx, "poc", OpRead, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(HaveOccurred())
	}
}

func SubTestSharedUser(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, MemberNonOwnerOptions())
		// non-member user but shared - cannot write if not allowed
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.Sharing = map[string][]ResourceOperation{
				AnotherUserId: {"read"},
			}
		})
		g.Expect(e).To(HaveOccurred())

		// non-member user but shared - can write if allowed
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.Sharing = map[string][]ResourceOperation{
				AnotherUserId: {"read", "write"},
			}
		})
		g.Expect(e).To(Succeed())
	}
}
