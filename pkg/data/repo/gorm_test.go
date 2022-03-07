package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
	"sort"
	"testing"
	"time"
)

var (
	mtoModelIDs = []uuid.UUID{
		uuid.MustParse("f73f7553-22eb-4cc8-87e2-3cd4fc005d55"),
		uuid.MustParse("fd768d0a-f682-47cd-bdc3-c8df264c8825"),
		uuid.MustParse("987ba819-9df5-4593-a64e-fb354686f48f"),
	}
	mtmModelIDs = []uuid.UUID{
		uuid.MustParse("7985f15d-c263-424e-a8da-450bd68fc9d3"),
		uuid.MustParse("349121b5-a0d1-4345-a29e-ef8fb4daeca3"),
		uuid.MustParse("0262915d-da46-42f5-be58-b14eb46dfe17"),
	}
	modelIDs = []uuid.UUID{
		uuid.MustParse("016c5ffa-d70b-4cd3-b80b-40de34f37ee2"),
		uuid.MustParse("9445da76-927c-488d-afba-da5c091ab9df"),
		uuid.MustParse("51018740-10ad-470b-9b2c-bc07461cbeb8"),
		uuid.MustParse("2927a658-1c0e-4117-8508-4b32ec41c684"),
		uuid.MustParse("1558612f-cb71-4e99-b903-b1393b25cb63"),
		uuid.MustParse("64d930d5-a205-45ce-9b69-ba0543699fbc"),
		uuid.MustParse("efe60947-605a-442b-b097-a6eb682cbedd"),
		uuid.MustParse("9c5fbf48-7154-4222-9af2-1f04140fc042"),
		uuid.MustParse("576286d7-0ee0-46b1-af18-3de00415eb8a"),
	}
	sortedModelIDs []uuid.UUID
	nonExistID = uuid.MustParse("2e15c4d2-d427-4af2-b0d9-c3bcb0a8485c")
)

func init() {
	sortedModelIDs = make([]uuid.UUID, len(modelIDs))
	for i, v := range modelIDs {
		sortedModelIDs[i] = v
	}
	sort.SliceStable(sortedModelIDs, func(i, j int) bool {
		return lessUUID(sortedModelIDs[i], sortedModelIDs[j])
	})
}

func lessUUID(l, r uuid.UUID) bool {
	for i := range l {
		if l[i] != r[i] {
			return l[i] < r[i]
		}
	}
	return false
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
	DB   *gorm.DB
	Repo TestRepository
}

func TestGormSchemaResolver(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(Module),
		apptest.WithTimeout(time.Minute),
		apptest.WithFxOptions(
			fx.Provide(NewTestRepository),
		),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestSchemaResolverDirect(di), "TestSchemaResolverDirect"),
		test.GomegaSubTest(SubTestSchemaResolverIndirect(di), "TestSchemaResolverIndirect"),
		test.GomegaSubTest(SubTestSchemaResolverMultiLvl(di), "TestSchemaResolverMultiLvl"),
	)
}

func TestGormCRUDRepository(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(Module),
		apptest.WithTimeout(time.Minute),
		apptest.WithFxOptions(
			fx.Provide(NewTestRepository),
		),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareTables(di)),
		test.GomegaSubTest(SubTestFindOne(di), "TestFindOne"),
		test.GomegaSubTest(SubTestFindAll(di), "TestFindAll"),
		test.GomegaSubTest(SubTestCount(di), "TestFindCount"),
		test.GomegaSubTest(SubTestCreate(di), "TestCreate"),
		test.GomegaSubTest(SubTestSave(di), "TestSave"),
		test.GomegaSubTest(SubTestUpdates(di), "TestUpdates"),
		test.GomegaSubTest(SubTestDelete(di), "TestDelete"),
		test.GomegaSubTest(SubTestRepoSyntax(di), "TestRepoSyntax"),
		test.GomegaSubTest(SubTestPageAndSort(di), "TestPageAndSort"),
		test.GomegaSubTest(SubTestTransaction(di), "TestTransaction"),
		test.GomegaSubTest(SubTestUtilFunctions(di), "TestUtilFunctions"),
	)
}

func TestGormUtils(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(Module),
		apptest.WithTimeout(time.Minute),
		apptest.WithFxOptions(
			fx.Provide(NewTestRepository),
		),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareTables(di)),
		test.GomegaSubTest(SubTestResolveSchema(di), "TestResolveSchema"),
		test.GomegaSubTest(SubTestCheckUniqueness(di), "TestCheckUniqueness"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestPrepareTables(di *testDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		prepareTable(di.DB, g)
		prepareRepoTestData(di.DB, g)
		return ctx, nil
	}
}

func SubTestSchemaResolverDirect(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Repo.ModelType()).
			To(BeEquivalentTo(reflect.TypeOf(TestModel{})), "ModelType should be correct")
		g.Expect(di.Repo.Table()).
			To(Equal("test_repo_models"), "Table should be correct")
		g.Expect(di.Repo.ColumnName("Value")).
			To(Equal("value"), "ColumnName of direct field should be correct")
		g.Expect(di.Repo.ColumnDataType("Value")).
			To(Equal("string"), "ColumnDataType of direct field should be correct")

		g.Expect(di.Repo.ColumnName("Unknown")).
			To(Equal(""), "ColumnName of unknown field should be empty")
		g.Expect(di.Repo.ColumnDataType("Unknown")).
			To(Equal(""), "ColumnDataType of unknown field should be empty")

		g.Expect(di.Repo.(GormSchemaResolver).Schema()).
			To(Not(BeNil()), "Schema() should be correct")
	}
}

func SubTestSchemaResolverIndirect(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// one to one
		g.Expect(di.Repo.ColumnName("OneToOne.RelationValue")).
			To(Equal("value"), "ColumnName of one-to-one field should be correct with field path")
		g.Expect(di.Repo.ColumnDataType("OneToOne.RelationValue")).
			To(Equal("string"), "ColumnDataType of one-to-one field should be correct with field path")

		resolver := di.Repo.RelationshipSchema("OneToOne")
		g.Expect(resolver.ModelType()).
			To(BeEquivalentTo(reflect.TypeOf(TestOTOModel{})), "ModelType of one-to-one relation should be correct")
		g.Expect(resolver.Table()).
			To(Equal("test_repo_model1"), "Table of one-to-one relation should be correct")
		g.Expect(resolver).To(Not(BeNil()), "RelationshipSchema of one-to-one shouldn't be nil")
		g.Expect(resolver.ColumnName("RelationValue")).
			To(Equal("value"), "ColumnName of one-to-one model's schema should be correct")
		g.Expect(resolver.ColumnDataType("RelationValue")).
			To(Equal("string"), "ColumnDataType of one-to-one model's schema should be correct")

		// many to one
		g.Expect(di.Repo.ColumnName("ManyToOne.RelationValue")).
			To(Equal("value"), "ColumnName of many-to-one field should be correct with field path")
		g.Expect(di.Repo.ColumnDataType("ManyToOne.RelationValue")).
			To(Equal("string"), "ColumnDataType of many-to-one field should be correct with field path")

		resolver = di.Repo.RelationshipSchema("ManyToOne")
		g.Expect(resolver.ModelType()).
			To(BeEquivalentTo(reflect.TypeOf(TestMTOModel{})), "ModelType of many-to-one relation should be correct")
		g.Expect(resolver.Table()).
			To(Equal("test_repo_model2"), "Table of many-to-one relation should be correct")
		g.Expect(resolver).To(Not(BeNil()), "RelationshipSchema of many-to-one shouldn't be nil")
		g.Expect(resolver.ColumnName("RelationValue")).
			To(Equal("value"), "ColumnName of many-to-one model's schema should be correct")
		g.Expect(resolver.ColumnDataType("RelationValue")).
			To(Equal("string"), "ColumnDataType of many-to-one model's schema should be correct")
	}
}

func SubTestSchemaResolverMultiLvl(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Repo.ColumnName("ManyToOne.RelatedMTMModels.MTMValue")).
			To(Equal("value"), "ColumnName of many-to-many field should be correct with field path")
		g.Expect(di.Repo.ColumnDataType("ManyToOne.RelatedMTMModels.MTMValue")).
			To(Equal("string"), "ColumnDataType of many-to-many field should be correct with field path")

		resolver := di.Repo.RelationshipSchema("ManyToOne.RelatedMTMModels")
		g.Expect(resolver.ModelType()).
			To(BeEquivalentTo(reflect.TypeOf(TestMTMModel{})), "ModelType of many-to-many relation should be correct")
		g.Expect(resolver.Table()).
			To(Equal("test_repo_model3"), "Table of many-to-many relation should be correct")
		g.Expect(resolver).To(Not(BeNil()), "RelationshipSchema of many-to-many shouldn't be nil")
		g.Expect(resolver.ColumnName("MTMValue")).
			To(Equal("value"), "ColumnName of many-to-many model's schema should be correct")
		g.Expect(resolver.ColumnDataType("MTMValue")).
			To(Equal("string"), "ColumnDataType of many-to-many model's schema should be correct")

	}
}

func SubTestFindOne(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var model TestModel
		var conditions []Condition
		// find by id
		e = di.Repo.FindById(ctx, &model, modelIDs[0],
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(e).To(Succeed(), "FindById shouldn't return error")
		assertFullyFetchedTestModel(&model, g, "FindById")

		// find one
		model = TestModel{}
		conditions = []Condition{
			Where("test_repo_models.search > 0"),
			clause.Where{Exprs: []clause.Expression{clause.Lt{
				Column: clause.Column{Table: clause.CurrentTable, Name: di.Repo.ColumnName("SearchIdx")},
				Value:  2,
			}}},
		}
		e = di.Repo.FindOneBy(ctx, &model, conditions,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(e).To(Succeed(), "FindOneBy shouldn't return error")
		g.Expect(model.ID).To(BeEquivalentTo(modelIDs[1]), "FindOneBy return correct result")
		assertFullyFetchedTestModel(&model, g, "FindOneBy")

		model = TestModel{}
		conditions = []Condition{
			`"ManyToOne"."search" = 9999`,
		}
		e = di.Repo.FindOneBy(ctx, &model, conditions,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(BeTrue(), "FindOneBy should return RecordNotFound error")
	}
}

func SubTestFindAll(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var models []*TestModel
		var conditions []Condition
		// find all
		models = nil
		e = di.Repo.FindAll(ctx, &models,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(e).To(Succeed(), "FindAll shouldn't return error")
		g.Expect(models).To(HaveLen(len(modelIDs)), "FindAll should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of FindAll")
		}

		models = nil
		e = di.Repo.FindAll(ctx, &models,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(e).To(Succeed(), "FindAll shouldn't return error")
		g.Expect(models).To(HaveLen(len(modelIDs)), "FindAll should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of FindAll")
		}

		// find all paginated
		models = nil
		e = di.Repo.FindAll(ctx, &models,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
			Page(1, 5),
		)
		g.Expect(e).To(Succeed(), "paginated FindAll shouldn't return error")
		g.Expect(models).To(HaveLen(4), "paginated FindAll should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of paginated FindAll")
		}

		// find all by
		models = nil
		conditions = []Condition{
			`test_repo_models.search > 0 AND "OneToOne".search > 0`,
		}
		e = di.Repo.FindAllBy(ctx, &models, conditions,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(e).To(Succeed(), "paginated FindAllBy shouldn't return error")
		g.Expect(models).To(HaveLen(7), "paginated FindAllBy should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of paginated FindAllBy")
		}

		// find all by, paginated
		models = nil
		e = di.Repo.FindAllBy(ctx, &models, conditions,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
			Page(1, 5),
		)
		g.Expect(e).To(Succeed(), "FindAllBy shouldn't return error")
		g.Expect(models).To(HaveLen(2), "FindAllBy should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of FindAllBy")
		}
	}
}

func SubTestCount(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var count int
		count, e = di.Repo.CountAll(ctx)
		g.Expect(e).To(Succeed(), "CountAll shouldn't return error")
		g.Expect(count).To(BeEquivalentTo(len(modelIDs)), "CountAll should return correct result")

		conditions := []Condition{
			`"ManyToOne".search > 0`,
		}
		count, e = di.Repo.CountBy(ctx, conditions, Joins("ManyToOne"))
		g.Expect(e).To(Succeed(), "CountBy shouldn't return error")
		g.Expect(count).To(BeEquivalentTo(6), "CountBy should return correct result")
	}
}

func SubTestCreate(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var models []*TestModel
		var otoModels []*TestOTOModel
		for i := 0; i < 9; i++ {
			m, oto := createMainModel(uuid.New(), i)
			models = append(models, m)
			otoModels = append(otoModels, oto)
		}
		rs := di.DB.WithContext(ctx).Create(otoModels)
		g.Expect(rs.Error).To(Succeed(), "Create OTO models shouldn't return error")

		e = di.Repo.Create(ctx, models, Omit("OneToOne"), Omit("ManyToOne"))
		g.Expect(e).To(Succeed(), "Create shouldn't return error")
	}
}

func SubTestSave(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var models []*TestModel
		e = di.Repo.FindAll(ctx, &models)
		g.Expect(e).To(Succeed(), "FindAll shouldn't return error")
		g.Expect(models).To(Not(BeEmpty()), "FindAll should return some records")

		// do save
		for _, m := range models {
			m.Value = "Updated Value"
		}
		e = di.Repo.Save(ctx, models)
		g.Expect(e).To(Succeed(), "Save shouldn't return error")

		// fetch again and validate
		models = nil
		e = di.Repo.FindAll(ctx, &models)
		g.Expect(e).To(Succeed(), "FindAll shouldn't return error")
		for _, m := range models {
			g.Expect(m.Value).To(BeEquivalentTo("Updated Value"), "saved values should be correct when re-fetch")
		}
	}
}

func SubTestUpdates(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		id := modelIDs[0]
		e = di.Repo.Update(ctx, &TestModel{ID: id}, &TestModel{Value: "Just Updated"}, ErrorOnZeroRows())
		g.Expect(e).To(Succeed(), "Update shouldn't return error")

		var model TestModel
		e = di.Repo.FindById(ctx, &model, id)
		g.Expect(e).To(Succeed(), "FindById shouldn't return error")
		g.Expect(model.Value).To(BeEquivalentTo("Just Updated"), "re-fetched value after Update should be correct")

		// update 0 rows by id without ErrorOnZeroRows option
		e = di.Repo.Update(ctx, &TestModel{ID: nonExistID}, &TestModel{Value: "Just Updated"})
		g.Expect(e).To(Succeed(), "Delete shouldn't return error with 0 affected rows")

		// delete 0 rows by with ErrorOnZeroRows option
		e = di.Repo.Update(ctx, &TestModel{ID: nonExistID}, &TestModel{Value: "Just Updated"}, ErrorOnZeroRows())
		g.Expect(e).To(HaveOccurred(), "Delete should return error with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(BeTrue(), "Delete should return gorm.ErrRecordNotFound with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, data.ErrorRecordNotFound)).To(BeTrue(), "Delete should return data.ErrorRecordNotFound with 0 affected rows and ErrorOnZeroRows")
	}
}

func SubTestDelete(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error

		// delete by id
		id := modelIDs[0]
		e = di.Repo.Delete(ctx, &TestModel{ID: id})
		g.Expect(e).To(Succeed(), "Delete shouldn't return error")

		var model TestModel
		e = di.Repo.FindById(ctx, &model, id, ErrorOnZeroRows())
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(BeTrue(), "re-fetch after Delete should yield RecordNotFound")

		// delete by
		e = di.Repo.DeleteBy(ctx, Where(`"test_repo_models"."search" < ?`, len(modelIDs)-1))
		g.Expect(e).To(Succeed(), "DeleteBy shouldn't return error")

		// delete 0 rows by id without ErrorOnZeroRows option
		e = di.Repo.Delete(ctx, &TestModel{ID: nonExistID})
		g.Expect(e).To(Succeed(), "Delete shouldn't return error with 0 affected rows")

		// delete 0 rows by with ErrorOnZeroRows option
		e = di.Repo.Delete(ctx, &TestModel{ID: nonExistID}, ErrorOnZeroRows())
		g.Expect(e).To(HaveOccurred(), "Delete should return error with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(BeTrue(), "Delete should return gorm.ErrRecordNotFound with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, data.ErrorRecordNotFound)).To(BeTrue(), "Delete should return data.ErrorRecordNotFound with 0 affected rows and ErrorOnZeroRows")

		// fetch again and validate
		var count int
		count, e = di.Repo.CountAll(ctx)
		g.Expect(e).To(Succeed(), "CountAll after Delete shouldn't return error")
		g.Expect(count).To(BeEquivalentTo(1), "Delete should result in 1 record left")
	}
}

func SubTestPageAndSort(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var models []*TestModel
		// by direct field
		e = di.Repo.FindAll(ctx, &models, SortBy("SearchIdx", true))
		g.Expect(e).To(Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(HaveLen(len(modelIDs)), "sort by field shouldn't change model count")
		g.Expect(models[0].SearchIdx).To(BeEquivalentTo(len(modelIDs)-1), "sorted result's first item should be correct")

		// by to-one relation field
		models = nil
		e = di.Repo.FindAll(ctx, &models, Joins("ManyToOne"), SortBy("ManyToOne.RelationSearch", true))
		g.Expect(e).To(Succeed(), "sort by relation's field shouldn't return error")
		g.Expect(models).To(HaveLen(len(modelIDs)), "sort by relation's field shouldn't change model count")
		g.Expect(models[0].ManyToOne.RelationSearch).To(BeEquivalentTo(len(mtoModelIDs)-1), "sorted result's first item should be correct")

		// by column
		e = di.Repo.FindAll(ctx, &models, Sort("search DESC"))
		g.Expect(e).To(Succeed(), "sort by column shouldn't return error")
		g.Expect(models).To(HaveLen(len(modelIDs)), "sort by column shouldn't change model count")
		g.Expect(models[0].SearchIdx).To(BeEquivalentTo(len(modelIDs)-1), "sorted result's first item should be correct")

		// page + sort
		models = nil
		e = di.Repo.FindAll(ctx, &models, Page(1, 2), SortBy("SearchIdx", false))
		g.Expect(e).To(Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].SearchIdx).To(BeEquivalentTo(2), "sorted page result should have correct order")

		// sort + page (reversed order)
		models = nil
		e = di.Repo.FindAll(ctx, &models, SortBy("SearchIdx", false), Page(1, 2))
		g.Expect(e).To(Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].SearchIdx).To(BeEquivalentTo(2), "sorted page result should have correct order")

		// page only
		models = nil
		e = di.Repo.FindAll(ctx, &models, Page(1, 2))
		g.Expect(e).To(Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].ID).To(BeEquivalentTo(sortedModelIDs[2]), "page result should have correct order (sorted by primary key)")

		// by unselected field
		e = di.Repo.FindAll(ctx, &models, Select("Value"), SortBy("SearchIdx", true))
		g.Expect(e).To(Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(HaveLen(len(modelIDs)), "sort by field shouldn't change model count")
		expected := fmt.Sprintf("Test %d", len(modelIDs)-1)
		g.Expect(models[0].Value).To(Equal(expected), "sorted result's first item should be correct even when sorted by field is not selected")

		// no model
		// note: we know that Save doesn't support SortBy
		e = di.Repo.Save(ctx, models[0], SortBy("Whatever", false))
		g.Expect(e).To(HaveOccurred(), "SortBy should return error when no model/schema is provided")

		// bad field
		models = nil
		e = di.Repo.FindAll(ctx, &models, SortBy("BadField.WhatEver", false))
		g.Expect(e).To(HaveOccurred(), "SortBy with invalid field name should return error")

		// bad pagination
		models = nil
		e = di.Repo.FindAll(ctx, &models, Page(0, 0))
		g.Expect(e).To(HaveOccurred(), "Page with 0 size should return error")
		e = di.Repo.FindAll(ctx, &models, Page(-1, 10))
		g.Expect(e).To(HaveOccurred(), "Page with negative page should return error")
		e = di.Repo.FindAll(ctx, &models, Page(int(^uint32(0))-20, 20))
		g.Expect(e).To(HaveOccurred(), "Page with too large offset should return error")
	}
}

func SubTestRepoSyntax(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		noop := func(db *gorm.DB) *gorm.DB {
			return db
		}
		var e error
		var model TestModel
		var opts []Option
		conditions := []Condition{noop}

		// slice of options
		model = TestModel{}
		opts = []Option{
			Joins("OneToOne"), Joins("ManyToOne"),
			Preload("ManyToOne.RelatedMTMModels"),
			noop,
		}
		e = di.Repo.FindById(ctx, &model, modelIDs[0].String(), opts)
		g.Expect(e).To(Succeed(), "FindById shouldn't return error")
		g.Expect(model.ID).To(BeEquivalentTo(modelIDs[0]), "FindById return correct result")
		assertFullyFetchedTestModel(&model, g, "FindById")

		// invalid options
		e = di.Repo.Create(ctx, &model, "unsupported option")
		g.Expect(e).To(HaveOccurred(), "CrudRepository should return error when option type is unsupported")

		// wrong model
		var wrong TestOTOModel
		var wrongSlice []*TestOTOModel
		e = di.Repo.FindById(ctx, &wrong, uuid.New())
		g.Expect(e).To(HaveOccurred(), "FindById should return error when wrong model is used")
		e = di.Repo.FindOneBy(ctx, &wrong, conditions)
		g.Expect(e).To(HaveOccurred(), "FindOneBy should return error when wrong model is used")
		e = di.Repo.FindAll(ctx, &wrongSlice)
		g.Expect(e).To(HaveOccurred(), "FindAll should return error when wrong model is used")
		e = di.Repo.FindAllBy(ctx, &wrongSlice, conditions)
		g.Expect(e).To(HaveOccurred(), "FindAllBy should return error when wrong model is used")
		e = di.Repo.Create(ctx, &wrong)
		g.Expect(e).To(HaveOccurred(), "Create should return error when wrong model is used")
		e = di.Repo.Save(ctx, &wrong)
		g.Expect(e).To(HaveOccurred(), "Save should return error when wrong model is used")
		e = di.Repo.Update(ctx, &wrong, map[string]interface{}{})
		g.Expect(e).To(HaveOccurred(), "Update should return error when wrong model is used")
		e = di.Repo.Delete(ctx, &wrong, conditions)
		g.Expect(e).To(HaveOccurred(), "Delete should return error when wrong model is used")

		// nil model
		e = di.Repo.FindById(ctx, nil, uuid.New())
		g.Expect(e).To(HaveOccurred(), "CrudRepository should return error when nil model is used")

		// truncate table
		e = di.Repo.Truncate(ctx)
		g.Expect(e).To(Succeed(), "Truncate shouldn't return error")
	}
}

func SubTestTransaction(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(BeTrue(), "repository should also GormApi")

		var model TestModel
		const newValue = "Updated"
		var e error
		// try single transaction with save/create
		id1 := uuid.New()
		id2 := uuid.New()
		e = tx.Transaction(ctx, func(ctx context.Context) (err error) {
			var e error
			// try create one
			model, oto := createMainModel(id1, 10)
			model.OneToOne = oto
			e = di.Repo.Save(ctx, model)
			g.Expect(e).To(Succeed(), "save in transaction shouldn't return error")
			// try create another with duplicate keys (will fail)
			another, _ := createMainModel(id2, 10)
			another.UniqueA = model.UniqueA
			another.UniqueB = model.UniqueB
			e = di.Repo.Create(ctx, another)
			g.Expect(e).To(HaveOccurred(), "save with duplicate key in transaction should return error")
			return e
		})

		e = di.Repo.FindById(ctx, &model, id1)
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(BeTrue(), "success save in transaction should be rolled back")

		// try multiple transaction
		mockedError := errors.New("just an error")
		// nested transaction
		e = tx.Transaction(ctx, func(ctx context.Context) error {
			var e error
			db := gormRepo.DB(ctx)
			g.Expect(db).To(Not(BeNil()), "DB(ctx) in transaction shouldn't be nil")

			// try update
			e = di.Repo.Update(ctx, &TestModel{ID: modelIDs[0]}, &TestModel{Value: newValue})
			g.Expect(e).To(Succeed(), "update within top transaction shouldn't fail")
			return gormRepo.Transaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
				g.Expect(tx).To(Not(BeNil()), "gorm.DB in transaction shouldn't be nil")
				// try update
				e = di.Repo.Update(ctx, &TestModel{ID: modelIDs[1]}, &TestModel{Value: newValue})
				g.Expect(e).To(Succeed(), "update within nested transaction shouldn't faile")
				return mockedError
			})
		})
		g.Expect(e).To(BeIdenticalTo(mockedError))

		// verify everything rolled back
		e = di.Repo.FindById(ctx, &model, modelIDs[0])
		g.Expect(e).To(Succeed(), "finding first model shouldn't fail")
		g.Expect(model.Value).ToNot(Equal(newValue), "first model shouldn't get updated")

		model = TestModel{}
		e = di.Repo.FindById(ctx, &model, modelIDs[1])
		g.Expect(e).To(Succeed(), "finding second model shouldn't fail")
		g.Expect(model.Value).ToNot(Equal(newValue), "second model shouldn't get updated")
	}
}

func SubTestUtilFunctions(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// MustApply...
		var db, rs *gorm.DB
		var model TestModel
		cond1 := Where("test_repo_models.search > 0")
		cond2 := clause.Where{Exprs: []clause.Expression{clause.Lt{
			Column: clause.Column{Table: clause.CurrentTable, Name: di.Repo.ColumnName("SearchIdx")},
			Value:  2,
		}}}
		opts := []Option{
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		}
		db = di.Repo.(GormApi).DB(ctx)
		db = MustApplyConditions(db, cond1, cond2)
		db = MustApplyOptions(db, opts)
		rs = db.Take(&model)
		g.Expect(rs.Error).To(Succeed(), "DB.Take() shouldn't return error")
		g.Expect(model.ID).To(BeEquivalentTo(modelIDs[1]), "DB.Take() return correct result")
		assertFullyFetchedTestModel(&model, g, "DB.Take()")

		// AsGormScope
		model = TestModel{}
		rs = di.Repo.(GormApi).DB(ctx).
			Scopes(AsGormScope(opts)).
			Scopes(AsGormScope(cond1)).
			Scopes(AsGormScope(cond2)).
			Take(&model)
		g.Expect(rs.Error).To(Succeed(), "DB.Take() shouldn't return error")
		g.Expect(model.ID).To(BeEquivalentTo(modelIDs[1]), "DB.Take() return correct result")
		assertFullyFetchedTestModel(&model, g, "DB.Take()")
	}
}

func SubTestResolveSchema(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var r SchemaResolver
		r, e = Utils().ResolveSchema(ctx, &TestOTOModel{})
		g.Expect(e).To(Succeed(), "ResolveSchema shouldn't return error")
		g.Expect(r.ModelName()).To(Equal("TestOTOModel"), "ResolveSchema have correct model name")

		r, e = Utils().ResolveSchema(ctx, map[string]interface{}{})
		g.Expect(e).To(HaveOccurred(), "ResolveSchema should return error on map value")
	}
}

func SubTestCheckUniqueness(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var model, toCheck TestModel
		var toChecks []*TestModel
		var toCheckMap map[string]interface{}
		var dups map[string]interface{}
		var e error
		// first, get an existing record
		e = di.Repo.FindById(ctx, &model, modelIDs[0])
		g.Expect(e).To(Succeed(), "database should have some data")

		// single check for no duplication
		toCheck = TestModel{ OneToOneKey: "whatever", UniqueA: model.UniqueA, UniqueB: "whatever" }
		dups, e = Utils().CheckUniqueness(ctx, &toCheck)
		g.Expect(e).To(Succeed(), "should not return error on single model check without duplicates")

		// single check for single key
		toCheck = TestModel{ OneToOneKey: model.OneToOneKey }
		dups, e = Utils().CheckUniqueness(ctx, &toCheck)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(BeTrue(), "should return error on single model check with duplicate simple keys")
		g.Expect(dups).To(HaveLen(1), "duplicates shouldn't be empty when uniqueness check fails")

		// single check for index key
		toCheck = TestModel{ UniqueA: model.UniqueA, UniqueB: model.UniqueB }
		dups, e = Utils().CheckUniqueness(ctx, &toCheck)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(BeTrue(), "should return error on single model check with duplicate composite keys")
		g.Expect(dups).To(HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// single check with field override fail
		toCheck = TestModel{ UniqueA: model.UniqueA, UniqueB: model.UniqueB, OneToOneKey: model.OneToOneKey}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck, []string{"UniqueA", "unique_b"})
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(BeTrue(), "should return error on single model check with duplicate composite keys and fields overrides")
		g.Expect(dups).To(HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// single check with field override succeed
		toCheck = TestModel{ UniqueA: model.UniqueA, UniqueB: model.UniqueB, OneToOneKey: "don't care"}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck, "OneToOneKey")
		g.Expect(e).To(Succeed(), "should not return error on single model check with duplicate single keys and fields overrides")

		// multi check failed
		toChecks = []*TestModel{
			{ UniqueA: "Not a issue", UniqueB: "shouldn't matter"},{UniqueA: model.UniqueA, UniqueB: model.UniqueB},
		}
		dups, e = Utils().CheckUniqueness(ctx, toChecks)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(BeTrue(), "should return error on multi models check with any model containing duplicate keys")
		g.Expect(dups).To(HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// multi check succeed
		toChecks = []*TestModel{
			{UniqueA: model.UniqueA, UniqueB: "not same"},{ UniqueA: "Not a issue", UniqueB: "shouldn't matter"},
		}
		dups, e = Utils().CheckUniqueness(ctx, toChecks)
		g.Expect(e).To(Succeed(), "should not return error on multi models check without any model containing duplicate keys")

		// map check failed
		toCheckMap = map[string]interface{}{"UniqueA": model.UniqueA, "unique_b": model.UniqueB}
		dups, e = Utils().Model(&TestModel{}).CheckUniqueness(ctx, toCheckMap)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(BeTrue(), "should return error on map check with duplicate keys")
		g.Expect(dups).To(HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// map check succeed
		toCheckMap = map[string]interface{}{"UniqueA": "doesn't matter", "unique_b": model.UniqueB}
		dups, e = Utils().Model(&TestModel{}).CheckUniqueness(ctx, toChecks)
		g.Expect(e).To(Succeed(), "should not return error on map check without duplicate keys")

		// invalid checks
		dups, e = Utils().CheckUniqueness(ctx, map[string]interface{}{})
		g.Expect(e).To(HaveOccurred(), "should return error of unsupported values")

		toCheck = TestModel{ OneToOneKey: "whatever", UniqueA: model.UniqueA, UniqueB: "whatever" }
		dups, e = Utils().CheckUniqueness(ctx, &toCheck, []string{"Invalid", "UniqueB"})
		g.Expect(e).To(HaveOccurred(), "should return error of invalid field/column name values")

		// all zero value
		dups, e = Utils().CheckUniqueness(ctx, &TestModel{})
		g.Expect(e).To(HaveOccurred(), "should return error of all zero values")

		// nil value
		dups, e = Utils().CheckUniqueness(ctx, nil)
		g.Expect(e).To(HaveOccurred(), "should return error of all zero values")

		// different options
		toCheck = TestModel{ OneToOneKey: "whatever", UniqueA: model.UniqueA, UniqueB: "whatever" }
		dups, e = Utils(&gorm.Session{NewDB: true}).CheckUniqueness(ctx, &toCheck)
		g.Expect(e).To(Succeed(), "should return error of all zero values")
	}
}

/*************************
	Helpers
 *************************/

func assertFullyFetchedTestModel(model *TestModel, g *gomega.WithT, target string) {
	g.Expect(model).To(Not(BeNil()), "assertFullyFetchedTestModel shouldn't get nil model")
	g.Expect(model.ID).To(Not(BeZero()), "%s should populate models", target)
	g.Expect(model.OneToOne).To(Not(BeNil()), "%s should populate OneToOne models", target)
	g.Expect(model.ManyToOne).To(Not(BeNil()), "%s should populate ManyToOne models", target)
	g.Expect(model.ManyToOne.RelatedMTMModels).To(Not(HaveLen(0)),
		"%s should populate ManyToOne.RelatedMTMModels models", target)
}

/*************************
	Setup Data
 *************************/

func prepareRepoTestData(db *gorm.DB, g *gomega.WithT) {
	var rs *gorm.DB
	// truncate table
	tables := []string{
		TestModel{}.TableName(), TestOTOModel{}.TableName(),
		"test_repo_relations",
		TestMTOModel{}.TableName(), TestMTMModel{}.TableName(),
	}
	for _, table := range tables {
		rs = db.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, table))
		g.Expect(rs.Error).To(Succeed(), "truncating table of %s should return no error", table)
	}

	// prepare models
	mtoModels := prepareMTOModels()
	mtmModels, relations := prepareMTMModels()
	models, otoModels := prepareMainModels()
	for _, list := range []interface{}{mtoModels, mtmModels, relations, otoModels, models} {
		rs = db.Create(list)
		g.Expect(rs.Error).To(Succeed(), "create models shouldn't fail")
		g.Expect(rs.RowsAffected).To(BeNumerically(">", 0), "create models should create correct number of rows")
	}
}

func prepareMTOModels() []*TestMTOModel {
	mtoModels := make([]*TestMTOModel, len(mtoModelIDs))
	for i, id := range mtoModelIDs {
		mtoModels[i] = &TestMTOModel{
			ID:             id,
			RelationValue:  fmt.Sprintf("MTO %d", i),
			RelationSearch: i,
		}
	}
	return mtoModels
}

func prepareMTMModels() ([]*TestMTMModel, []*TestMTMRelation) {
	mtmModels := make([]*TestMTMModel, len(mtmModelIDs))
	var relations []*TestMTMRelation
	for i, id := range mtmModelIDs {
		for j, mtoId := range mtoModelIDs {
			if j != i {
				relations = append(relations, &TestMTMRelation{
					MTOIModelD: mtoId,
					MTMModelID: id,
				})
			}
		}
		mtmModels[i] = &TestMTMModel{
			ID:        id,
			MTMValue:  fmt.Sprintf("MTM %d", i),
			MTMSearch: len(mtmModelIDs) - i - 1,
		}
	}
	return mtmModels, relations
}

func prepareMainModels() ([]*TestModel, []*TestOTOModel) {
	models := make([]*TestModel, len(modelIDs))
	otoModels := make([]*TestOTOModel, len(modelIDs))
	for i, id := range modelIDs {
		models[i], otoModels[i] = createMainModel(id, i)
	}
	return models, otoModels
}

func createMainModel(id uuid.UUID, i int) (*TestModel, *TestOTOModel) {
	refkey := utils.RandomString(8)
	oto := &TestOTOModel{
		RefKey:         refkey,
		RelationValue:  fmt.Sprintf("OTO %d", i),
		RelationSearch: len(modelIDs) - i - 1,
	}
	mtoId := mtoModelIDs[i%len(mtoModelIDs)]
	main := &TestModel{
		ID:          id,
		Value:       fmt.Sprintf("Test %d", i),
		SearchIdx:   i,
		UniqueA:     utils.RandomString(8),
		UniqueB:     utils.RandomString(8),
		OneToOneKey: refkey,
		ManyToOneID: mtoId,
	}
	return main, oto
}

/*************************
	Repository
 *************************/

type TestRepository CrudRepository

func NewTestRepository(factory Factory) TestRepository {
	return factory.NewCRUD(&TestModel{},
		&gorm.Session{CreateBatchSize: 20},
		gorm.Session{NewDB: true},
	)
}

/*************************
	Mocks
 *************************/

// relationship setup:
// MainModel <--> OTOModel
// MainModel >--> MTOModel
// MTOModel >--< MTMModel

const tableSQL1 = `
CREATE TABLE IF NOT EXISTS public.test_repo_model2 (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"value" STRING,
	search INT NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	INDEX idx_search (search ASC),
	FAMILY "primary" (id, "value")
);`

const tableSQL2 = `
CREATE TABLE IF NOT EXISTS public.test_repo_model1 (
	ref_key STRING NOT NULL,
	"value" STRING,
	search INT NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (ref_key ASC),
	INDEX idx_search (search ASC),
	FAMILY "primary" (ref_key, "value")
);`

const tableSQL3 = `
CREATE TABLE IF NOT EXISTS public.test_repo_model3 (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"value" STRING,
	search INT NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	INDEX idx_search (search ASC),
	FAMILY "primary" (id, "value")
);`

const tableSQL4 = `
CREATE TABLE IF NOT EXISTS public.test_repo_relations (
	mto_id UUID NULL,
	mtm_id UUID NULL,
	INDEX idx_mto_id (mto_id ASC),
	INDEX idx_mtm_id (mtm_id ASC),
	UNIQUE INDEX idx_pair (mto_id ASC, mtm_id ASC),
	FAMILY "primary" (mto_id, mtm_id, rowid)
);`

const tableSQLMain = `
CREATE TABLE IF NOT EXISTS public.test_repo_models (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"value" STRING,
	search INT NOT NULL,
	unique_a STRING,
	unique_b STRING,
	one_to_one_key STRING NOT NULL,
	many_to_one_id UUID NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	CONSTRAINT fk_one_to_one FOREIGN KEY (one_to_one_key) REFERENCES public.test_repo_model1(ref_key),
	CONSTRAINT fk_many_to_one FOREIGN KEY (many_to_one_id) REFERENCES public.test_repo_model2(id) ON DELETE SET NULL,
	UNIQUE INDEX idx_one_to_one (one_to_one_key ASC),
	UNIQUE INDEX idx_unique_composite (unique_a ASC, unique_b ASC),
	INDEX idx_many_to_one (many_to_one_id ASC),
	INDEX idx_search (search ASC),
	FAMILY "primary" (id, "value", one_to_one_key, many_to_one_id)
);`

func prepareTable(db *gorm.DB, g *gomega.WithT) {
	tableSQL := []string{
		tableSQL1, tableSQL2, tableSQL3, tableSQL4, tableSQLMain,
	}
	for _, q := range tableSQL {
		rs := db.Exec(q)
		g.Expect(rs.Error).To(Succeed(), "create table if not exists shouldn't fail")
	}
}

type TestModel struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value       string
	SearchIdx   int    `gorm:"column:search;"`
	UniqueA     string `gorm:"column:unique_a;uniqueIndex:idx_unique_composite;"`
	UniqueB     string `gorm:"column:unique_b;uniqueIndex:idx_unique_composite;"`
	OneToOneKey string `gorm:"uniqueIndex"`
	ManyToOneID uuid.UUID
	OneToOne    *TestOTOModel `gorm:"foreignKey:RefKey;references:OneToOneKey;not null"`
	ManyToOne   *TestMTOModel `gorm:"foreignKey:ManyToOneID;"`
}

func (TestModel) TableName() string {
	return "test_repo_models"
}

type TestOTOModel struct {
	RefKey         string `gorm:"primary_key;column:ref_key;type:TEXT;"`
	RelationValue  string `gorm:"column:value;"`
	RelationSearch int    `gorm:"column:search;"`
}

func (TestOTOModel) TableName() string {
	return "test_repo_model1"
}

type TestMTOModel struct {
	ID               uuid.UUID       `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	RelationValue    string          `gorm:"column:value;"`
	RelationSearch   int             `gorm:"column:search;"`
	RelatedMTMModels []*TestMTMModel `gorm:"many2many:test_repo_relations;joinForeignKey:mto_id;joinReferences:mtm_id"`
}

func (TestMTOModel) TableName() string {
	return "test_repo_model2"
}

type TestMTMModel struct {
	ID               uuid.UUID       `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	MTMValue         string          `gorm:"column:value;"`
	MTMSearch        int             `gorm:"column:search;"`
	RelatedMTOModels []*TestMTOModel `gorm:"many2many:test_repo_relations;joinForeignKey:mtm_id;joinReferences:mto_id"`
}

func (TestMTMModel) TableName() string {
	return "test_repo_model3"
}

type TestMTMRelation struct {
	MTOIModelD uuid.UUID `gorm:"column:mto_id;"`
	MTMModelID uuid.UUID `gorm:"column:mtm_id;"`
}

func (TestMTMRelation) TableName() string {
	return "test_repo_relations"
}
