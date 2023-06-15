package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

const (
	specialPermissionSkipTenancyCheck = "SKIP_TENANCY_CHECK"
)

type loadModelFunc func(ctx context.Context, db *gorm.DB, tenantId uuid.UUID, g *gomega.WithT) *ModelA

var (
	MockedModelLookupByTenant = map[uuid.UUID][]uuid.UUID{}
	MockedModelLookupByOwner  = map[uuid.UUID][]uuid.UUID{}
)

/*************************
	Test
 *************************/

func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		dbtest.EnableDBRecordMode(),
	)
}

type FilterTestDI struct {
	fx.In
	dbtest.DI
	TA tenancy.Accessor
}

func TestOPAFilter(t *testing.T) {
	di := &FilterTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(tenancy.Module, opa.Module),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithFxOptions(
			fx.Provide(testdata.ProvideMockedTenancyAccessor),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestCreateTenancyModels(di)),
		test.GomegaSubTest(SubTestModelListWithAllSupportedFields(di), "TestModelListWithAllSupportedFields"),
		test.GomegaSubTest(SubTestModelListWithoutTenancy(di), "TestModelListWithoutTenancy"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestCreateTenancyModels(di *FilterTestDI) test.SetupFunc {
	var models []*ModelA
	closure := func(ctx context.Context, db *gorm.DB) {
		for _, m := range models {
			// by tenant
			ids, ok := MockedModelLookupByTenant[m.TenantID]
			if !ok || ids == nil {
				ids = make([]uuid.UUID, 0, 5)
			}
			ids = append(ids, m.ID)
			MockedModelLookupByTenant[m.TenantID] = ids
			// by owner
			ids, ok = MockedModelLookupByOwner[m.OwnerID]
			if !ok || ids == nil {
				ids = make([]uuid.UUID, 0, 5)
			}
			ids = append(ids, m.ID)
			MockedModelLookupByOwner[m.OwnerID] = ids
		}
	}
	return dbtest.PrepareData(&di.DI,
		dbtest.SetupUsingSQLFile(testdata.ModelADataFS, "create_table_a.sql"),
		dbtest.SetupTruncateTables(ModelA{}.TableName()),
		dbtest.SetupUsingModelSeedFile(testdata.ModelADataFS, &models, "model_a.yml", closure),
	)
}

func SubTestModelListWithAllSupportedFields(di *FilterTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelA
		var rs *gorm.DB
		// user1
		ctx = sectest.ContextWithSecurity(ctx, user1Options())
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(1), "user1 should see %d models", 1)

		// user1 with parent Tenant A
		ctx = sectest.ContextWithSecurity(ctx, user1Options(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(3), "user1 should see %d models", 3)

		// user2
		ctx = sectest.ContextWithSecurity(ctx, user2Options())
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(1), "user1 should see %d models", 1)
	}
}

func SubTestModelListWithoutTenancy(di *FilterTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelB
		var rs *gorm.DB
		// user1
		ctx = sectest.ContextWithSecurity(ctx, user1Options())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(5), "user1 should see %d models", 5)

		// user2
		ctx = sectest.ContextWithSecurity(ctx, user2Options())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(4), "user1 should see %d models", 4)
	}
}

/*************************
	Helpers
 *************************/

func user1Options(tenantId ...uuid.UUID) sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "user1"
		d.UserId = testdata.MockedUserId1.String()
		d.TenantExternalId = "Tenant A"
		d.Permissions = utils.NewStringSet("NO_VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(testdata.MockedTenantIdA.String())
		d.TenantId = testdata.MockedTenantIdA1.String()
		if len(tenantId) != 0 {
			d.TenantId = tenantId[0].String()
		}
	})
}

func user2Options(tenantId ...uuid.UUID) sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "user2"
		d.UserId = testdata.MockedUserId2.String()
		d.TenantExternalId = "Tenant B"
		d.Permissions = utils.NewStringSet("NO_VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(testdata.MockedTenantIdB.String())
		d.TenantId = testdata.MockedTenantIdB1.String()
		if len(tenantId) != 0 {
			d.TenantId = tenantId[0].String()
		}
	})
}

func loadModelWithId(ctx context.Context, db *gorm.DB, id uuid.UUID, g *gomega.WithT) *ModelA {
	m := ModelA{}
	r := db.WithContext(ctx).Take(&m, id)
	g.Expect(r.Error).To(Succeed(), "load model with ID [%v] should return no error", id)
	return &m
}

func loadModelForTenantId(ctx context.Context, db *gorm.DB, tenantId uuid.UUID, g *gomega.WithT) *ModelA {
	return loadModelWithId(ctx, db, MockedModelLookupByTenant[tenantId][0], g)
}

func toSoftDeleteVariation(m *ModelA) *ModelASoftDelete {
	return &ModelASoftDelete{
		ID:         m.ID,
		TenantName: m.TenantName,
		OwnerName:  m.OwnerName,
		Value:      m.Value,
		TenantID:   m.TenantID,
		TenantPath: m.TenantPath,
		OwnerID:    m.OwnerID,
		Audit:      m.Audit,
	}
}

/*************************
	Models
 *************************/

type ModelA struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value       string
	TenantName  string
	OwnerName   string
	TenantID    uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:tenant_id"`
	TenantPath  pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null" opa:"field:tenant_path"`
	OwnerID     uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:owner_id"`
	PolicyAware `opa:"type:poc"`
	types.Audit
}

func (ModelA) TableName() string {
	return "test_opa_model_a"
}

//func (t *ModelA) BeforeUpdate(tx *gorm.DB) error {
//	if security.HasPermissions(security.Get(tx.Statement.Context), specialPermissionSkipTenancyCheck) {
//		t.SkipTenancyCheck(tx)
//	}
//	return t.Tenancy.BeforeUpdate(tx)
//}

// ModelB has no tenancy
type ModelB struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value      string
	TenantName string
	OwnerName  string
	OwnerID    uuid.UUID    `gorm:"type:KeyID;not null" opa:"field:owner_id"`
	OPAPolicyFilter PolicyFilter `gorm:"-" opa:"type:poc"`
	types.Audit
}

func (ModelB) TableName() string {
	return "test_opa_model_a"
}

type ModelASoftDelete struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value      string
	TenantName string
	OwnerName  string
	TenantID   uuid.UUID     `gorm:"type:KeyID;not null"`
	TenantPath pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null"  json:"-"`
	OwnerID    uuid.UUID     `gorm:"type:KeyID;not null"`
	types.Audit
	types.SoftDelete
}

func (ModelASoftDelete) TableName() string {
	return "test_opa_model_a"
}
