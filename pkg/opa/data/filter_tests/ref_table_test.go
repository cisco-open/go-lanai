package filter_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	opadata "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data/filter_tests/testdata"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"fmt"
	"github.com/google/uuid"
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

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

func SetupCustomRefTable(db *gorm.DB) error {
	if e := db.SetupJoinTable(&ModelDUser{}, "Models", &ModelDRef{}); e != nil {
		return e
	}
	if e := db.SetupJoinTable(&ModelD{}, "Users", &ModelDRef{}); e != nil {
		return e
	}
	return nil
}

/*************************
	Test
 *************************/

func TestOPAFilterOnRelations(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(10*time.Minute),
		dbtest.WithDBPlayback("testdb"),
		opatest.WithBundles(opatest.DefaultBundleFS, testdata.ModelDBundleFS),
		apptest.WithModules(tenancy.Module),
		apptest.WithConfigFS(testdata.ConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.ProvideMockedTenancyAccessor),
			fx.Invoke(SetupCustomRefTable),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareModelD(&di.DI)),
		test.GomegaSubTest(SubTestModelDNewRelations(di), "TestModelNewRelations"),
		test.GomegaSubTest(SubTestModelDNewForeignRelations(di), "TestModelNewForeignRelations"),
		test.GomegaSubTest(SubTestModelDListByPreload(di), "TestModelListByPreload"),
		test.GomegaSubTest(SubTestModelDListWithPreload(di), "TestModelListWithPreload"),
		test.GomegaSubTest(SubTestModelDListByRef(di), "TestModelListByRef"),
		test.GomegaSubTest(SubTestModelDRemoveRelations(di), "TestModelRemoveRelations"),
		test.GomegaSubTest(SubTestModelDRemoveForeignRelations(di), "TestModelRemoveForeignRelations"),
	)
}

/*************************
	Data Setup
 *************************/

func SetupTestPrepareModelD(di *dbtest.DI) test.SetupFunc {
	var users []*ModelDUser
	var models []*ModelD
	tables := []string{ModelDUser{}.TableName(), ModelD{}.TableName(), ModelDRef{}.TableName()}
	// We use special DB scope to prepare data, to by-pass policy filtering
	return dbtest.PrepareDataWithScope(di,
		dbtest.SetupWithGormScopes(opadata.SkipFiltering()),
		dbtest.SetupDropTables(tables...),
		dbtest.SetupUsingSQLFile(testdata.ModelDataFS, "create_table_d.sql"),
		dbtest.SetupUsingModelSeedFile(testdata.ModelDataFS, &users, "model_d_user.yml"),
		dbtest.SetupUsingModelSeedFile(testdata.ModelDataFS, &models, "model_d.yml"),
		SetupModelDRelationshipData(&users, &models),
	)
}

func SetupModelDRelationshipData(usersPtr *[]*ModelDUser, modelsPtr *[]*ModelD) dbtest.DataSetupStep {
	return func(ctx context.Context, t *testing.T, db *gorm.DB) context.Context {
		g := gomega.NewWithT(t)
		pool, e := testdata.NewUUIDPool()
		g.Expect(e).To(Succeed(), "loading UUID pool should succeed")
		//e = SetupCustomRefTable(db)
		//g.Expect(e).To(Succeed(), "setting custom table shouldn't fail")
		const more = 9
		const perUser = 4
		users := *usersPtr
		models := *modelsPtr
		resetIdLookup()
		extra := make([]*ModelD, 0, len(models)*more)
		refs := make([]*ModelDRef, 0, len(users) * len(models) * perUser)
		for _, m := range models {
			dups := make([]*ModelD, more + 1)
			dups[0] = m
			// duplicate more models with same template
			for i := 0; i < more; i++ {
				newM := ModelD{
					ID:            pool.PopOrNew(),
					Value:         fmt.Sprintf("%s - Dup %d", m.Value, i),
					TenantName:    m.TenantName,
					TenantID:      m.TenantID,
					TenantPath:    m.TenantPath,
				}
				extra = append(extra, &newM)
				dups[i+1] = &newM
			}
			// setup many-to-many relations and prepare lookup
			idx := 0
			for _, u := range users {
				for i := 0; i < perUser; i++ {
					dup := dups[idx]
					key := LookupKey{Tenant: dup.TenantID, Owner: u.ID}
					prepareIdLookup(dup.ID, key)
					refs = append(refs, &ModelDRef{UserID: u.ID, ModelID: dup.ID})
					idx = (idx + 1) % len(dups)
				}
			}
		}
		rs := db.WithContext(ctx).CreateInBatches(extra, 50)
		g.Expect(rs.Error).To(Succeed(), "creating extra models shouldn't fail")
		rs = db.WithContext(ctx).CreateInBatches(refs, 50)
		g.Expect(rs.Error).To(Succeed(), "creating user model references shouldn't fail")
		// need to clean up models for repeated setup
		*usersPtr = nil
		*modelsPtr = nil
		return ctx
	}
}

/*************************
	Sub Tests
 *************************/

// SubTestModelDNewRelations add association from ModelD side.
// # Note: If we allow someone to assign many2many relationships, we also need to allow user to update "updated_at" field in policy.
func SubTestModelDNewRelations(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var model *ModelD
		//var rs *gorm.DB
		var e error
		user := mustLoadModel[ModelDUser](ctx, g, di.DB, testdata.MockedUserId1)
		// user1 - try to add user1 to a model under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA1))
		model = mustLoadModel[ModelD](ctx, g, di.DB, findIDByTenant(testdata.MockedTenantIdA1))
		e = di.DB.WithContext(ctx).Model(model).Omit("Users.*").Association("Users").Append(user)
		g.Expect(e).To(Succeed(), "adding current user to model should not fail")

		// user2 - try to add user1 to a model under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdA1))
		model = mustLoadModel[ModelD](ctx, g, di.DB, findIDByTenant(testdata.MockedTenantIdA1))
		e = di.DB.WithContext(ctx).Model(model).Omit("Users.*").Association("Users").Append(user)
		g.Expect(e).To(HaveOccurred(), "adding another user to model should fail")

		// admin - try to add all models under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.AdminSecurityOptions(testdata.MockedTenantIdA1))
		model = mustLoadModel[ModelD](ctx, g, di.DB, findIDByTenant(testdata.MockedTenantIdA1))
		e = di.DB.WithContext(ctx).Model(model).Omit("Users.*").Association("Users").Append(user)
		g.Expect(e).To(Succeed(), "adding current user to model as an admin should not fail")
	}
}

// SubTestModelDNewForeignRelations add association from foreign side (User)
func SubTestModelDNewForeignRelations(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelD
		var rs *gorm.DB
		var e error
		user := mustLoadModel[ModelDUser](ctx, g, di.DB, testdata.MockedUserId1)
		// user1 - try to add all models under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA1))
		rs = di.DB.WithContext(ctx).Model(&ModelD{}).Find(&models)
		assertDBResult(ctx, g, rs, "find models under selected tenant", nil, 10)
		e = di.DB.WithContext(ctx).Model(&user).Omit("Models.*").Association("Models").Append(models)
		g.Expect(e).To(Succeed(), "add new relations to current user should not fail")

		// user2 - try to add all models under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdA1))
		rs = di.DB.WithContext(ctx).Model(&ModelD{}).Find(&models)
		assertDBResult(ctx, g, rs, "find models under selected tenant", nil, 10)
		e = di.DB.WithContext(ctx).Model(&user).Omit("Models.*").Association("Models").Append(models)
		g.Expect(e).To(HaveOccurred(), "add new relations to another user should fail")

		// admin - try to add all models under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.AdminSecurityOptions(testdata.MockedTenantIdA1))
		rs = di.DB.WithContext(ctx).Model(&ModelD{}).Find(&models)
		assertDBResult(ctx, g, rs, "find models under selected tenant", nil, 10)
		e = di.DB.WithContext(ctx).Model(&user).Omit("Models.*").Association("Models").Append(models)
		g.Expect(e).To(Succeed(), "add new relations to another user as admin should not fail")
	}
}

func SubTestModelDListByPreload(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelDUser
		var rs *gorm.DB
		// user1 - parent tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelDUser{}).Preload("Models").Find(&models)
		assertDBResult(ctx, g, rs, "list users as regular user", nil, 3)
		g.Expect(models).To(HaveLen(3), "user1 should see %d users", 3)
		for _, m := range models {
			if m.ID == testdata.MockedUserId1 {
				g.Expect(m.Models).To(HaveLen(16), "user1 should see %d models of user1", 16)
			} else {
				g.Expect(m.Models).To(HaveLen(0), "user1 should see %d models of other users", 0)
			}
		}

		// admin - parent tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.AdminSecurityOptions(testdata.MockedTenantIdA))

		rs = di.DB.WithContext(ctx).Model(&ModelDUser{}).Preload("Models").Find(&models)
		assertDBResult(ctx, g, rs, "list users as admin", nil, 3)
		g.Expect(models).To(HaveLen(3), "admin should see %d users", 3)
		for _, m := range models {
			g.Expect(m.Models).To(HaveLen(16), "admin should see %d models of any user", 16)
		}
	}
}

func SubTestModelDListWithPreload(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelD
		var rs *gorm.DB
		// user1 - parent tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelD{}).Preload("Users").Find(&models)
		assertDBResult(ctx, g, rs, "list models as regular user", nil, 40)
		g.Expect(models).To(HaveLen(40), "user1 should see %d models", 40)
		for _, m := range models {
			g.Expect(len(m.Users)).To(BeNumerically("<=", 1), "user1 should only see user1 in association")
		}

		// admin - parent tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.AdminSecurityOptions(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelD{}).Preload("Users").Find(&models)
		assertDBResult(ctx, g, rs, "list models as admin", nil, 40)
		g.Expect(models).To(HaveLen(40), "admin should see %d models", 40)
		for _, m := range models {
			g.Expect(len(m.Users)).To(BeNumerically(">=", 1), "admin should see all association")
		}
	}
}

func SubTestModelDListByRef(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelDRef
		var rs *gorm.DB
		var count int
		// user1 - parent tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelDRef{}).Distinct("ModelID").Preload("Model").Find(&models)
		assertDBResult(ctx, g, rs, "list refs as regular user", nil, 36)
		g.Expect(models).To(HaveLen(36), "user1 should see %d refs", 36)
		count = 0
		for _, m := range models {
			if m.Model != nil {
				count ++
			}
		}
		g.Expect(count).To(Equal(16), "user1 should see %d models", 16)

		// admin - parent tenant A
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.AdminSecurityOptions(testdata.MockedTenantIdA))
		rs = di.DB.WithContext(ctx).Model(&ModelDRef{}).Distinct("ModelID").Preload("Model").Find(&models)
		assertDBResult(ctx, g, rs, "list refs as regular user", nil, 90)
		g.Expect(models).To(HaveLen(90), "user1 should see %d refs", 90)
		count = 0
		for _, m := range models {
			if m.Model != nil {
				count ++
			}
		}
		g.Expect(count).To(Equal(40), "user1 should see %d models", 40)
	}
}

// SubTestModelDRemoveRelations remove association from ModelD side.
// # Note: If we allow someone to assign many2many relationships, we also need to allow user to update "updated_at" field in policy.
func SubTestModelDRemoveRelations(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var model *ModelD
		var e error
		var id uuid.UUID
		user := mustLoadModel[ModelDUser](ctx, g, di.DB, testdata.MockedUserId1)
		// user1 - try to remove user1 from a model under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA1))
		id = findIDByOwner(testdata.MockedUserId1)
		model = mustLoadModel[ModelD](ctx, g, di.DB, id)
		e = di.DB.WithContext(ctx).Model(model).Association("Users").Delete(user)
		g.Expect(e).To(Succeed(), "removing current user from model should not fail")
		assertPostOpModel[ModelDRef](ctx, g, di.DB, &ModelDRef{UserID: testdata.MockedUserId1, ModelID: id}, "delete")

		// user2 - try to add user1 to a model under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdA1))
		id = findIDByOwner(testdata.MockedUserId1)
		model = mustLoadModel[ModelD](ctx, g, di.DB, id)
		e = di.DB.WithContext(ctx).Model(model).Association("Users").Delete(user)
		g.Expect(e).To(Succeed(), "removing another user from model should not fail")
		assertPostOpModel[ModelDRef](ctx, g, di.DB, &ModelDRef{UserID: testdata.MockedUserId1, ModelID: id}, "delete", "exists")

		// admin - try to add all models under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.AdminSecurityOptions(testdata.MockedTenantIdA1))
		id = findIDByOwner(testdata.MockedUserId1)
		model = mustLoadModel[ModelD](ctx, g, di.DB, id)
		e = di.DB.WithContext(ctx).Model(model).Association("Users").Delete(user)
		g.Expect(e).To(Succeed(), "removing current user from model as an admin should not fail")
		assertPostOpModel[ModelDRef](ctx, g, di.DB, &ModelDRef{UserID: testdata.MockedUserId1, ModelID: id}, "delete")
	}
}

// SubTestModelDRemoveForeignRelations remove association from foreign side (User)
func SubTestModelDRemoveForeignRelations(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var model *ModelD
		var e error
		var id uuid.UUID
		user := mustLoadModel[ModelDUser](ctx, g, di.DB, testdata.MockedUserId1)
		// user1 - try to remove user1 from a model under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions(testdata.MockedTenantIdA1))
		id = findIDByOwner(testdata.MockedUserId1)
		model = mustLoadModel[ModelD](ctx, g, di.DB, id)
		e = di.DB.WithContext(ctx).Model(user).Association("Models").Delete(model)
		g.Expect(e).To(Succeed(), "removing model from current user should not fail")
		assertPostOpModel[ModelDRef](ctx, g, di.DB, &ModelDRef{UserID: testdata.MockedUserId1, ModelID: id}, "delete")

		// user2 - try to add user1 to a model under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions(testdata.MockedTenantIdA1))
		id = findIDByOwner(testdata.MockedUserId1)
		model = mustLoadModel[ModelD](ctx, g, di.DB, id)
		e = di.DB.WithContext(ctx).Model(user).Association("Models").Delete(model)
		g.Expect(e).To(Succeed(), "removing model from another user should not fail")
		assertPostOpModel[ModelDRef](ctx, g, di.DB, &ModelDRef{UserID: testdata.MockedUserId1, ModelID: id}, "delete", "exists")

		// admin - try to add all models under currently selected tenant (A-1)
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.AdminSecurityOptions(testdata.MockedTenantIdA1))
		id = findIDByOwner(testdata.MockedUserId1)
		model = mustLoadModel[ModelD](ctx, g, di.DB, id)
		e = di.DB.WithContext(ctx).Model(user).Association("Models").Delete(model)
		g.Expect(e).To(Succeed(), "removing model from current user as an admin should not fail")
		assertPostOpModel[ModelDRef](ctx, g, di.DB, &ModelDRef{UserID: testdata.MockedUserId1, ModelID: id}, "delete")
	}
}

/*************************
	Helpers
 *************************/

/*************************
	Models
 *************************/

type ModelDUser struct {
	ID               uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Username string
	Models   []*ModelD `gorm:"many2many:test_opa_model_d_ref;joinForeignKey:user_id;joinReferences:model_id;"`
	// ModelDRelations is read-only, don't set associations via this field
	ModelDRelations []*ModelDRef `gorm:"foreignKey:UserID;<-:false"`
	types.Audit
}

func (ModelDUser) TableName() string {
	return "test_opa_model_d_user"
}

type ModelD struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value      string
	TenantName string
	TenantID   uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:tenant_id"`
	TenantPath pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null" opa:"field:tenant_path"`
	Users      []*ModelDUser `gorm:"many2many:test_opa_model_d_ref;joinForeignKey:model_id;joinReferences:user_id;"`
	// UserRelations is read-only, don't set associations via this field
	UserRelations []*ModelDRef `gorm:"foreignKey:ModelID;<-:false"`

	opadata.FilteredModel `opa:"type:model"`
	types.Audit
}

func (ModelD) TableName() string {
	return "test_opa_model_d"
}

type ModelDRef struct {
	// Note: gorm@1.25.0/callbacks/create.go ln219 ConvertToCreateValues(stmt)
	// 		 The "Create()" function uses map to generate INSERT statement's fields that have "default" gorm tag.
	// 		 Therefore, if the model have multiple fields with "default" gorm tag (like following commented out fields),
	//		 the order of them in INSERT statement is undefined, which is not friendly with Copyist lib.
	//UserID               uuid.UUID   `gorm:"primaryKey;type:uuid;default:gen_random_uuid();" opa:"field:owner_id"`
	//ModelID              uuid.UUID   `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`

	UserID               uuid.UUID   `gorm:"primaryKey;type:uuid;" opa:"field:owner_id"`
	ModelID              uuid.UUID   `gorm:"primaryKey;type:uuid;"`
	User                 *ModelDUser `gorm:"foreignKey:user_id"`
	Model                 *ModelD     `gorm:"foreignKey:model_id"`
	opadata.FilteredModel `opa:"type:user_model_ref"`
	types.Audit
}

func (ModelDRef) TableName() string {
	return "test_opa_model_d_ref"
}
