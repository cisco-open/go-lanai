package types

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"database/sql"
	"errors"
	"fmt"
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

var (
	MockedRootTenantId = uuid.MustParse("23967dfe-d90f-4e1b-9406-e2df6685f232")
	MockedTenantIdA    = uuid.MustParse("d8423acc-28cb-4209-95d6-089de7fb27ef")
	MockedTenantIdB    = uuid.MustParse("37b7181a-0892-4706-8f26-60d286b63f14")
	MockedTenantIdA1   = uuid.MustParse("be91531e-ca96-46eb-aea6-b7e0e2a50e21")
	MockedTenantIdA2   = uuid.MustParse("b50c18d9-1741-49bd-8536-30943dfffb45")
	MockedTenantIdB1   = uuid.MustParse("1513b015-6a7d-4de3-8b4f-cbb090ac126d")
	MockedTenantIdB2   = uuid.MustParse("b21445de-9192-45de-acd7-91745ab4cc13")
	MockedModelIDs     = map[uuid.UUID]uuid.UUID{
		MockedRootTenantId: uuid.MustParse("23202b9c-9752-46fa-89ae-9c76277e9bab"),
		MockedTenantIdA:    uuid.MustParse("c60a624a-271a-4a95-96db-9cb7f395f10f"),
		MockedTenantIdB:    uuid.MustParse("435a81e9-1e39-4b66-9211-7cdeea0cda8f"),
		MockedTenantIdA1:   uuid.MustParse("ff9887fb-6809-46f3-b8f1-2de9f4054e36"),
		MockedTenantIdA2:   uuid.MustParse("d1225359-c075-4b0f-ad61-c6e5318f6056"),
		MockedTenantIdB1:   uuid.MustParse("c7547b63-6631-43e0-815d-03bf5e2728a1"),
		MockedTenantIdB2:   uuid.MustParse("2739ee86-d9fb-48f9-9ff9-29f78bfc96c4"),
	}
)

type loadModelFunc func(ctx context.Context, db *gorm.DB, tenantId uuid.UUID, g *gomega.WithT) *TenancyModel

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
	DB *gorm.DB
}

func TestTenancyEnforcement(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(tenancy.Module),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithFxOptions(
			fx.Provide(provideMockedTenancyAccessor),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestCreateTenancyModels(di)),
		test.GomegaSubTest(SubTestSkipTenancyCheck(di), "TestSkipTenancyCheck"),
		test.GomegaSubTest(SubTestTenancySave(di, loadModelForTenantId), "TestSaveLoadedModel"),
		test.GomegaSubTest(SubTestTenancySaveNoAccess(di, loadModelForTenantId), "TestSaveLoadedModelNoAccess"),
		test.GomegaSubTest(SubTestTenancySave(di, synthesizeModelForTenantId), "TestSaveSynthesizedModel"),
		test.GomegaSubTest(SubTestTenancySaveNoAccess(di, synthesizeModelForTenantId), "TestSaveSynthesizedModelNoAccess"),
		test.GomegaSubTest(SubTestTenancyUpdates(di), "TestUpdates"),
		test.GomegaSubTest(SubTestTenancyUpdatesNoAccess(di), "TestUpdatesNoAccess"),
		test.GomegaSubTest(SubTestTenancyUpdatesInvalidTarget(di), "TestUpdatesInvalidTarget"),
		test.GomegaSubTest(SubTestTenancyDelete(di, loadModelForTenantId), "TestDeleteLoadedModel"),
		test.GomegaSubTest(SubTestTenancyDelete(di, synthesizeModelForTenantId), "TestDeleteSynthesizedModel"),
		test.GomegaSubTest(SubTestTenancyDeleteNoAccess(di, loadModelForTenantId), "TestDeleteLoadedModelNoAccess"),
		test.GomegaSubTest(SubTestTenancyDeleteNoAccess(di, synthesizeModelForTenantId), "TestDeleteSynthesizedModelNoAccess"),
		test.GomegaSubTest(SubTestTenancyWithoutSecurity(di), "TestWithoutSecurity"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestCreateTenancyModels(di *testDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		prepareTable(di.DB, g)
		table := TenancyModel{}.TableName()
		r := di.DB.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" RESTRICT`, table))
		g.Expect(r.Error).To(Succeed(), "truncating table of %s should return no error", table)

		reqA := []*TenancyModel{
			newModelWithTenantId(MockedTenantIdA, "Tenant A"),
			newModelWithTenantId(MockedTenantIdA1, "Tenant A-1"),
			newModelWithTenantId(MockedTenantIdA2, "Tenant A-2"),
		}
		reqRoot := []*TenancyModel{
			newModelWithTenantId(MockedTenantIdB, "Tenant B"),
			newModelWithTenantId(MockedTenantIdB1, "Tenant B-1"),
			newModelWithTenantId(MockedTenantIdB2, "Tenant B-2"),
			newModelWithTenantId(MockedRootTenantId, "Root Tenant"),
		}

		// check some invalid cases
		m := newModelWithTenantId(uuid.Nil, "No Tenant")
		r = di.DB.WithContext(ctx).Create(m)
		g.Expect(r.Error).To(Not(Succeed()), "creation of model without tenant ID should return error")

		// mock security with access to Tenant A only
		secCtx := mockedSecurityWithTenantAccess(ctx, MockedTenantIdA)
		for _, m := range reqA {
			r := di.DB.WithContext(secCtx).Create(m)
			g.Expect(r.Error).To(Succeed(), "creation of model belonging to %s should return no error", m.TenantName)
			g.Expect(m.TenantPath).To(Not(BeEmpty()), "creation of model belonging to %s should populate tenant path", m.TenantName)
		}
		for _, m := range reqRoot {
			r := di.DB.WithContext(secCtx).Create(m)
			g.Expect(r.Error).To(Not(Succeed()), "creation of model belonging to %s should return error due to insufficient access", m.TenantName)
		}

		// mock with access to Root tenant and try previously failed creation again
		secCtx = mockedSecurityWithTenantAccess(ctx, MockedRootTenantId)
		for _, m := range reqRoot {
			r := di.DB.WithContext(secCtx).Create(m)
			g.Expect(r.Error).To(Succeed(), "creation of model belonging to %s should return no error", m.TenantName)
			g.Expect(m.TenantPath).To(Not(BeEmpty()), "creation of model belonging to %s should populate tenant path", m.TenantName)
		}
		return ctx, nil
	}
}

func SubTestTenancySave(di *testDI, loadFn loadModelFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		type testCase struct{
			ctx context.Context
			tid uuid.UUID
		}
		cases := []testCase{
			{ctx: mockedSecurityWithTenantAccess(ctx, MockedTenantIdA), tid: MockedTenantIdA},
			{ctx: mockedSecurityWithAllTenantAccess(ctx), tid: MockedTenantIdB},
		}
		for _, c := range cases {
			// Save without changing TenantID
			m := loadFn(ctx, di.DB, c.tid, g)
			cpy := *m
			cpy.Value = "Updated"
			r := di.DB.WithContext(c.ctx).Save(&cpy)
			g.Expect(r.Error).To(Succeed(), "save model belonging to %s should return no error", m.TenantName)
			g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "save model belonging to %s should change 1 row", m.TenantName)
			g.Expect(cpy.TenantPath).To(HaveLen(2), "save model belonging to %s should have correct tenant path", m.TenantName)

			// Save with changed TenantID
			m = loadFn(ctx, di.DB, c.tid, g)
			cpy = *m
			cpy.Value = "Updated"
			cpy.TenantID = MockedTenantIdA1 // move to sub tenant
			r = di.DB.WithContext(c.ctx).Save(&cpy)
			g.Expect(r.Error).To(Succeed(), "save model belonging to %s should return no error", m.TenantName)
			g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "save model belonging to %s should change 1 row", m.TenantName)
			g.Expect(cpy.TenantPath).To(HaveLen(3), "save model belonging to %s should have correct tenant path", m.TenantName)

		}
	}
}

func SubTestTenancySaveNoAccess(di *testDI, loadFn loadModelFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		secCtx := mockedSecurityWithTenantAccess(ctx, MockedTenantIdB)
		// insufficient access Save without changing TenantID
		m := loadFn(ctx, di.DB, MockedTenantIdA, g)
		cpy := *m
		cpy.Value = "Updated"
		r := di.DB.WithContext(secCtx).Save(&cpy)
		g.Expect(r.Error).To(Not(Succeed()), "save model belonging to %s should fail due to insufficient access", m.TenantName)

		// insufficient access Save after TenantID changed (model moved to an inaccessible tenant)
		m = loadFn(ctx, di.DB, MockedTenantIdB, g)
		cpy = *m
		cpy.Value = "Updated"
		cpy.TenantID = MockedTenantIdA1 // move to sub tenant
		r = di.DB.WithContext(secCtx).Save(&cpy)
		g.Expect(r.Error).To(Not(Succeed()), "save model belonging to %s should fail due to insufficient access", m.TenantName)
	}
}

func SubTestTenancyUpdates(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		secCtx := mockedSecurityWithTenantAccess(ctx, MockedTenantIdA1, MockedTenantIdA2)
		// Updates using Map without changing tenant ID
		id := MockedModelIDs[MockedTenantIdA1]
		r := di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{"Value": "Updated"})
		m := loadModelWithId(ctx, di.DB, id, g)
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", m.TenantName)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should change 1 row", m.TenantName)
		g.Expect(m.Value).To(Equal("Updated"), "updated model belonging to %s should have correct Value", m.TenantName)
		g.Expect(m.TenantPath).To(HaveLen(3), "updated model belonging to %s should have correct tenant path", m.TenantName)

		// Updates using Struct without changing tenant ID
		id = MockedModelIDs[MockedTenantIdA2]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(&TenancyModel{Value: "Updated"})
		m = loadModelWithId(ctx, di.DB, id, g)
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", m.TenantName)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should change 1 row", m.TenantName)
		g.Expect(m.Value).To(Equal("Updated"), "updated model belonging to %s should have correct Value", m.TenantName)
		g.Expect(m.TenantPath).To(HaveLen(3), "updated model belonging to %s should have correct tenant path", m.TenantName)

		// Updates using Map with changed TenantID (move to another tenant)
		secCtx = mockedSecurityWithTenantAccess(ctx, MockedTenantIdA, MockedTenantIdB1, MockedTenantIdB2)
		id = MockedModelIDs[MockedTenantIdB1]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{"tenant_id": MockedTenantIdA, "Value": "Updated"})
		m = loadModelWithId(ctx, di.DB, id, g)
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", m.TenantName)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should change 1 row", m.TenantName)
		g.Expect(m.Value).To(Equal("Updated"), "updated model belonging to %s should have correct Value", m.TenantName)
		g.Expect(m.TenantPath).To(HaveLen(2), "updated model belonging to %s should have correct tenant path", m.TenantName)

		// Updates using Struct with changed TenantID (move to another tenant)
		id = MockedModelIDs[MockedTenantIdB2]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(&TenancyModel{Tenancy: Tenancy{TenantID: MockedTenantIdA}, Value: "Updated"})
		m = loadModelWithId(ctx, di.DB, id, g)
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", m.TenantName)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should change 1 row", m.TenantName)
		g.Expect(m.Value).To(Equal("Updated"), "updated model belonging to %s should have correct Value", m.TenantName)
		g.Expect(m.TenantPath).To(HaveLen(2), "updated model belonging to %s should have correct tenant path", m.TenantName)

		// Updates with WHERE clause (only update for tenant A and move tenant for tenant B)
		secCtx = mockedSecurityWithTenantAccess(ctx, MockedTenantIdA, MockedTenantIdB)
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{}).
			Where(&TenancyModel{TenantName: "Tenant A"}).Or(&TenancyModel{TenantName: "Tenant B"}).
			Updates(&TenancyModel{Tenancy: Tenancy{TenantID: MockedTenantIdA1}, Value: "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", m.TenantName)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(2), "update model with WHERE clause should changed 2 rows", m.TenantName)
		for _, id := range []uuid.UUID{MockedModelIDs[MockedTenantIdA], MockedModelIDs[MockedTenantIdB]} {
			m = loadModelWithId(ctx, di.DB, id, g)
			g.Expect(m.Value).To(Equal("Updated"), "updated model belonging to %s should have correct Value", m.TenantName)
			g.Expect(m.TenantPath).To(HaveLen(3), "updated model belonging to %s should have correct tenant path", m.TenantName)
		}

		// Updates with changed TenantID using AccessAllTenants permission
		secCtx = mockedSecurityWithAllTenantAccess(ctx)
		id = MockedModelIDs[MockedTenantIdA]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(&TenancyModel{Tenancy: Tenancy{TenantID: MockedTenantIdB1}, Value: "Updated Again"})
		m = loadModelWithId(ctx, di.DB, id, g)
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", m.TenantName)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should change 1 row", m.TenantName)
		g.Expect(m.Value).To(Equal("Updated Again"), "updated model belonging to %s should have correct Value", m.TenantName)
		g.Expect(m.TenantPath).To(HaveLen(3), "updated model belonging to %s should have correct tenant path", m.TenantName)

		// Updates without security
		secCtx = ctx
		id = MockedModelIDs[MockedTenantIdB]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(&TenancyModel{Value: "Updated Again"})
		m = loadModelWithId(ctx, di.DB, id, g)
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", m.TenantName)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should change 1 row", m.TenantName)
		g.Expect(m.Value).To(Equal("Updated Again"), "updated model belonging to %s should have correct Value", m.TenantName)
		// note this was changed from previous test cases
		g.Expect(m.TenantPath).To(HaveLen(3), "updated model belonging to %s should have correct tenant path", m.TenantName)
	}
}

func SubTestTenancyUpdatesNoAccess(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		secCtx := mockedSecurityWithTenantAccess(ctx, MockedTenantIdB)
		// Updates using Map without changing tenant ID
		id := MockedModelIDs[MockedTenantIdA1]
		r := di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{"Value": "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", MockedTenantIdA1)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(0), "update model belonging to %s should update 0 rows due to insufficient access", MockedTenantIdA1)

		// Updates using Struct without changing tenant ID
		secCtx = mockedSecurityWithTenantAccess(ctx, MockedTenantIdB, MockedTenantIdB1, MockedTenantIdB2)
		id = MockedModelIDs[MockedTenantIdA2]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(&TenancyModel{Value: "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", MockedTenantIdA2)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(0), "update model belonging to %s should update 0 rows due to insufficient access", MockedTenantIdA2)

		// Move to tenant using Map, target tenant is not accessible
		secCtx = mockedSecurityWithTenantAccess(ctx, MockedTenantIdB1, MockedTenantIdB2)
		id = MockedModelIDs[MockedTenantIdB1]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{"tenant_id": MockedTenantIdA, "Value": "Updated"})
		g.Expect(r.Error).To(Not(Succeed()), "update model belonging to %s should return error due to target tenant is inaccessible", MockedTenantIdB1)

		// Move to tenant using Struct, target tenant is not accessible
		id = MockedModelIDs[MockedTenantIdB2]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(&TenancyModel{Tenancy: Tenancy{TenantID: MockedTenantIdA}, Value: "Updated"})
		g.Expect(r.Error).To(Not(Succeed()), "update model belonging to %s should return error due to target tenant is inaccessible", MockedTenantIdB1)

		// Move to tenant using Map, source tenant is not accessible
		secCtx = mockedSecurityWithTenantAccess(ctx, MockedTenantIdB1, MockedTenantIdB2)
		id = MockedModelIDs[MockedTenantIdA1]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{"TenantID": MockedTenantIdB1, "Value": "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", MockedTenantIdA1)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(0), "update model belonging to %s should update 0 rows due to insufficient access", MockedTenantIdB1)

		// Updates using Map with WHERE clause (matched rows should have 1 updated and 1 unchanged)
		secCtx = mockedSecurityWithTenantAccess(ctx, MockedTenantIdA)
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{}).
			Where(&TenancyModel{TenantName: "Tenant A"}).Or(&TenancyModel{TenantName: "Tenant B"}).
			Updates(map[string]interface{}{"Value": "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model with WHERE clause should return no error")
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model with WHERE clause should update 0 rows due to insufficient access")
	}
}

func SubTestTenancyUpdatesInvalidTarget(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		secCtx := mockedSecurityWithAllTenantAccess(ctx)
		// Updates using Struct
		id := MockedModelIDs[MockedTenantIdA1]
		target1 := TenancyModel{Tenancy: Tenancy{TenantID: MockedTenantIdA}, Value: "Updated"}
		r := di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(target1)
		g.Expect(r.Error).To(Not(Succeed()), "update model should return error due to invalid update target %T", target1)

		// Updates using Map with non-string key
		id = MockedModelIDs[MockedTenantIdA1]
		target2 := map[sql.NullString]interface{}{sql.NullString{String: "tenant_id", Valid: true}: MockedTenantIdA}
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(target2)
		g.Expect(r.Error).To(Not(Succeed()), "update model should return error due to invalid update target %T", target2)

		// Updates using other types
		id = MockedModelIDs[MockedTenantIdA1]
		target3 := "tenant_id = " + id.String()
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(target3)
		g.Expect(r.Error).To(Not(Succeed()), "update model should return error due to invalid update target %T", target3)

		// Updates using Map with changed TenantID (move to another tenant) and incorrect tenant path
		id = MockedModelIDs[MockedTenantIdB1]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{
				"tenant_id": MockedTenantIdA,
				"TenantPath": []uuid.UUID{uuid.New(), uuid.New()},
			})
		g.Expect(r.Error).To(Not(Succeed()), "update model should return error due to wrong tenant path is explicitly set")
	}
}

func SubTestTenancyDelete(di *testDI, loadFn loadModelFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		type testCase struct{
			ctx context.Context
			tid []uuid.UUID
		}
		cases := []testCase{
			{ctx: mockedSecurityWithTenantAccess(ctx, MockedTenantIdA), tid: []uuid.UUID{MockedTenantIdA1, MockedTenantIdA2}},
			{ctx: mockedSecurityWithAllTenantAccess(ctx), tid: []uuid.UUID{MockedTenantIdB1, MockedTenantIdB2}},
		}
		for _, c := range cases {
			// Hard Delete with access
			tid := c.tid[0]
			m := loadFn(ctx, di.DB, tid, g)
			r := di.DB.WithContext(c.ctx).Model(&TenancyModel{}).Delete(m)
			g.Expect(r.Error).To(Succeed(), "delete model belonging to %s should return no error", m.TenantID)
			g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "delete model belonging to %s should affect 1 rows", m.TenantID)
			reFetch := di.DB.WithContext(ctx).Take(&TenancyModel{}, MockedModelIDs[tid])
			g.Expect(errors.Is(reFetch.Error, gorm.ErrRecordNotFound)).To(BeTrue(), "fetch deleted model belonging to %s should return not found error", m.TenantID)

			// Soft Delete with access (use special variation TenancySoftDeleteModel)
			tid = c.tid[1]
			m = loadFn(ctx, di.DB, tid, g)
			r = di.DB.WithContext(c.ctx).Model(&TenancySoftDeleteModel{}).Delete(toSoftDeleteVariation(m))
			g.Expect(r.Error).To(Succeed(), "delete model belonging to %s should return no error", m.TenantID)
			g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "delete model belonging to %s should affect 1 rows", m.TenantID)
			reFetch = di.DB.WithContext(ctx).Take(&TenancySoftDeleteModel{}, MockedModelIDs[tid])
			g.Expect(errors.Is(reFetch.Error, gorm.ErrRecordNotFound)).To(BeTrue(), "fetch deleted model belonging to %s should return not found error", m.TenantID)
		}
	}
}

func SubTestTenancyDeleteNoAccess(di *testDI, loadFn loadModelFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		secCtx := mockedSecurityWithTenantAccess(ctx, MockedTenantIdA)
		// Hard Delete without access
		tid := MockedTenantIdB1
		m := loadFn(ctx, di.DB, tid, g)
		r := di.DB.WithContext(secCtx).Model(&TenancyModel{}).Delete(m)
		g.Expect(r.Error).To(Succeed(), "delete model belonging to %s should return no error", m.TenantID)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(0), "delete model belonging to %s should affect 0 rows due to insufficient access", m.TenantID)
		reFetch := di.DB.WithContext(ctx).Take(&TenancyModel{}, MockedModelIDs[tid])
		g.Expect(reFetch.Error).To(Succeed(), "fetching not-deleted model belonging to %s should return no error", m.TenantID)

		// Soft Delete without access
		tid = MockedTenantIdB2
		m = loadFn(ctx, di.DB, tid, g)
		r = di.DB.WithContext(secCtx).Model(&TenancySoftDeleteModel{}).Delete(toSoftDeleteVariation(m))
		g.Expect(r.Error).To(Succeed(), "delete model belonging to %s should return no error", m.TenantID)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(0), "delete model belonging to %s should affect 0 rows due to insufficient access", m.TenantID)
		reFetch = di.DB.WithContext(ctx).Take(&TenancySoftDeleteModel{}, MockedModelIDs[tid])
		g.Expect(reFetch.Error).To(Succeed(), "fetching not-deleted model belonging to %s should return no error", m.TenantID)
	}
}

func SubTestSkipTenancyCheck(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		secCtx := mockedSecurityWithSkipTenancyCheck(ctx, MockedTenantIdB)
		// Updates without changing tenant ID
		id := MockedModelIDs[MockedTenantIdA1]
		r := di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{"Value": "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", MockedTenantIdA1)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should update 1 rows", MockedTenantIdA1)

		// Move to tenant, target tenant is not accessible
		id = MockedModelIDs[MockedTenantIdB2]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(&TenancyModel{Tenancy: Tenancy{TenantID: MockedTenantIdA}, Value: "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", MockedTenantIdB2)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should update 1 rows", MockedTenantIdB2)

		// Move to tenant, source tenant is not accessible
		secCtx = mockedSecurityWithSkipTenancyCheck(ctx, MockedTenantIdB1, MockedTenantIdB2)
		id = MockedModelIDs[MockedTenantIdA2]
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{ID: id}).
			Updates(map[string]interface{}{"TenantID": MockedTenantIdB1, "Value": "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model belonging to %s should return no error", MockedTenantIdA2)
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "update model belonging to %s should update 1 rows", MockedTenantIdA2)

		// Updates using Map with WHERE clause (matched rows should have 2 updated)
		secCtx = mockedSecurityWithSkipTenancyCheck(ctx, MockedTenantIdB)
		r = di.DB.WithContext(secCtx).Model(&TenancyModel{}).
			Where(&TenancyModel{TenantName: "Tenant A"}).Or(&TenancyModel{TenantName: "Tenant B"}).
			Updates(map[string]interface{}{"Value": "Updated"})
		g.Expect(r.Error).To(Succeed(), "update model with WHERE clause should return no error")
		g.Expect(r.RowsAffected).To(BeEquivalentTo(2), "update model with WHERE clause should update 2 rows")
	}
}

func SubTestTenancyWithoutSecurity(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// create
		var r *gorm.DB
		m := newModelWithTenantId(MockedTenantIdA, "Test Tenant")
		r = di.DB.WithContext(ctx).Save(m)
		g.Expect(r.Error).To(Succeed(), "create model without security context should succeed")
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "create model without security context affect 1 rows")
		g.Expect(m.TenantPath).To(HaveLen(2), "create model without security context should populate tenant path")

		// save
		m.TenantID = MockedTenantIdA1
		r = di.DB.WithContext(ctx).Save(m)
		g.Expect(r.Error).To(Succeed(), "save model without security context should succeed")
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "save model without security context affect 1 rows")
		g.Expect(m.TenantPath).To(HaveLen(3), "save model without security context should populate tenant path")

		// update
		r = di.DB.WithContext(ctx).Model(m).
			Updates(&TenancyModel{Tenancy: Tenancy{TenantID: MockedTenantIdB}, Value: "Updated"})
		g.Expect(r.Error).To(Succeed(), "updates model without security context should succeed")
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "updates model without security context affect 1 rows")
		m = loadModelWithId(ctx, di.DB, m.ID, g)
		g.Expect(m.TenantPath).To(HaveLen(2), "updates model without security context should populate tenant path")
		g.Expect(m.Value).To(Equal("Updated"), "updates model without security context should update Value")

		// delete
		r = di.DB.WithContext(ctx).Model(&TenancyModel{}).Delete(m)
		g.Expect(r.Error).To(Succeed(), "delete model without security context should succeed")
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "delete model without security context affect 1 rows")
		r = di.DB.WithContext(ctx).Model(&TenancyModel{}).Take(&TenancyModel{}, m.ID)
		g.Expect(errors.Is(r.Error, gorm.ErrRecordNotFound)).To(BeTrue(), "delete model without security context should actually delete the record")
	}
}

/*************************
	Helpers
 *************************/

func populateDefaults(d *sectest.SecurityDetailsMock) {
	d.Username = "any-username"
	d.UserId = "any-user-id"
	d.TenantExternalId = "any-tenant-ext-id"
	d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
}

func mockedSecurityWithTenantAccess(parent context.Context, tenantId ...uuid.UUID) context.Context {
	return sectest.WithMockedSecurity(parent, func(m *sectest.SecurityDetailsMock) {
		populateDefaults(m)
		tidStrs := make([]string, len(tenantId))
		for i, id := range tenantId {
			tidStrs[i] = id.String()
		}
		m.Tenants = utils.NewStringSet(tidStrs...)
		m.TenantId = tidStrs[0]
	})
}

func mockedSecurityWithAllTenantAccess(parent context.Context) context.Context {
	return sectest.WithMockedSecurity(parent, func(m *sectest.SecurityDetailsMock) {
		populateDefaults(m)
		m.Tenants = utils.NewStringSet(MockedRootTenantId.String())
		m.TenantId = MockedRootTenantId.String()
		m.Permissions.Add(security.SpecialPermissionAccessAllTenant)
	})
}

func mockedSecurityWithSkipTenancyCheck(parent context.Context, tenantId ...uuid.UUID) context.Context {
	return sectest.WithMockedSecurity(parent, func(m *sectest.SecurityDetailsMock) {
		populateDefaults(m)
		tidStrs := make([]string, len(tenantId))
		for i, id := range tenantId {
			tidStrs[i] = id.String()
		}
		m.Tenants = utils.NewStringSet(tidStrs...)
		m.TenantId = tidStrs[0]
		m.Permissions.Add(specialPermissionSkipTenancyCheck)
	})
}

func newModelWithTenantId(tenantId uuid.UUID, value string) *TenancyModel {
	id, ok := MockedModelIDs[tenantId]
	if !ok {
		id = uuid.New()
	}
	return &TenancyModel{
		ID:         id,
		TenantName: value,
		Value:      value,
		Tenancy: Tenancy{
			TenantID: tenantId,
		},
	}
}

func loadModelWithId(ctx context.Context, db *gorm.DB, id uuid.UUID, g *gomega.WithT) *TenancyModel {
	m := TenancyModel{}
	r := db.WithContext(ctx).Take(&m, id)
	g.Expect(r.Error).To(Succeed(), "load model with ID [%v] should return no error", id)
	return &m
}

func loadModelForTenantId(ctx context.Context, db *gorm.DB, tenantId uuid.UUID, g *gomega.WithT) *TenancyModel {
	return loadModelWithId(ctx, db, MockedModelIDs[tenantId], g)
}

func synthesizeModelForTenantId(_ context.Context, _ *gorm.DB, tenantId uuid.UUID, _ *gomega.WithT) *TenancyModel {
	return newModelWithTenantId(tenantId, "")
}

func toSoftDeleteVariation(m *TenancyModel) *TenancySoftDeleteModel {
	return &TenancySoftDeleteModel{
		ID:         m.ID,
		TenantName: m.TenantName,
		Value:      m.Value,
		Tenancy:    m.Tenancy,
		Audit:      m.Audit,
	}
}

/*************************
	Mocks
 *************************/

func provideMockedTenancyAccessor() tenancy.Accessor {
	tenancyRelationship := []mocks.TenancyRelation{
		{Parent: MockedRootTenantId, Child: MockedTenantIdA},
		{Parent: MockedRootTenantId, Child: MockedTenantIdB},
		{Parent: MockedTenantIdA, Child: MockedTenantIdA1},
		{Parent: MockedTenantIdA, Child: MockedTenantIdA2},
		{Parent: MockedTenantIdB, Child: MockedTenantIdB1},
		{Parent: MockedTenantIdB, Child: MockedTenantIdB2},
	}
	return mocks.NewMockTenancyAccessor(tenancyRelationship, MockedRootTenantId)
}

const tableSQL = `
CREATE TABLE IF NOT EXISTS public.test_tenancy (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"tenant_name" STRING NOT NULL,
	"value" STRING NOT NULL,
	tenant_id UUID NULL,
	tenant_path UUID[] NULL,
	created_at TIMESTAMPTZ NULL,
	updated_at TIMESTAMPTZ NULL,
	created_by UUID NULL,
	updated_by UUID NULL,
	deleted_at TIMESTAMPTZ NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	INVERTED INDEX idx_tenant_path (tenant_path),
	INDEX idx_tenant_name (tenant_name ASC),
	FAMILY "primary" (id, tenant_name, value, tenant_id, tenant_path, created_at, updated_at, created_by, updated_by, deleted_at)
);`

func prepareTable(db *gorm.DB, g *gomega.WithT) {
	r := db.Exec(tableSQL)
	g.Expect(r.Error).To(Succeed(), "create table if not exists shouldn't fail")
}

type TenancyModel struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	TenantName string
	Value      string
	Tenancy
	Audit
}

func (TenancyModel) TableName() string {
	return "test_tenancy"
}

func (t *TenancyModel) BeforeUpdate(tx *gorm.DB) error {
	if security.HasPermissions(security.Get(tx.Statement.Context), specialPermissionSkipTenancyCheck) {
		t.SkipTenancyCheck(tx)
	}
	return t.Tenancy.BeforeUpdate(tx)
}

type TenancySoftDeleteModel struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	TenantName string
	Value      string
	Tenancy
	Audit
	SoftDelete
}

func (TenancySoftDeleteModel) TableName() string {
	return "test_tenancy"
}
