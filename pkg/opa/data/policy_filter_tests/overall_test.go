package policy_filter_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opadata "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data/policy_filter_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"errors"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
	"time"
)

type LookupKey struct {
	Tenant uuid.UUID
	Owner  uuid.UUID
}

var (
	MockedModelLookupByTenant = map[uuid.UUID][]uuid.UUID{}
	MockedModelLookupByOwner  = map[uuid.UUID][]uuid.UUID{}
	MockedModelLookupByKey    = map[LookupKey][]uuid.UUID{}
)

/*************************
	Test
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type TestDI struct {
	fx.In
	dbtest.DI
	TA tenancy.Accessor
}

func TestOPAFilterWithAllFields(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(10 * time.Minute),
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
		test.SubTestSetup(SetupTestCreateModels(&di.DI)),
		test.GomegaSubTest(SubTestModelCreate(di), "TestModelCreate"),
		test.GomegaSubTest(SubTestModelCreateByMap(di), "TestModelCreateByMap"),
		test.GomegaSubTest(SubTestModelList(di), "TestModelList"),
		test.GomegaSubTest(SubTestModelGet(di), "TestModelGet"),
		test.GomegaSubTest(SubTestModelUpdate(di), "TestModelUpdate"),
		test.GomegaSubTest(SubTestModelUpdateWithDelta(di), "TestModelUpdateWithDelta"),
		test.GomegaSubTest(SubTestModelDelete(di), "TestModelDelete"),
		//test.GomegaSubTest(SubTestModelSave(di), "TestModelSave"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestCreateModels(di *dbtest.DI) test.SetupFunc {
	var models []*ModelA
	closure := func(ctx context.Context, db *gorm.DB) {
		for _, m := range models {
			key := LookupKey{Tenant: m.TenantID, Owner: m.OwnerID}
			prepareIdLookup(m.ID, key)
		}
	}
	// We use special DB scope to prepare data, to by-pass policy filtering
	return dbtest.PrepareDataWithScope(di,
		dbtest.SetupWithGormScopes(opadata.SkipPolicyFiltering()),
		dbtest.SetupUsingSQLFile(testdata.ModelADataFS, "create_table_a.sql"),
		dbtest.SetupTruncateTables(ModelA{}.TableName()),
		dbtest.SetupUsingModelSeedFile(testdata.ModelADataFS, &models, "model_a.yml", closure),
	)
}

func SubTestModelCreate(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var model, model2 ModelA
		var rs *gorm.DB
		model = ModelA{
			Value:      "test created",
			TenantName: "Tenant A-1",
			OwnerName:  "user1",
			TenantID:   testdata.MockedTenantIdA1,
			TenantPath: pqx.UUIDArray{testdata.MockedRootTenantId, testdata.MockedTenantIdA, testdata.MockedTenantIdA1},
			OwnerID:    testdata.MockedUserId1,
		}
		model2 = model
		// user1 - tenant A-1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		rs = di.DB.WithContext(ctx).Create(&model)
		g.Expect(rs.Error).To(Succeed(), "create model of currently selected tenant should return no error")

		// user1 with parent Tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		model2.ID = uuid.New()
		rs = di.DB.WithContext(ctx).CreateInBatches([]*ModelA{&model, &model2}, 10)
		g.Expect(rs.Error).To(Succeed(), "create model of current tenant's sub-tenant should return no error")

		// user1 with other tenant branch
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdB), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		rs = di.DB.WithContext(ctx).Create(&model)
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")

		// user1 without permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		model.ID = uuid.New()
		model2.ID = uuid.New()
		rs = di.DB.WithContext(ctx).CreateInBatches([]*ModelA{&model, &model2}, 10)
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		rs = di.DB.WithContext(ctx).Create(&model)
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")
	}
}

func SubTestModelCreateByMap(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		modelMap := map[string]interface{} {
			"ID": uuid.New(),
			"Value": "test created",
			"TenantName": "Tenant A-1",
			"OwnerName": "user1",
			"TenantID": testdata.MockedTenantIdA1,
			"TenantPath": pqx.UUIDArray{testdata.MockedRootTenantId, testdata.MockedTenantIdA, testdata.MockedTenantIdA1},
			"OwnerID": testdata.MockedUserId1,
			"CreatedAt": time.Now(),
			"CreatedBy": testdata.MockedUserId1,
		}
		// user1 - tenant A-1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		g.Expect(rs.Error).To(Succeed(), "create model of currently selected tenant should return no error")

		// user1 with parent Tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		g.Expect(rs.Error).To(Succeed(), "create model of current tenant's sub-tenant should return no error")

		// user1 with other tenant branch
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdB), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")

		// user1 without permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")
	}
}

func SubTestModelList(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelA
		var rs *gorm.DB
		// user1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(1), "user1 should see %d models", 1)

		// user1 with parent Tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(3), "user1 should see %d models", 3)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(1), "user1 should see %d models", 1)
	}
}

func SubTestModelGet(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var id uuid.UUID
		var model ModelA
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = takeModelById(ctx, di.DB, &model, id)
		g.Expect(rs.Error).To(Succeed(), "get model as owner should return no error")

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("VIEW"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = takeModelById(ctx, di.DB, &model, id)
		g.Expect(rs.Error).To(Succeed(), "get model with permission should return no error")

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = takeModelById(ctx, di.DB, &model, id)
		g.Expect(rs.Error).To(HaveOccurred(), "get model of others should return error")

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = takeModelById(ctx, di.DB, &model, id)
		g.Expect(rs.Error).To(HaveOccurred(), "get model of other tenant should return error")
	}
}

func SubTestModelUpdate(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const NewValue = `Updated`
		var id uuid.UUID
		var model *ModelA
		var rs *gorm.DB
		var e error
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{Value: NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model as owner should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "update model as owner should affect rows")
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		g.Expect(model.Value).To(Equal(NewValue), "model's value should be updated")

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(map[string]interface{}{"value": NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model with permission should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "update model with permission should affect rows")
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		g.Expect(model.Value).To(Equal(NewValue), "model's value should be updated")

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{Value: NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model of others should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "update model of others should not affect any rows")

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{Value: NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model of other tenant should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "update model of other tenant should not affect any rows")
	}
}

func SubTestModelUpdateWithDelta(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var NewValue = uuid.MustParse("a5aaa07a-e7d7-4f66-bec8-1e651badacbd")
		var id uuid.UUID
		var model *ModelA
		var rs *gorm.DB
		var e error
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{OwnerID: NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model as owner should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "update model as owner should affect rows")
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		g.Expect(model.OwnerID).To(Equal(NewValue), "model's value should be updated")

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(map[string]interface{}{"owner_id": NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model with permission should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "update model with permission should affect rows")
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		g.Expect(model.OwnerID).To(Equal(NewValue), "model's value should be updated")

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{OwnerID: NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model of others should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "update model of others should not affect any rows")

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{OwnerID: NewValue})
		g.Expect(rs.Error).To(Succeed(), "update model of other tenant should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "update model of other tenant should not affect any rows")
	}
}

func SubTestModelDelete(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var id uuid.UUID
		var rs *gorm.DB
		var e error
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		g.Expect(rs.Error).To(Succeed(), "delete model as owner should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "delete model as owner should affect rows")
		_, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(HaveOccurred(), "get model after delete should return error")
		g.Expect(errors.Is(e, data.ErrorRecordNotFound)).To(BeTrue(), "get model after delete should return record not found error")

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		g.Expect(rs.Error).To(Succeed(), "delete model with permission should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "delete model with permission should affect rows")
		_, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(HaveOccurred(), "get model after delete should return error")
		g.Expect(errors.Is(e, data.ErrorRecordNotFound)).To(BeTrue(), "get model after delete should return record not found error")

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB), testdata.ExtraPermsSecurityOptions("VIEW"))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		g.Expect(rs.Error).To(Succeed(), "delete model of others should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "delete model of others should not affect any rows")
		_, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "get model after delete should return no error")

		// user1 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdB1)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		g.Expect(rs.Error).To(Succeed(), "delete model of other permission should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "delete model of others should not affect any rows")
		_, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "get model after delete should return no error")
	}
}

func SubTestModelSave(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const NewValue = `Saved`
		var id uuid.UUID
		var model *ModelA
		var rs *gorm.DB
		var e error
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		g.Expect(rs.Error).To(Succeed(), "save model as owner should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "save model as owner should affect rows")
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		g.Expect(model.Value).To(Equal(NewValue), "model's value should be updated")

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		g.Expect(rs.Error).To(Succeed(), "save model as owner should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(1), "save model as owner should affect rows")
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		g.Expect(model.Value).To(Equal(NewValue), "model's value should be updated")

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		g.Expect(rs.Error).To(Succeed(), "save model of others should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "save model of others should not affect any rows")

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		model, e = loadModel[ModelA](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "model should exist")
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		g.Expect(rs.Error).To(Succeed(), "update model of other tenant should return no error")
		g.Expect(rs.RowsAffected).To(BeEquivalentTo(0), "update model of other tenant should not affect any rows")
	}
}

/*************************
	Helpers
 *************************/

func appendOrNew[T any](slice []T, values ...T) []T {
	if slice == nil {
		slice = make([]T, 0, 5)
	}
	slice = append(slice, values...)
	return slice
}

func prepareIdLookup(modelId uuid.UUID, key LookupKey) {
	var ids []uuid.UUID
	ids, _ = MockedModelLookupByKey[key]
	MockedModelLookupByKey[key] = appendOrNew(ids, modelId)
	ids, _ = MockedModelLookupByTenant[key.Tenant]
	MockedModelLookupByTenant[key.Tenant] = appendOrNew(ids, modelId)
	ids, _ = MockedModelLookupByOwner[key.Owner]
	MockedModelLookupByOwner[key.Owner] = appendOrNew(ids, modelId)
}



func findID(ownerId, tenantId uuid.UUID) uuid.UUID {
	key := LookupKey{Tenant: tenantId, Owner: ownerId}
	ids, _ := MockedModelLookupByKey[key]
	if len(ids) == 0 {
		return uuid.UUID{}
	}
	return ids[0]
}

func findIDByTenant(tenantId uuid.UUID) uuid.UUID {
	ids, _ := MockedModelLookupByTenant[tenantId]
	if len(ids) == 0 {
		return uuid.UUID{}
	}
	return ids[0]
}

func findIDByOwner(ownerId uuid.UUID) uuid.UUID {
	ids, _ := MockedModelLookupByOwner[ownerId]
	if len(ids) == 0 {
		return uuid.UUID{}
	}
	return ids[0]
}

func takeModelById[T any](ctx context.Context, db *gorm.DB, dest *T, id uuid.UUID) (rs *gorm.DB) {
	var zero T
	*dest = zero
	rs = db.WithContext(ctx).Take(&dest, id)
	return
}

// loadModel load model without policy filtering
func loadModel[T any](ctx context.Context, db *gorm.DB, id uuid.UUID) (*T, error) {
	var dest T
	rs := db.WithContext(ctx).Scopes(opadata.SkipPolicyFiltering()).Take(&dest, id)
	if rs.Error != nil {
		return nil, rs.Error
	}
	return &dest, nil
}

func shallowCopyMap[K comparable, V any](src map[K]V) map[K]V {
	dest := map[K]V{}
	for k, v := range src {
		dest[k] = v
	}
	return dest
}

/*************************
	Models
 *************************/

type ModelA struct {
	ID                  uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value               string
	TenantName          string
	OwnerName           string
	TenantID            uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:tenant_id"`
	TenantPath          pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null" opa:"field:tenant_path"`
	OwnerID             uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:owner_id"`
	opadata.PolicyAware `opa:"type:poc"`
	types.Audit
	types.SoftDelete
}

func (ModelA) TableName() string {
	return "test_opa_model_a"
}
