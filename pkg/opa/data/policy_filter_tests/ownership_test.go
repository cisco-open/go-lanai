package policy_filter_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opadata "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data/constraints"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data/policy_filter_tests/testdata"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
	"time"
)

/*************************
	Test Setup
 *************************/

var TypicalSharing = constraints.Sharing{
	testdata.MockedUserId2: []constraints.SharedPermission{constraints.SharedPermissionRead, constraints.SharedPermissionWrite},
	testdata.MockedUserId3: []constraints.SharedPermission{constraints.SharedPermissionRead},
}

/*************************
	Test
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type OwnerTestDI struct {
	fx.In
	dbtest.DI
}

func TestOPAFilterWithOwnership(t *testing.T) {
	di := &OwnerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(10*time.Minute),
		dbtest.WithDBPlayback("testdb"),
		opatest.WithBundles(opatest.DefaultBundleFS, testdata.ModelBBundleFS),
		apptest.WithConfigFS(testdata.ConfigFS),
		apptest.WithFxOptions(),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareModelB(&di.DI)),
		test.GomegaSubTest(SubTestModelBCreate(di), "TestModelBCreate"),
		test.GomegaSubTest(SubTestModelBCreateByMap(di), "TestModelBCreateByMap"),
		test.GomegaSubTest(SubTestModelBList(di), "TestModelBList"),
		test.GomegaSubTest(SubTestModelBUpdate(di), "TestModelBUpdate"),
		test.GomegaSubTest(SubTestModelBDelete(di), "TestModelBDelete"),
		test.GomegaSubTest(SubTestModelBUpdateOwner(di), "TestModelBUpdateOwner"),
		test.GomegaSubTest(SubTestModelBUpdateSharing(di), "TestModelBUpdateSharing"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestPrepareModelB(di *dbtest.DI) test.SetupFunc {
	var models []*ModelB
	var share []*Shared
	closure := func(ctx context.Context, db *gorm.DB) {
		resetIdLookup()
		const more = 9
		extra := make([]*ModelB, 0, len(models)*more)
		for _, m := range models {
			key := LookupKey{Owner: m.OwnerID}
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
		dbtest.SetupWithGormScopes(opadata.SkipPolicyFiltering()),
		dbtest.SetupUsingSQLFile(testdata.ModelDataFS, "create_table_b.sql"),
		dbtest.SetupTruncateTables(ModelB{}.TableName(), Shared{}.TableName()),
		dbtest.SetupUsingModelSeedFile(testdata.ModelDataFS, &models, "model_b.yml", closure),
		dbtest.SetupUsingModelSeedFile(testdata.ModelDataFS, &share, "model_b_share.yml"),
	)
}

func SubTestModelBCreate(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var model, model2 ModelB
		var rs *gorm.DB
		model = ModelB{
			Value:     "test created",
			OwnerName: "user1",
			OwnerID:   testdata.MockedUserId1,
			Sharing:   TypicalSharing,
		}
		model2 = model
		// user1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		rs = di.DB.WithContext(ctx).Create(&model)
		assertDBResult(ctx, g, rs, "create model with correct owner ID", nil, 1)

		// user1 - batch
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		model2.ID = uuid.New()
		rs = di.DB.WithContext(ctx).CreateInBatches([]*ModelB{&model, &model2}, 10)
		assertDBResult(ctx, g, rs, "batch create models with correct owner ID", nil, 2)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		rs = di.DB.WithContext(ctx).Create(&model)
		assertDBResult(ctx, g, rs, "create model with incorrect owner ID", opa.ErrAccessDenied, 0)

		// user2 - batch
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		model.ID = uuid.New()
		model.OwnerID = testdata.MockedUserId2 // this is correct
		model2.ID = uuid.New()
		rs = di.DB.WithContext(ctx).CreateInBatches([]*ModelB{&model, &model2}, 10)
		assertDBResult(ctx, g, rs, "batch create model with incorrect owner ID", opa.ErrAccessDenied, 0)
	}
}

func SubTestModelBCreateByMap(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		modelMap := map[string]interface{}{
			"ID":        uuid.New(),
			"Value":     "test created",
			"OwnerName": "user1",
			"OwnerID":   testdata.MockedUserId1,
			"Sharing":   TypicalSharing,
			"CreatedAt": time.Now(),
			"CreatedBy": testdata.MockedUserId1,
		}
		// user1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Create(shallowCopyMap(modelMap))
		assertDBResult(ctx, g, rs, "create model with correct owner ID", nil, 1)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		modelMap["ID"] = uuid.New()
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Create(shallowCopyMap(modelMap))
		assertDBResult(ctx, g, rs, "create model with incorrect owner ID", opa.ErrAccessDenied, 0)
	}
}

func SubTestModelBList(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelB
		var rs *gorm.DB
		// user1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(80), "user1 should see %d models", 80)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(80), "user1 should see %d models", 80)
	}
}

func SubTestModelBUpdate(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const NewValue = `Updated`
		var id uuid.UUID
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(&ModelB{Value: NewValue})
		assertDBResult(ctx, g, rs, "update as owner", nil, 1)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "update as owner", "Value", NewValue)

		// user3 - not owner, but shared with write permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User3SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId3)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(map[string]interface{}{"value": NewValue})
		assertDBResult(ctx, g, rs, "update with shared permission", nil, 1)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "update with shared permission", "Value", NewValue)

		// user2 - not owner, not shared with write permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(&ModelB{Value: NewValue})
		assertDBResult(ctx, g, rs, "update model without shared write permission", nil, 0)
	}
}

func SubTestModelBDelete(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var id uuid.UUID
		var rs *gorm.DB
		// user3 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User3SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId3)
		rs = di.DB.WithContext(ctx).Delete(&ModelB{ID: id})
		assertDBResult(ctx, g, rs, "delete model as owner", nil, 1)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "delete model as owner")

		// user3 - not owner, shared with delete
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User3SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Delete(&ModelB{ID: id})
		assertDBResult(ctx, g, rs, "delete model with shared permission", nil, 1)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "delete model with permission")

		// user2 - not owner, not shared with delete (read only)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Delete(&ModelB{ID: id})
		assertDBResult(ctx, g, rs, "delete model of others", nil, 0)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "delete model of others", "exists")
	}
}

func SubTestModelBUpdateOwner(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var NewValue = uuid.MustParse("a5aaa07a-e7d7-4f66-bec8-1e651badacbd")
		var id uuid.UUID
		var rs *gorm.DB
		// user1 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(&ModelB{OwnerID: NewValue})
		assertDBResult(ctx, g, rs, "change model's owner as owner", nil, 0)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "change model's owner as owner", "OwnerID", testdata.MockedUserId1)

		// user3 - not owner, shared with write, no special permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User3SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(&ModelB{OwnerID: NewValue})
		assertDBResult(ctx, g, rs, "change model's owner with shared permission", nil, 0)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "change model's owner with shared permission", "OwnerID", testdata.MockedUserId1)

		// user2 - not owner, but have special permission
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(), testdata.ExtraPermsSecurityOptions("MANAGE"))
		id = findIDByOwner(testdata.MockedUserId1)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(map[string]interface{}{"owner_id": NewValue})
		assertDBResult(ctx, g, rs, "change model's owner with special permission", nil, 1)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "change model's owner with special permission", "OwnerID", NewValue)
	}
}

func SubTestModelBUpdateSharing(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var id uuid.UUID
		var rs *gorm.DB
		// prepare some data
		id = findIDByOwner(testdata.MockedUserId2)
		model, e := loadModel[ModelB](ctx, di.DB, id)
		g.Expect(e).To(Succeed(), "original model should be exists")
		var OriginalSharing = constraints.Sharing(shallowCopyMap(model.Sharing))
		var UpdatedSharing = constraints.Sharing(shallowCopyMap(model.Sharing))
		UpdatedSharing.Share(testdata.MockedUserId1, constraints.SharedPermissionWrite, constraints.SharedPermissionRead)

		// user2 - owner
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId2)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(&ModelB{Sharing: UpdatedSharing})
		assertDBResult(ctx, g, rs, "change model's sharing as owner", nil, 1)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "change model's sharing as owner", "Sharing", UpdatedSharing)

		// user3 - not owner, shared with "share"
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User3SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId2)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(map[string]interface{}{"sharing": UpdatedSharing})
		assertDBResult(ctx, g, rs, "change model's sharing with shared permission", nil, 1)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "change model's sharing with shared permission", "Sharing", UpdatedSharing)

		// user2 - not owner, and not shared with "share
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		id = findIDByOwner(testdata.MockedUserId2)
		rs = di.DB.WithContext(ctx).Model(&ModelB{ID: id}).Updates(&ModelB{Sharing: UpdatedSharing})
		assertDBResult(ctx, g, rs, "change model's sharing without shared permission", nil, 0)
		assertPostOpModel[ModelB](ctx, g, di.DB, id, "change model's sharing without shared permission", "Sharing", OriginalSharing)
	}
}

/*************************
	Helpers
 *************************/

/*************************
	Models
 *************************/

// ModelB has no tenancy
type ModelB struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value           string
	OwnerName       string
	OwnerID         uuid.UUID            `gorm:"type:KeyID;not null" opa:"field:owner_id"`
	Sharing         constraints.Sharing  `opa:"field:share"`
	OPAPolicyFilter opadata.PolicyFilter `gorm:"-" opa:"type:model"`
	types.Audit
	// For testing utils only
	Shared []*Shared `gorm:"foreignKey:ResID;references:ID" opa:"field:shared"`
}

func (ModelB) TableName() string {
	return "test_opa_model_b"
}

type Shared struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	ResID      uuid.UUID
	UserID     uuid.UUID
	Operations pq.StringArray `gorm:"type:string[];" opa:"field:operations"`
	Username   string
}

func (Shared) TableName() string {
	return "test_opa_model_b_share"
}
