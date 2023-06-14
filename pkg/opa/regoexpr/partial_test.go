package regoexpr

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const (
	OwnerUserId  = "20523d89-d5e9-40d0-afe3-3a74c298b55a"
	AnotherUserId  = "e7498b90-cec3-41fd-ac20-acd41769fb88"
	RootTenantId = "7b3934fc-edc4-4a1c-9249-3dc7055eb124"
	TenantId = "8eebb711-7d24-48fb-94da-361c573d7c20"
	AnotherTenantId = "b11ef279-1309-4c43-8355-99c9d494097b"
	ProviderId = "fe3ad89c-449f-42f2-b4f8-b10ab7bc0266"
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

// TODO
func TestFilterResource(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		apptest.WithModules(opa.Module),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestFilterByTenantID(di), "TestFilterByTenantID"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestFilterByTenantID(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		//var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, memberAdminOptions())
		// member admin - can read
		//result, e := FilterResource(ctx, "poc", opa.OpRead, func(rf *ResourceFilter) {
		//	rf.QueryMapper = NewGormPartialQueryMapper(&GormMapperConfig{})
		//})
		//g.Expect(e).To(Succeed())
		//fmt.Println(result)
	}
}


