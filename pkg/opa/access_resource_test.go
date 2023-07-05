package opa

import (
	"context"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const (
	OwnerUserId     = "20523d89-d5e9-40d0-afe3-3a74c298b55a"
	AnotherUserId   = "e7498b90-cec3-41fd-ac20-acd41769fb88"
	RootTenantId    = "7b3934fc-edc4-4a1c-9249-3dc7055eb124"
	TenantId        = "8eebb711-7d24-48fb-94da-361c573d7c20"
	AnotherTenantId = "b11ef279-1309-4c43-8355-99c9d494097b"
	ProviderId      = "fe3ad89c-449f-42f2-b4f8-b10ab7bc0266"
)

/*************************
	Test Setup
 *************************/

func memberAdminOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "admin"
		d.UserId = AnotherUserId
		d.TenantId = TenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("IS_API_ADMIN", "VIEW", "MANAGE")
		d.Roles = utils.NewStringSet("SUPERUSER")
		d.Tenants = utils.NewStringSet(TenantId, AnotherTenantId)
	})
}

func memberOwnerOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "testuser-member-owner"
		d.UserId = OwnerUserId
		d.TenantId = TenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(TenantId)
	})
}

func memberNonOwnerOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "testuser-member-non-owner"
		d.UserId = AnotherUserId
		d.TenantId = TenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(TenantId)
	})
}

func nonMemberAdminOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "testuser-non-member"
		d.UserId = AnotherUserId
		d.TenantId = AnotherTenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("IS_API_ADMIN")
		d.Roles = utils.NewStringSet("SUPERUSER")
		d.Tenants = utils.NewStringSet(AnotherTenantId)
	})
}

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

func SubTestMemberAdmin(di *testDI) test.GomegaSubTestFunc {
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

func SubTestMemberOwner(di *testDI) test.GomegaSubTestFunc {
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

func SubTestMemberNonOwner(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// member user
		ctx = sectest.ContextWithSecurity(ctx, memberNonOwnerOptions())
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

func SubTestNonMember(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// non-member admin - can't read
		ctx = sectest.ContextWithSecurity(ctx, nonMemberAdminOptions())
		e = AllowResource(ctx, "poc", OpRead, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(HaveOccurred())
	}
}

func SubTestSharedUser(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, memberNonOwnerOptions())
		// non-member user but shared - cannot write if not allowed
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.Share = map[string][]ResourceOperation{
				AnotherUserId: {"read"},
			}
		})
		g.Expect(e).To(HaveOccurred())

		// non-member user but shared - can write if allowed
		e = AllowResource(ctx, "poc", OpWrite, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.Share = map[string][]ResourceOperation{
				AnotherUserId: {"read", "write"},
			}
		})
		g.Expect(e).To(Succeed())
	}
}
