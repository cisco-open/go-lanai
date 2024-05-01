// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package filter_tests

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/pkg/data/types"
	"github.com/cisco-open/go-lanai/pkg/data/types/pqx"
	"github.com/cisco-open/go-lanai/pkg/opa"
	opadata "github.com/cisco-open/go-lanai/pkg/opa/data"
	"github.com/cisco-open/go-lanai/pkg/opa/data/filter_tests/testdata"
	opatest "github.com/cisco-open/go-lanai/pkg/opa/test"
	"github.com/cisco-open/go-lanai/pkg/tenancy"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
	"time"
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

func TestOPAFilterWithAllFieldsPostgresql(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(10*time.Minute),
		dbtest.WithDBPlayback("testdb", dbtest.DBPort(5432), dbtest.DBCredentials("postgres", "")),
		opatest.WithBundles(opatest.DefaultBundleFS, testdata.ModelABundleFS),
		apptest.WithModules(tenancy.Module),
		apptest.WithConfigFS(testdata.ConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.ProvideMockedTenancyAccessor),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareModelA(&di.DI)),
		test.GomegaSubTest(SubTestModelCreate(di), "TestModelCreate"),
		test.GomegaSubTest(SubTestModelCreateByMap(di), "TestModelCreateByMap"),
		test.GomegaSubTest(SubTestModelList(di), "TestModelList"),
		test.GomegaSubTest(SubTestModelGet(di), "TestModelGet"),
		test.GomegaSubTest(SubTestModelUpdate(di), "TestModelUpdate"),
		test.GomegaSubTest(SubTestModelUpdateWithDelta(di), "TestModelUpdateWithDelta"),
		test.GomegaSubTest(SubTestModelDelete(di), "TestModelDelete"),
		test.GomegaSubTest(SubTestModelSave(di), "TestModelSave"),
	)
}

func TestOPAFilterWithAllFields(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(10*time.Minute),
		dbtest.WithDBPlayback("testdb"),
		opatest.WithBundles(opatest.DefaultBundleFS, testdata.ModelABundleFS),
		apptest.WithModules(tenancy.Module),
		apptest.WithConfigFS(testdata.ConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.ProvideMockedTenancyAccessor),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareModelA(&di.DI)),
		test.GomegaSubTest(SubTestModelCreate(di), "TestModelCreate"),
		test.GomegaSubTest(SubTestModelCreateByMap(di), "TestModelCreateByMap"),
		test.GomegaSubTest(SubTestModelList(di), "TestModelList"),
		test.GomegaSubTest(SubTestModelGet(di), "TestModelGet"),
		test.GomegaSubTest(SubTestModelUpdate(di), "TestModelUpdate"),
		test.GomegaSubTest(SubTestModelUpdateWithDelta(di), "TestModelUpdateWithDelta"),
		test.GomegaSubTest(SubTestModelDelete(di), "TestModelDelete"),
		test.GomegaSubTest(SubTestModelSave(di), "TestModelSave"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestPrepareModelA(di *dbtest.DI) test.SetupFunc {
	var models []*ModelA
	closure := func(ctx context.Context, db *gorm.DB) {
		resetIdLookup()
		const more = 9
		extra := make([]*ModelA, 0, len(models)*more)
		for _, m := range models {
			key := LookupKey{Tenant: m.TenantID, Owner: m.OwnerID}
			prepareIdLookup(m.ID, key)
			for i := 0; i < more; i++ {
				newM := *m
				newM.ID = uuid.New()
				prepareIdLookup(newM.ID, key)
				extra = append(extra, &newM)
			}
		}
		db.WithContext(ctx).CreateInBatches(extra, 50)
	}
	// We use special DB scope to prepare data, to by-pass policy filtering
	return dbtest.PrepareDataWithScope(di,
		dbtest.SetupWithGormScopes(opadata.SkipFiltering()),
		dbtest.SetupUsingSQLFile(testdata.ModelDataFS, "create_table_a.sql"),
		dbtest.SetupTruncateTables(ModelA{}.TableName()),
		dbtest.SetupUsingModelSeedFile(testdata.ModelDataFS, &models, "model_a.yml", closure),
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
		assertDBResult(ctx, g, rs, "create model of currently selected tenant", nil, 1)

		// user1 with parent Tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		model2.ID = uuid.New()
		rs = di.DB.WithContext(ctx).CreateInBatches([]*ModelA{&model, &model2}, 10)
		assertDBResult(ctx, g, rs, "batch create models of current tenant's sub-tenant", nil, 2)

		// user1 with other tenant branch
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdB), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		rs = di.DB.WithContext(ctx).Create(&model)
		assertDBResult(ctx, g, rs, "create model without correctly selected tenant", opa.ErrAccessDenied, 0)

		// user1 without permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		model.ID = uuid.New()
		model2.ID = uuid.New()
		rs = di.DB.WithContext(ctx).CreateInBatches([]*ModelA{&model, &model2}, 10)
		assertDBResult(ctx, g, rs, "batch create model without permission", opa.ErrAccessDenied, 0)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		rs = di.DB.WithContext(ctx).Create(&model)
		assertDBResult(ctx, g, rs, "create model without correct owner", opa.ErrAccessDenied, 0)
	}
}

func SubTestModelCreateByMap(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		modelMap := map[string]interface{}{
			"ID":         uuid.New(),
			"Value":      "test created",
			"TenantName": "Tenant A-1",
			"OwnerName":  "user1",
			"TenantID":   testdata.MockedTenantIdA1,
			"TenantPath": pqx.UUIDArray{testdata.MockedRootTenantId, testdata.MockedTenantIdA, testdata.MockedTenantIdA1},
			"OwnerID":    testdata.MockedUserId1,
			"CreatedAt":  time.Now(),
			"CreatedBy":  testdata.MockedUserId1,
		}
		// user1 - tenant A-1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		assertDBResult(ctx, g, rs, "create model of currently selected tenant", nil, 1)

		// user1 with parent Tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		assertDBResult(ctx, g, rs, "create model of current tenant's sub-tenant", nil, 1)

		// user1 with other tenant branch
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdB), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		assertDBResult(ctx, g, rs, "create model without correctly selected tenant", opa.ErrAccessDenied, 0)
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")

		// user1 without permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		assertDBResult(ctx, g, rs, "create model without permission", opa.ErrAccessDenied, 0)
		g.Expect(rs.Error).To(HaveOccurred(), "create model by non tenant member should return error")

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Create(shallowCopyMap(modelMap))
		assertDBResult(ctx, g, rs, "create model without correct owner", opa.ErrAccessDenied, 0)
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
		assertDBResult(ctx, g, rs, "list models of user1", nil, 10)
		g.Expect(models).To(HaveLen(10), "user1 should see %d models", 10)
		assertOwnership(g, testdata.MockedUserId1, "list models of user1", models...)

		// user1 with parent Tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		assertDBResult(ctx, g, rs, "list models with parent tenant admin", nil, 30)
		g.Expect(models).To(HaveLen(30), "user1 should see %d models", 30)
		assertOwnership(g, testdata.MockedUserId1, "list models with parent tenant admin", models...)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelA{}).Find(&models)
		assertDBResult(ctx, g, rs, "list models of user2", nil, 10)
		g.Expect(models).To(HaveLen(10), "user2 should see %d models", 10)
		assertOwnership(g, testdata.MockedUserId2, "list models of user2", models...)
	}
}

func SubTestModelGet(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var id uuid.UUID
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Take(new(ModelA), id)
		assertDBResult(ctx, g, rs, "get model as owner", nil, 1)

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("VIEW"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Take(new(ModelA), id)
		assertDBResult(ctx, g, rs, "get model with permission", nil, 1)

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = di.DB.WithContext(ctx).Take(new(ModelA), id)
		assertDBResult(ctx, g, rs, "get model of others", data.ErrorRecordNotFound, 0)

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Take(new(ModelA), id)
		assertDBResult(ctx, g, rs, "get model of other tenant", data.ErrorRecordNotFound, 0)
	}
}

func SubTestModelUpdate(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const NewValue = `Updated`
		var id uuid.UUID
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{Value: NewValue})
		assertDBResult(ctx, g, rs, "update as owner", nil, 1)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "update as owner", "Value", NewValue)

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(map[string]interface{}{"value": NewValue})
		assertDBResult(ctx, g, rs, "update with permission", nil, 1)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "update with permission", "Value", NewValue)

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{Value: NewValue})
		assertDBResult(ctx, g, rs, "update model of others", nil, 0)

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{Value: NewValue})
		assertDBResult(ctx, g, rs, "update model of other tenant", nil, 0)
	}
}

func SubTestModelUpdateWithDelta(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var NewValue = uuid.MustParse("a5aaa07a-e7d7-4f66-bec8-1e651badacbd")
		var id uuid.UUID
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{OwnerID: NewValue})
		assertDBResult(ctx, g, rs, "change model's owner as owner", nil, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "change model's owner as owner", "OwnerID", testdata.MockedUserId1)

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(map[string]interface{}{"owner_id": NewValue})
		assertDBResult(ctx, g, rs, "change model's owner with permission", nil, 1)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "change model's owner with permission", "OwnerID", NewValue)

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{OwnerID: NewValue})
		assertDBResult(ctx, g, rs, "change model's owner of others", nil, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "change model's owner of others", "OwnerID", testdata.MockedUserId1)

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelA{ID: id}).Updates(&ModelA{OwnerID: NewValue})
		assertDBResult(ctx, g, rs, "update model's owner of other tenant", nil, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "update model's owner of other tenant", "OwnerID", testdata.MockedUserId1)
	}
}

func SubTestModelDelete(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var id uuid.UUID
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		assertDBResult(ctx, g, rs, "delete model as owner", nil, 1)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "delete model as owner")

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		assertDBResult(ctx, g, rs, "delete model with permission", nil, 1)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "delete model with permission")

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB), testdata.ExtraPermsSecurityOptions("VIEW"))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		assertDBResult(ctx, g, rs, "delete model of others", nil, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "delete model of others", "exists")

		// user1 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdB1)
		rs = di.DB.WithContext(ctx).Delete(&ModelA{ID: id})
		assertDBResult(ctx, g, rs, "delete model of other tenant", nil, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "delete model of other tenant", "exists")
	}
}

func SubTestModelSave(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const NewValue = `Saved`
		var id uuid.UUID
		var model *ModelA
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdA2)
		model = mustLoadModel[ModelA](ctx, g, di.DB, id)
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		assertDBResult(ctx, g, rs, "save model as owner", nil, 1)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "save model of other tenant", "Value", NewValue)

		// user1 - not owner, but have permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findID(testdata.MockedUserId2, testdata.MockedTenantIdA2)
		model = mustLoadModel[ModelA](ctx, g, di.DB, id)
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		assertDBResult(ctx, g, rs, "save model with permission", nil, 1)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "save model with permission", "Value", NewValue)

		// user2 - not owner, is member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		model = mustLoadModel[ModelA](ctx, g, di.DB, id)
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		assertDBResult(ctx, g, rs, "save model of others", opa.ErrAccessDenied, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "save model of others", "OwnerID", testdata.MockedUserId1)

		// user2 - not owner, not member, no permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		model = mustLoadModel[ModelA](ctx, g, di.DB, id)
		model.Value = NewValue
		rs = di.DB.WithContext(ctx).Save(model)
		assertDBResult(ctx, g, rs, "save model of other tenant", opa.ErrAccessDenied, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "save model of other tenant", "OwnerID", testdata.MockedUserId1)

		// user2 - not owner, not member, no permission, attempt to change owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdB))
		id = findID(testdata.MockedUserId1, testdata.MockedTenantIdB2)
		model = mustLoadModel[ModelA](ctx, g, di.DB, id)
		model.OwnerID = testdata.MockedUserId2
		rs = di.DB.WithContext(ctx).Save(model)
		assertDBResult(ctx, g, rs, "save model with different owner", opa.ErrAccessDenied, 0)
		assertPostOpModel[ModelA](ctx, g, di.DB, id, "save model with different owner", "OwnerID", testdata.MockedUserId1)
	}
}

/*************************
	Helpers
 *************************/

/*************************
	Models
 *************************/

type ModelA struct {
	ID                    uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value                 string
	TenantName            string
	OwnerName             string
	TenantID              uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:tenant_id"`
	TenantPath            pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null" opa:"field:tenant_path"`
	OwnerID               uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:owner_id"`
	opadata.FilteredModel `opa:"type:model"`
	types.Audit
	types.SoftDelete
}

func (ModelA) TableName() string {
	return "test_opa_model_a"
}
