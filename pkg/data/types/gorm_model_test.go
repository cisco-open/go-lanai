package types

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

func provideMockTenancyAccessor() tenancy.Accessor {
	return &mocks.MockTenancyAccessor{}
}

type tenancyAccessorDI struct {
	fx.In
	MockAccessor tenancy.Accessor
}

func TestGormModel(t *testing.T) {
	di := &tenancyAccessorDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		apptest.WithModules(tenancy.Module),
		apptest.WithFxOptions(
			fx.Provide(provideMockTenancyAccessor),
		),
		test.GomegaSubTest(SubTestExampleUseScope(di), "Tenancy"),
	)
}

func SubTestExampleUseScope(di *tenancyAccessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tenantId, _ := uuid.NewUUID()
		parentId, _ := uuid.NewUUID()

		accessor := di.MockAccessor.(*mocks.MockTenancyAccessor)
		accessor.Reset([]mocks.TenancyRelation{{tenantId, parentId}}, parentId)

		ctx = sectest.WithMockedSecurity(ctx, func(d *sectest.SecurityDetailsMock) {
			d.Username = "any-username"
			d.UserId = "any-user-id"
			d.TenantId = tenantId.String()
			d.Tenants = utils.NewStringSet(tenantId.String())
			d.TenantExternalId = "any-tenant-ext-id"
			d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
		})

		db := &gorm.DB{
			Statement: &gorm.Statement{
				Context: ctx,
			},
		}

		tenancy := Tenancy{
			TenantID: tenantId,
		}
		err := tenancy.BeforeSave(db)

		g.Expect(err).To(gomega.BeNil(), "tenancy hook shouldn't returns error")
		g.Expect(len(tenancy.TenantPath)).To(gomega.Equal(2))
		g.Expect(tenancy.TenantPath[0]).To(gomega.Equal(parentId))
		g.Expect(tenancy.TenantPath[1]).To(gomega.Equal(tenantId))

		invalidTenantId, _ := uuid.NewUUID()
		invalidTenancy := Tenancy{
			TenantID: invalidTenantId,
		}
		err = invalidTenancy.BeforeSave(db)
		g.Expect(err).To(gomega.Not(gomega.BeNil()), "tenancy hook should returns error because of invalid tenant id")
	}
}

