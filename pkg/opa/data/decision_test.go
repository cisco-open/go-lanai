package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
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
	UserId = "20523d89-d5e9-40d0-afe3-3a74c298b55a"
	RootTenantId = "7b3934fc-edc4-4a1c-9249-3dc7055eb124"
	TenantId = "8eebb711-7d24-48fb-94da-361c573d7c20"
	ProviderId = "fe3ad89c-449f-42f2-b4f8-b10ab7bc0266"
)

/*************************
	Test Setup
 *************************/

func adminUserOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "testuser"
		d.UserId = UserId
		d.TenantId = TenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("IS_API_ADMIN")
		d.Roles = utils.NewStringSet("SUPERUSER")
	})
}

/*************************
	Test
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type testDI struct {
	fx.In
}

func TestAllowResource(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(opa.Module),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestTenancy(di), "TestTenancy"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestTenancy(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.ContextWithSecurity(ctx, adminUserOptions())
		e := AllowResource(ctx, "poc", opa.OpRead, func(res *Resource) {
			res.TenantID = TenantId
			res.OwnerID = UserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())
	}
}
