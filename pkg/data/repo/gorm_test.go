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
	nonExistID     = uuid.MustParse("2e15c4d2-d427-4af2-b0d9-c3bcb0a8485c")
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

func TestCockroachTransactions(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(Module),
		apptest.WithTimeout(time.Minute),
		apptest.WithFxOptions(
			fx.Provide(NewTestRepository),
			fx.Supply(fx.Annotated{
				Target: tx.MaxRetries(1, 0),
				Group:  tx.FxTransactionExecuterOption,
			}),
		),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareTables(di)),
		test.GomegaSubTest(SubTestTransactionRetry(di), "TestTransactionRetry"),

		// The below 3 Transaction tests use nested transactions, which automatically use SavePoint with a name that is
		// equal to the memory address of the function so copyist is not happy with that. Comment them out when pushing to a branch
		//test.GomegaSubTest(SubTestTransaction(di), "TestTransaction"),
		//test.GomegaSubTest(SubTestNestedTransactionRetry(di), "SubTestNestedTransactionRetry"),
		//test.GomegaSubTest(SubTestNestedTransactionRetryExpectRollback(di), "SubTestNestedTransactionRetryExpectRollback"),

		test.GomegaSubTest(SubTestManualTransactionRetry(di), "SubTestManualTransactionRetry"),
		test.GomegaSubTest(SubTestNestedManualTransactionRollback(di), "SubTestNestedManualTransactionRollback"),
		test.GomegaSubTest(SubTestNestedManualTransaction(di), "SubTestNestedManualTransaction"),
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
			To(gomega.BeEquivalentTo(reflect.TypeOf(TestModel{})), "ModelType should be correct")
		g.Expect(di.Repo.Table()).
			To(gomega.Equal("test_repo_models"), "Table should be correct")
		g.Expect(di.Repo.ColumnName("Value")).
			To(gomega.Equal("value"), "ColumnName of direct field should be correct")
		g.Expect(di.Repo.ColumnDataType("Value")).
			To(gomega.Equal("string"), "ColumnDataType of direct field should be correct")

		g.Expect(di.Repo.ColumnName("Unknown")).
			To(gomega.Equal(""), "ColumnName of unknown field should be empty")
		g.Expect(di.Repo.ColumnDataType("Unknown")).
			To(gomega.Equal(""), "ColumnDataType of unknown field should be empty")

		g.Expect(di.Repo.(GormSchemaResolver).Schema()).
			To(gomega.Not(gomega.BeNil()), "Schema() should be correct")
	}
}

func SubTestSchemaResolverIndirect(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// one to one
		g.Expect(di.Repo.ColumnName("OneToOne.RelationValue")).
			To(gomega.Equal("value"), "ColumnName of one-to-one field should be correct with field path")
		g.Expect(di.Repo.ColumnDataType("OneToOne.RelationValue")).
			To(gomega.Equal("string"), "ColumnDataType of one-to-one field should be correct with field path")

		resolver := di.Repo.RelationshipSchema("OneToOne")
		g.Expect(resolver.ModelType()).
			To(gomega.BeEquivalentTo(reflect.TypeOf(TestOTOModel{})), "ModelType of one-to-one relation should be correct")
		g.Expect(resolver.Table()).
			To(gomega.Equal("test_repo_model1"), "Table of one-to-one relation should be correct")
		g.Expect(resolver).To(gomega.Not(gomega.BeNil()), "RelationshipSchema of one-to-one shouldn't be nil")
		g.Expect(resolver.ColumnName("RelationValue")).
			To(gomega.Equal("value"), "ColumnName of one-to-one model's schema should be correct")
		g.Expect(resolver.ColumnDataType("RelationValue")).
			To(gomega.Equal("string"), "ColumnDataType of one-to-one model's schema should be correct")

		// many to one
		g.Expect(di.Repo.ColumnName("ManyToOne.RelationValue")).
			To(gomega.Equal("value"), "ColumnName of many-to-one field should be correct with field path")
		g.Expect(di.Repo.ColumnDataType("ManyToOne.RelationValue")).
			To(gomega.Equal("string"), "ColumnDataType of many-to-one field should be correct with field path")

		resolver = di.Repo.RelationshipSchema("ManyToOne")
		g.Expect(resolver.ModelType()).
			To(gomega.BeEquivalentTo(reflect.TypeOf(TestMTOModel{})), "ModelType of many-to-one relation should be correct")
		g.Expect(resolver.Table()).
			To(gomega.Equal("test_repo_model2"), "Table of many-to-one relation should be correct")
		g.Expect(resolver).To(gomega.Not(gomega.BeNil()), "RelationshipSchema of many-to-one shouldn't be nil")
		g.Expect(resolver.ColumnName("RelationValue")).
			To(gomega.Equal("value"), "ColumnName of many-to-one model's schema should be correct")
		g.Expect(resolver.ColumnDataType("RelationValue")).
			To(gomega.Equal("string"), "ColumnDataType of many-to-one model's schema should be correct")
	}
}

func SubTestSchemaResolverMultiLvl(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Repo.ColumnName("ManyToOne.RelatedMTMModels.MTMValue")).
			To(gomega.Equal("value"), "ColumnName of many-to-many field should be correct with field path")
		g.Expect(di.Repo.ColumnDataType("ManyToOne.RelatedMTMModels.MTMValue")).
			To(gomega.Equal("string"), "ColumnDataType of many-to-many field should be correct with field path")

		resolver := di.Repo.RelationshipSchema("ManyToOne.RelatedMTMModels")
		g.Expect(resolver.ModelType()).
			To(gomega.BeEquivalentTo(reflect.TypeOf(TestMTMModel{})), "ModelType of many-to-many relation should be correct")
		g.Expect(resolver.Table()).
			To(gomega.Equal("test_repo_model3"), "Table of many-to-many relation should be correct")
		g.Expect(resolver).To(gomega.Not(gomega.BeNil()), "RelationshipSchema of many-to-many shouldn't be nil")
		g.Expect(resolver.ColumnName("MTMValue")).
			To(gomega.Equal("value"), "ColumnName of many-to-many model's schema should be correct")
		g.Expect(resolver.ColumnDataType("MTMValue")).
			To(gomega.Equal("string"), "ColumnDataType of many-to-many model's schema should be correct")

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
		g.Expect(e).To(gomega.Succeed(), "FindById shouldn't return error")
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
		g.Expect(e).To(gomega.Succeed(), "FindOneBy shouldn't return error")
		g.Expect(model.ID).To(gomega.BeEquivalentTo(modelIDs[1]), "FindOneBy return correct result")
		assertFullyFetchedTestModel(&model, g, "FindOneBy")

		model = TestModel{}
		conditions = []Condition{
			`"ManyToOne"."search" = 9999`,
		}
		e = di.Repo.FindOneBy(ctx, &model, conditions,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(gomega.BeTrue(), "FindOneBy should return RecordNotFound error")
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
		g.Expect(e).To(gomega.Succeed(), "FindAll shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(len(modelIDs)), "FindAll should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of FindAll")
		}

		models = nil
		e = di.Repo.FindAll(ctx, &models,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
		)
		g.Expect(e).To(gomega.Succeed(), "FindAll shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(len(modelIDs)), "FindAll should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of FindAll")
		}

		// find all paginated
		models = nil
		e = di.Repo.FindAll(ctx, &models,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
			Page(1, 5),
		)
		g.Expect(e).To(gomega.Succeed(), "paginated FindAll shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(4), "paginated FindAll should returns correct number of records")
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
		g.Expect(e).To(gomega.Succeed(), "paginated FindAllBy shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(7), "paginated FindAllBy should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of paginated FindAllBy")
		}

		// find all by, paginated
		models = nil
		e = di.Repo.FindAllBy(ctx, &models, conditions,
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"),
			Page(1, 5),
		)
		g.Expect(e).To(gomega.Succeed(), "FindAllBy shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(2), "FindAllBy should returns correct number of records")
		for _, m := range models {
			assertFullyFetchedTestModel(m, g, "for each model of FindAllBy")
		}

		// find all by, using OR
		models = nil
		e = di.Repo.FindAllBy(ctx, &models, Or("test_repo_models.value = ?", "Test 0"), Or("test_repo_models.value = ?", "Test 1"),
			Joins("OneToOne"), Joins("ManyToOne"), Preload("ManyToOne.RelatedMTMModels"))
		g.Expect(e).To(gomega.Succeed(), "FindAllBy shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(2), "FindAllBy should returns correct number of records")
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
		g.Expect(e).To(gomega.Succeed(), "CountAll shouldn't return error")
		g.Expect(count).To(gomega.BeEquivalentTo(len(modelIDs)), "CountAll should return correct result")

		conditions := []Condition{
			`"ManyToOne".search > 0`,
		}
		count, e = di.Repo.CountBy(ctx, conditions, Joins("ManyToOne"))
		g.Expect(e).To(gomega.Succeed(), "CountBy shouldn't return error")
		g.Expect(count).To(gomega.BeEquivalentTo(6), "CountBy should return correct result")
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
		g.Expect(rs.Error).To(gomega.Succeed(), "Create OTO models shouldn't return error")

		e = di.Repo.Create(ctx, models, Omit("OneToOne"), Omit("ManyToOne"))
		g.Expect(e).To(gomega.Succeed(), "Create shouldn't return error")
	}
}

func SubTestSave(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var models []*TestModel
		e = di.Repo.FindAll(ctx, &models)
		g.Expect(e).To(gomega.Succeed(), "FindAll shouldn't return error")
		g.Expect(models).To(gomega.Not(gomega.BeEmpty()), "FindAll should return some records")

		// do save
		for _, m := range models {
			m.Value = "Updated Value"
		}
		e = di.Repo.Save(ctx, models)
		g.Expect(e).To(gomega.Succeed(), "Save shouldn't return error")

		// fetch again and validate
		models = nil
		e = di.Repo.FindAll(ctx, &models)
		g.Expect(e).To(gomega.Succeed(), "FindAll shouldn't return error")
		for _, m := range models {
			g.Expect(m.Value).To(gomega.BeEquivalentTo("Updated Value"), "saved values should be correct when re-fetch")
		}
	}
}

func SubTestUpdates(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		id := modelIDs[0]
		e = di.Repo.Update(ctx, &TestModel{ID: id}, &TestModel{Value: "Just Updated"}, ErrorOnZeroRows())
		g.Expect(e).To(gomega.Succeed(), "Update shouldn't return error")

		var model TestModel
		e = di.Repo.FindById(ctx, &model, id)
		g.Expect(e).To(gomega.Succeed(), "FindById shouldn't return error")
		g.Expect(model.Value).To(gomega.BeEquivalentTo("Just Updated"), "re-fetched value after Update should be correct")

		// update 0 rows by id without ErrorOnZeroRows option
		e = di.Repo.Update(ctx, &TestModel{ID: nonExistID}, &TestModel{Value: "Just Updated"})
		g.Expect(e).To(gomega.Succeed(), "Delete shouldn't return error with 0 affected rows")

		// delete 0 rows by with ErrorOnZeroRows option
		e = di.Repo.Update(ctx, &TestModel{ID: nonExistID}, &TestModel{Value: "Just Updated"}, ErrorOnZeroRows())
		g.Expect(e).To(gomega.HaveOccurred(), "Delete should return error with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(gomega.BeTrue(), "Delete should return gorm.ErrRecordNotFound with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, data.ErrorRecordNotFound)).To(gomega.BeTrue(), "Delete should return data.ErrorRecordNotFound with 0 affected rows and ErrorOnZeroRows")
	}
}

func SubTestDelete(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error

		// delete by id
		id := modelIDs[0]
		e = di.Repo.Delete(ctx, &TestModel{ID: id})
		g.Expect(e).To(gomega.Succeed(), "Delete shouldn't return error")

		var model TestModel
		e = di.Repo.FindById(ctx, &model, id, ErrorOnZeroRows())
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(gomega.BeTrue(), "re-fetch after Delete should yield RecordNotFound")

		//consecutive delete by with returning clause
		returningOpts := func(db *gorm.DB) *gorm.DB {
			return db.Clauses(clause.Returning{})
		}
		e = di.Repo.DeleteBy(ctx, Where(`"test_repo_models"."search" = ?`, 1), ErrorOnZeroRows(), returningOpts)
		g.Expect(e).To(gomega.Succeed(), "DeleteBy shouldn't return error")
		e = di.Repo.DeleteBy(ctx, Where(`"test_repo_models"."search" = ?`, 2), ErrorOnZeroRows(), returningOpts)
		g.Expect(e).To(gomega.Succeed(), "DeleteBy shouldn't return error")

		//delete by
		e = di.Repo.DeleteBy(ctx, Where(`"test_repo_models"."search" < ?`, len(modelIDs)-1))
		g.Expect(e).To(gomega.Succeed(), "DeleteBy shouldn't return error")

		// delete 0 rows by id without ErrorOnZeroRows option
		e = di.Repo.Delete(ctx, &TestModel{ID: nonExistID})
		g.Expect(e).To(gomega.Succeed(), "Delete shouldn't return error with 0 affected rows")

		// delete 0 rows by with ErrorOnZeroRows option
		e = di.Repo.Delete(ctx, &TestModel{ID: nonExistID}, ErrorOnZeroRows())
		g.Expect(e).To(gomega.HaveOccurred(), "Delete should return error with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(gomega.BeTrue(), "Delete should return gorm.ErrRecordNotFound with 0 affected rows and ErrorOnZeroRows")
		g.Expect(errors.Is(e, data.ErrorRecordNotFound)).To(gomega.BeTrue(), "Delete should return data.ErrorRecordNotFound with 0 affected rows and ErrorOnZeroRows")

		// fetch again and validate
		var count int
		count, e = di.Repo.CountAll(ctx)
		g.Expect(e).To(gomega.Succeed(), "CountAll after Delete shouldn't return error")
		g.Expect(count).To(gomega.BeEquivalentTo(1), "Delete should result in 1 record left")
	}
}

func SubTestPageAndSort(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var models []*TestModel
		// by direct field
		e = di.Repo.FindAll(ctx, &models, SortBy("SearchIdx", true))
		g.Expect(e).To(gomega.Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(len(modelIDs)), "sort by field shouldn't change model count")
		g.Expect(models[0].SearchIdx).To(gomega.BeEquivalentTo(len(modelIDs)-1), "sorted result's first item should be correct")

		// by to-one relation field
		models = nil
		e = di.Repo.FindAll(ctx, &models, Joins("ManyToOne"), SortBy("ManyToOne.RelationSearch", true))
		g.Expect(e).To(gomega.Succeed(), "sort by relation's field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(len(modelIDs)), "sort by relation's field shouldn't change model count")
		g.Expect(models[0].ManyToOne.RelationSearch).To(gomega.BeEquivalentTo(len(mtoModelIDs)-1), "sorted result's first item should be correct")

		// by column
		e = di.Repo.FindAll(ctx, &models, Sort("search DESC"))
		g.Expect(e).To(gomega.Succeed(), "sort by column shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(len(modelIDs)), "sort by column shouldn't change model count")
		g.Expect(models[0].SearchIdx).To(gomega.BeEquivalentTo(len(modelIDs)-1), "sorted result's first item should be correct")

		// page + sort
		models = nil
		e = di.Repo.FindAll(ctx, &models, Page(1, 2), SortBy("SearchIdx", false))
		g.Expect(e).To(gomega.Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].SearchIdx).To(gomega.BeEquivalentTo(2), "sorted page result should have correct order")

		// sort + page (reversed order)
		models = nil
		e = di.Repo.FindAll(ctx, &models, SortBy("SearchIdx", false), Page(1, 2))
		g.Expect(e).To(gomega.Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].SearchIdx).To(gomega.BeEquivalentTo(2), "sorted page result should have correct order")

		// page + sort as single option
		models = nil
		e = di.Repo.FindAll(ctx, &models, []Option{Page(1, 2), SortBy("SearchIdx", false)})
		g.Expect(e).To(gomega.Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].SearchIdx).To(gomega.BeEquivalentTo(2), "sorted page result should have correct order")

		// sort + page (reversed order) as single option
		models = nil
		e = di.Repo.FindAll(ctx, &models, []Option{SortBy("SearchIdx", false), Page(1, 2)})
		g.Expect(e).To(gomega.Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].SearchIdx).To(gomega.BeEquivalentTo(2), "sorted page result should have correct order")

		// page only
		models = nil
		e = di.Repo.FindAll(ctx, &models, Page(1, 2))
		g.Expect(e).To(gomega.Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(2), "paged result should have correct count")
		g.Expect(models[0].ID).To(gomega.BeEquivalentTo(sortedModelIDs[2]), "page result should have correct order (sorted by primary key)")

		// by unselected field
		e = di.Repo.FindAll(ctx, &models, Select("Value"), SortBy("SearchIdx", true))
		g.Expect(e).To(gomega.Succeed(), "sort by field shouldn't return error")
		g.Expect(models).To(gomega.HaveLen(len(modelIDs)), "sort by field shouldn't change model count")
		expected := fmt.Sprintf("Test %d", len(modelIDs)-1)
		g.Expect(models[0].Value).To(gomega.Equal(expected), "sorted result's first item should be correct even when sorted by field is not selected")

		// no model
		// note: we know that Save doesn't support SortBy
		e = di.Repo.Save(ctx, models[0], SortBy("Whatever", false))
		g.Expect(e).To(gomega.HaveOccurred(), "SortBy should return error when no model/schema is provided")

		// bad field
		models = nil
		e = di.Repo.FindAll(ctx, &models, SortBy("BadField.WhatEver", false))
		g.Expect(e).To(gomega.HaveOccurred(), "SortBy with invalid field name should return error")

		// bad pagination
		models = nil
		e = di.Repo.FindAll(ctx, &models, Page(0, 0))
		g.Expect(e).To(gomega.HaveOccurred(), "Page with 0 size should return error")
		e = di.Repo.FindAll(ctx, &models, Page(-1, 10))
		g.Expect(e).To(gomega.HaveOccurred(), "Page with negative page should return error")
		e = di.Repo.FindAll(ctx, &models, Page(int(^uint32(0))-20, 20))
		g.Expect(e).To(gomega.HaveOccurred(), "Page with too large offset should return error")
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
		g.Expect(e).To(gomega.Succeed(), "FindById shouldn't return error")
		g.Expect(model.ID).To(gomega.BeEquivalentTo(modelIDs[0]), "FindById return correct result")
		assertFullyFetchedTestModel(&model, g, "FindById")

		// invalid options
		e = di.Repo.Create(ctx, &model, "unsupported option")
		g.Expect(e).To(gomega.HaveOccurred(), "CrudRepository should return error when option type is unsupported")

		// wrong model
		var wrong TestOTOModel
		var wrongSlice []*TestOTOModel
		e = di.Repo.FindById(ctx, &wrong, uuid.New())
		g.Expect(e).To(gomega.HaveOccurred(), "FindById should return error when wrong model is used")
		e = di.Repo.FindOneBy(ctx, &wrong, conditions)
		g.Expect(e).To(gomega.HaveOccurred(), "FindOneBy should return error when wrong model is used")
		e = di.Repo.FindAll(ctx, &wrongSlice)
		g.Expect(e).To(gomega.HaveOccurred(), "FindAll should return error when wrong model is used")
		e = di.Repo.FindAllBy(ctx, &wrongSlice, conditions)
		g.Expect(e).To(gomega.HaveOccurred(), "FindAllBy should return error when wrong model is used")
		e = di.Repo.Create(ctx, &wrong)
		g.Expect(e).To(gomega.HaveOccurred(), "Create should return error when wrong model is used")
		e = di.Repo.Save(ctx, &wrong)
		g.Expect(e).To(gomega.HaveOccurred(), "Save should return error when wrong model is used")
		e = di.Repo.Update(ctx, &wrong, map[string]interface{}{})
		g.Expect(e).To(gomega.HaveOccurred(), "Update should return error when wrong model is used")
		e = di.Repo.Delete(ctx, &wrong, conditions)
		g.Expect(e).To(gomega.HaveOccurred(), "Delete should return error when wrong model is used")

		// nil model
		e = di.Repo.FindById(ctx, nil, uuid.New())
		g.Expect(e).To(gomega.HaveOccurred(), "CrudRepository should return error when nil model is used")

		// truncate table
		e = di.Repo.Truncate(ctx)
		g.Expect(e).To(gomega.Succeed(), "Truncate shouldn't return error")
	}
}

func SubTestTransaction(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")

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
			g.Expect(e).To(gomega.Succeed(), "save in transaction shouldn't return error")
			// try create another with duplicate keys (will fail)
			another, _ := createMainModel(id2, 10)
			another.UniqueA = model.UniqueA
			another.UniqueB = model.UniqueB
			e = di.Repo.Create(ctx, another)
			g.Expect(e).To(gomega.HaveOccurred(), "save with duplicate key in transaction should return error")
			return e
		})

		e = di.Repo.FindById(ctx, &model, id1)
		g.Expect(errors.Is(e, gorm.ErrRecordNotFound)).To(gomega.BeTrue(), "success save in transaction should be rolled back")

		// try multiple transaction
		mockedError := errors.New("just an error")
		// nested transaction
		e = tx.Transaction(ctx, func(ctx context.Context) error {
			var e error
			db := gormRepo.DB(ctx)
			g.Expect(db).To(gomega.Not(gomega.BeNil()), "DB(ctx) in transaction shouldn't be nil")

			// try update
			e = di.Repo.Update(ctx, &TestModel{ID: modelIDs[0]}, &TestModel{Value: newValue})
			g.Expect(e).To(gomega.Succeed(), "update within top transaction shouldn't fail")
			return gormRepo.Transaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
				g.Expect(tx).To(gomega.Not(gomega.BeNil()), "gorm.DB in transaction shouldn't be nil")
				// try update
				e = di.Repo.Update(ctx, &TestModel{ID: modelIDs[1]}, &TestModel{Value: newValue})
				g.Expect(e).To(gomega.Succeed(), "update within nested transaction shouldn't faile")
				return mockedError
			})
		})
		g.Expect(e).To(gomega.BeIdenticalTo(mockedError))

		// verify everything rolled back
		e = di.Repo.FindById(ctx, &model, modelIDs[0])
		g.Expect(e).To(gomega.Succeed(), "finding first model shouldn't fail")
		g.Expect(model.Value).ToNot(gomega.Equal(newValue), "first model shouldn't get updated")

		model = TestModel{}
		e = di.Repo.FindById(ctx, &model, modelIDs[1])
		g.Expect(e).To(gomega.Succeed(), "finding second model shouldn't fail")
		g.Expect(model.Value).ToNot(gomega.Equal(newValue), "second model shouldn't get updated")
	}
}

func SubTestTransactionRetry(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// Setup the test case to be similar to https://www.cockroachlabs.com/docs/v22.2/demo-serializable
		// 1) We will create two models to setup the scenario.
		// 2) Create two concurrent transactions where each transaction is in charge of modifying one of the models
		// 3) In each transaction, they will perform a search that includes the other model.
		// 4) After the search, the transactions will modify the search field of their model.
		// If the retry works automatically, then the model 1 transaction will re-run and should successfully
		// change model1's search value to the updatedSearchValue
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())

		var testModels []*TestModel
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")

		// This Transaction2 happens during the execution of Transaction1. This will
		// cause a serializable isolation violation by modifying an entry that shows up
		// in the Transaction1's query.
		model2Transaction := func(ctx context.Context, tx2 *gorm.DB) error {
			err := tx2.Where("search = ?", model2.SearchIdx).Find(&testModels).Error
			g.Expect(err).To(gomega.Succeed())
			model2.SearchIdx = updatedSearchValue
			err = tx2.Updates(model2).Error
			g.Expect(err).To(gomega.Succeed())
			return nil
		}

		// Transaction1 will execute transaction2 not as a nested transaction, but as a separate
		// transaction. This will mimic two transactions happening simultaneously where transaction 2
		// completes first.
		// We expect this Transaction1 to complete more than once (depends on number of retries), so
		// if the transaction has already run once, we do not want to trigger Transaction2 again.
		var executedModel1Transaction bool
		model1Transaction := func(ctx1 context.Context, tx1 *gorm.DB) error {
			err := tx1.Where("search = ?", model1.SearchIdx).Find(&testModels).Error
			g.Expect(err).To(gomega.Succeed())

			if !executedModel1Transaction {
				// using ctx and not ctx1 to make sure we're doing a concurrent transaction
				// and not a nested transaction
				err = gormRepo.Transaction(ctx, model2Transaction)
				g.Expect(err).To(gomega.Succeed())
			}
			model1.SearchIdx = updatedSearchValue
			err = tx1.Updates(model1).Error
			g.Expect(err).To(gomega.Succeed())
			executedModel1Transaction = true
			return nil
		}

		// We trigger Transaction1, which will then trigger Transaction2
		err = gormRepo.Transaction(ctx, model1Transaction)

		if tx.ErrIsRetryable(err) {
			t.Fatalf("the transaction was not retried: %v", err)
		}
		g.Expect(err).To(gomega.Succeed())

		// Validation
		err = gormRepo.DB(ctx).Where("search = ?", originalSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(0))
		err = gormRepo.DB(ctx).Where("search = ?", updatedSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(2))

	}
}

// SubTestNestedTransactionRetry will be the same as the non nested one, however we'll wrap the model 1 transaction
// around another transaction. This will test to see if an inner transaction is able to be retried.
func SubTestNestedTransactionRetry(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())

		var testModels []*TestModel
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")

		model2Transaction := func(ctx2 context.Context, tx2 *gorm.DB) error {
			err := tx2.Where("search = ?", model2.SearchIdx).Find(&testModels).Error
			g.Expect(err).To(gomega.Succeed())
			model2.SearchIdx = updatedSearchValue
			err = tx2.Updates(model2).Error
			g.Expect(err).To(gomega.Succeed())
			return nil
		}

		var executedModel1Transaction bool
		model1Transaction := func(ctx1 context.Context, tx1 *gorm.DB) error {
			err := tx1.Where("search = ?", model1.SearchIdx).Find(&testModels).Error
			g.Expect(err).To(gomega.Succeed())

			if !executedModel1Transaction {
				// we use ctx1 to nest the transaction
				err = gormRepo.Transaction(ctx1, model2Transaction)
				g.Expect(err).To(gomega.Succeed())
			}
			model1.SearchIdx = updatedSearchValue
			err = tx1.Updates(model1).Error
			g.Expect(err).To(gomega.Succeed())
			executedModel1Transaction = true
			return nil
		}

		var model1TransactionErr error
		outerTransaction := func(ctxOuter context.Context, txOuter *gorm.DB) error {
			model1TransactionErr = gormRepo.Transaction(ctxOuter, model1Transaction)
			return model1TransactionErr
		}
		// We trigger Transaction1, which will then trigger Transaction2
		err = gormRepo.Transaction(ctx, outerTransaction)

		if tx.ErrIsRetryable(err) {
			t.Fatalf("the transaction was not retried: %v", err)
		}
		g.Expect(err).To(gomega.Succeed())

		// Validation
		err = gormRepo.DB(ctx).Where("search = ?", originalSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(0))
		err = gormRepo.DB(ctx).Where("search = ?", updatedSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(2))
	}
}

// SubTestNestedTransactionRetryExpectRollback will be the same as the retry, however instead
// of succeeding the retry, we will rollback the outermost transaction.
// This is a sanity check to make sure that the transactions are actually nested, and a rollback
// on the outermost transaction will result in all the inner transactions being void.
func SubTestNestedTransactionRetryExpectRollback(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())

		var testModels []*TestModel
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")

		model2Transaction := func(ctx2 context.Context) error {
			err := di.Repo.FindAllBy(ctx2, &testModels, Where("search = ?", model2.SearchIdx))
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.Update(ctx2, model2, TestModel{SearchIdx: updatedSearchValue})
			g.Expect(err).To(gomega.Succeed())
			return nil
		}

		var executedModel1Transaction bool
		model1Transaction := func(ctx1 context.Context) error {
			err := di.Repo.FindAllBy(ctx1, &testModels, Where("search = ?", model1.SearchIdx))
			g.Expect(err).To(gomega.Succeed())

			if !executedModel1Transaction {
				// we use ctx1 to nest the transaction
				err = tx.Transaction(ctx1, model2Transaction)
				g.Expect(err).To(gomega.Succeed())
			}
			err = di.Repo.Update(ctx1, model1, TestModel{SearchIdx: updatedSearchValue})
			g.Expect(err).To(gomega.Succeed())
			executedModel1Transaction = true
			return nil
		}

		//var model1TransactionErr error
		outerTransaction := func(ctxOuter context.Context) error {
			_ = tx.Transaction(ctxOuter, model1Transaction)
			return errors.New("Hello World, this should rollback transaction")
			//return model1TransactionErr
		}
		// We trigger Transaction1, which will then trigger Transaction2
		err = tx.Transaction(ctx, outerTransaction)

		if tx.ErrIsRetryable(err) {
			t.Fatalf("the transaction was not retried: %v", err)
		}
		//g.Expect(err).To(gomega.Succeed())

		// Validation
		err = gormRepo.DB(ctx).Where("search = ?", originalSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(2))
		err = gormRepo.DB(ctx).Where("search = ?", updatedSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(0))
	}
}

// SubTestManualTransactionRetry will test that a retry happens when two concurrent transactions cause
// a serializable error due to contention.
func SubTestManualTransactionRetry(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())
		var testModels []TestModel

		transaction2 := func(tx2Ctx context.Context) {
			tx2Ctx, err := tx.Begin(tx2Ctx)
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.FindAllBy(tx2Ctx, &testModels, Where("search = ? ", model2.SearchIdx))
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.Update(tx2Ctx, model2, TestModel{SearchIdx: updatedSearchValue})
			g.Expect(err).To(gomega.Succeed())
			tx2Ctx, err = tx.Commit(tx2Ctx)
		}
		maxRetries := 1
		retryCount := 0
		for {
			tx1Ctx, err := tx.Begin(ctx)
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.FindAllBy(tx1Ctx, &testModels, Where("search = ? ", model1.SearchIdx))
			g.Expect(err).To(gomega.Succeed())
			if retryCount == 0 {
				// A non nested transaction. This is a concurrent transaction that will modify
				// the searchIdx
				transaction2(ctx)
			}
			err = di.Repo.Update(tx1Ctx, model1, TestModel{SearchIdx: updatedSearchValue})
			g.Expect(err).To(gomega.Succeed())
			_, err = tx.Commit(tx1Ctx) // should have rolled back automatically if it failed
			if err == nil || !tx.ErrIsRetryable(err) {
				break
			}
			retryCount++
			if maxRetries > retryCount {
				break
			}
		}
		// Validation
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")
		err = gormRepo.DB(ctx).Where("search = ?", originalSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(0))
		err = gormRepo.DB(ctx).Where("search = ?", updatedSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(2))
	}
}

func SubTestNestedManualTransaction(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())
		var testModels []TestModel

		tx1Ctx, err := tx.Begin(ctx)
		g.Expect(err).To(gomega.Succeed())
		// Optional transaction where we update model1, we are going to fail it and rollback
		{
			tx2Ctx, err := tx.SavePoint(tx1Ctx, "item1")
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.FindAllBy(tx2Ctx, &testModels, Where("search = ? ", model1.SearchIdx))
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.Update(tx2Ctx, model1, TestModel{SearchIdx: updatedSearchValue})
			g.Expect(err).To(gomega.Succeed())
			tx2Ctx, err = tx.RollbackTo(tx2Ctx, "item1")
		}
		// Transaction where we update model 2
		{
			tx2Ctx, err := tx.SavePoint(tx1Ctx, "item2")
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.FindAllBy(tx2Ctx, &testModels, Where("search = ? ", model2.SearchIdx))
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.Update(tx2Ctx, model2, TestModel{SearchIdx: updatedSearchValue})
			g.Expect(err).To(gomega.Succeed())
		}
		tx1Ctx, err = tx.Commit(tx1Ctx) // should have rolled back automatically if it failed
		g.Expect(err).To(gomega.Succeed())
		// Validation
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")
		err = gormRepo.DB(ctx).Where("search = ?", originalSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(1))
		err = gormRepo.DB(ctx).Where("search = ?", updatedSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(1))
	}
}

// SubTestNestedManualTransactionRollback tests to ensure that the nested transaction can be rolled back if the
// outer transaction is also rolled back
func SubTestNestedManualTransactionRollback(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())
		var testModels []TestModel

		transaction2 := func(tx1Ctx context.Context) {
			tx2Ctx, err := tx.SavePoint(tx1Ctx, "savepoint")
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.FindAllBy(tx2Ctx, &testModels, Where("search = ? ", model2.SearchIdx))
			g.Expect(err).To(gomega.Succeed())
			err = di.Repo.Update(tx2Ctx, model2, TestModel{SearchIdx: updatedSearchValue})
			g.Expect(err).To(gomega.Succeed())
		}
		tx1Ctx, err := tx.Begin(ctx)
		g.Expect(err).To(gomega.Succeed())
		err = di.Repo.FindAllBy(tx1Ctx, &testModels, Where("search = ? ", model1.SearchIdx))
		g.Expect(err).To(gomega.Succeed())
		transaction2(tx1Ctx)
		err = di.Repo.Update(tx1Ctx, model1, TestModel{SearchIdx: updatedSearchValue})
		g.Expect(err).To(gomega.Succeed())
		tx1Ctx, err = tx.Rollback(tx1Ctx)
		// Validation
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")
		err = gormRepo.DB(ctx).Where("search = ?", originalSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(2))
		err = gormRepo.DB(ctx).Where("search = ?", updatedSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(0))
	}
}

func SubTestManualTransactionRetryExpectRollbackVanillaGorm(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())
		var testModels []TestModel
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")

		transaction2 := func(db *gorm.DB) {
			innerTx := db.Begin()
			err = innerTx.Error
			g.Expect(err).To(gomega.Succeed())
			err = innerTx.Find(&testModels, innerTx.Where("search = ? ", model2.SearchIdx)).Error
			g.Expect(err).To(gomega.Succeed())
			model2.SearchIdx = updatedSearchValue
			innerTx.Updates(model2)
			err = innerTx.Commit().Error
			g.Expect(err).To(gomega.Succeed())
		}
		maxRetries := 1
		retryCount := 0
		for {
			innerTx := gormRepo.DB(ctx).Begin()
			err = innerTx.Error
			g.Expect(err).To(gomega.Succeed())
			err = innerTx.Find(&testModels, innerTx.Where("search = ? ", model2.SearchIdx)).Error
			g.Expect(err).To(gomega.Succeed())
			if retryCount == 0 {
				transaction2(innerTx)
			}
			model1.SearchIdx = updatedSearchValue
			innerTx.Updates(model1)
			g.Expect(err).To(gomega.Succeed())
			if retryCount == 0 {
				err = innerTx.Commit().Error
			} else {
				err = innerTx.Rollback().Error
			}
			if err == nil || !tx.ErrIsRetryable(err) {
				break
			}
			retryCount++
			if maxRetries > retryCount {
				break
			}
		}
		// Validation
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")
		err = gormRepo.DB(ctx).Where("search = ?", originalSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(2))
		err = gormRepo.DB(ctx).Where("search = ?", updatedSearchValue).Find(&testModels).Error
		g.Expect(err).To(gomega.Succeed())
		g.Expect(len(testModels)).To(gomega.Equal(0))
	}
}

// SubTestNestedTransaction, will verify that things are happening through a nested
// transaction by nesting one transaction, and then rolling it back
func SubTestNestedTransaction(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		originalSearchValue, updatedSearchValue := 20, 21
		ID1, ID2 := uuid.New(), uuid.New()
		model1, oto := createMainModel(ID1, originalSearchValue)
		model1.OneToOne = oto
		err := di.Repo.Save(ctx, model1)
		g.Expect(err).To(gomega.Succeed())
		model2, oto := createMainModel(ID2, originalSearchValue)
		model2.OneToOne = oto
		err = di.Repo.Save(ctx, model2)
		g.Expect(err).To(gomega.Succeed())

		var throwayModels []*TestModel
		gormRepo, ok := di.Repo.(GormApi)
		g.Expect(ok).To(gomega.BeTrue(), "repository should also GormApi")

		// This Transaction2 happens during the execution of Transaction1. This will
		// cause a serializable isolation violation by modifying an entry that shows up
		// in the Transaction1's query.
		model2Transaction := func(ctx context.Context) error {
			err = di.Repo.FindAll(ctx, &throwayModels, Where("search = ?", model2.SearchIdx))
			g.Expect(err).To(gomega.Succeed())
			model2.SearchIdx = updatedSearchValue
			err = di.Repo.Update(ctx, model2, model2)
			g.Expect(err).To(gomega.Succeed())
			return ErrorInvalidCrudParam
		}

		var executedModel1Transaction bool
		model1Transaction := func(ctx context.Context, tx1 *gorm.DB) error {
			if !executedModel1Transaction {
				err = tx.Transaction(ctx, model2Transaction)
				//err = gormRepo.Transaction(ctx, model2Transaction)
				//g.Expect(err).To(gomega.Succeed())
			}
			model1.SearchIdx = updatedSearchValue
			err = tx1.Updates(model1).Error
			g.Expect(err).To(gomega.Succeed())
			executedModel1Transaction = true
			return nil
		}

		// We trigger Transaction1, which will then trigger Transaction2
		err = gormRepo.Transaction(ctx, model1Transaction)

		// errWithSQLState is implemented by pgx (pgconn.PgError) and lib/pq
		type errWithSQLState interface {
			SQLState() string
		}
		var sqlErr errWithSQLState
		if err != nil && errors.As(err, &sqlErr) {
			code := sqlErr.SQLState()
			if code == "CR000" || code == "40001" {
				// If the transaction was retried, it would re-run and pass.
				t.Fatalf("the transaction was not retried: %v", err)
			}
		}
		g.Expect(err).To(gomega.Succeed())

		// TODO: add validation
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
		g.Expect(rs.Error).To(gomega.Succeed(), "DB.Take() shouldn't return error")
		g.Expect(model.ID).To(gomega.BeEquivalentTo(modelIDs[1]), "DB.Take() return correct result")
		assertFullyFetchedTestModel(&model, g, "DB.Take()")

		// AsGormScope
		model = TestModel{}
		rs = di.Repo.(GormApi).DB(ctx).
			Scopes(AsGormScope(opts)).
			Scopes(AsGormScope(cond1)).
			Scopes(AsGormScope(cond2)).
			Take(&model)
		g.Expect(rs.Error).To(gomega.Succeed(), "DB.Take() shouldn't return error")
		g.Expect(model.ID).To(gomega.BeEquivalentTo(modelIDs[1]), "DB.Take() return correct result")
		assertFullyFetchedTestModel(&model, g, "DB.Take()")
	}
}

func SubTestResolveSchema(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var r SchemaResolver
		r, e = Utils().ResolveSchema(ctx, &TestOTOModel{})
		g.Expect(e).To(gomega.Succeed(), "ResolveSchema shouldn't return error")
		g.Expect(r.ModelName()).To(gomega.Equal("TestOTOModel"), "ResolveSchema have correct model name")

		r, e = Utils().ResolveSchema(ctx, map[string]interface{}{})
		g.Expect(e).To(gomega.HaveOccurred(), "ResolveSchema should return error on map value")
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
		g.Expect(e).To(gomega.Succeed(), "database should have some data")

		// single check for no duplication
		toCheck = TestModel{OneToOneKey: "whatever", UniqueA: model.UniqueA, UniqueB: "whatever"}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck)
		g.Expect(e).To(gomega.Succeed(), "should not return error on single model check without duplicates")

		// single check for single key
		toCheck = TestModel{OneToOneKey: model.OneToOneKey}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(gomega.BeTrue(), "should return error on single model check with duplicate simple keys")
		g.Expect(dups).To(gomega.HaveLen(1), "duplicates shouldn't be empty when uniqueness check fails")

		// single check for index key
		toCheck = TestModel{UniqueA: model.UniqueA, UniqueB: model.UniqueB}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(gomega.BeTrue(), "should return error on single model check with duplicate composite keys")
		g.Expect(dups).To(gomega.HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// single check with field override fail
		toCheck = TestModel{UniqueA: model.UniqueA, UniqueB: model.UniqueB, OneToOneKey: model.OneToOneKey}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck, []string{"UniqueA", "unique_b"})
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(gomega.BeTrue(), "should return error on single model check with duplicate composite keys and fields overrides")
		g.Expect(dups).To(gomega.HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// single check with field override succeed
		toCheck = TestModel{UniqueA: model.UniqueA, UniqueB: model.UniqueB, OneToOneKey: "don't care"}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck, "OneToOneKey")
		g.Expect(e).To(gomega.Succeed(), "should not return error on single model check with duplicate single keys and fields overrides")

		// multi check failed
		toChecks = []*TestModel{
			{UniqueA: "Not a issue", UniqueB: "shouldn't matter"}, {UniqueA: model.UniqueA, UniqueB: model.UniqueB},
		}
		dups, e = Utils().CheckUniqueness(ctx, toChecks)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(gomega.BeTrue(), "should return error on multi models check with any model containing duplicate keys")
		g.Expect(dups).To(gomega.HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// multi check succeed
		toChecks = []*TestModel{
			{UniqueA: model.UniqueA, UniqueB: "not same"}, {UniqueA: "Not a issue", UniqueB: "shouldn't matter"},
		}
		dups, e = Utils().CheckUniqueness(ctx, toChecks)
		g.Expect(e).To(gomega.Succeed(), "should not return error on multi models check without any model containing duplicate keys")

		// map check failed
		toCheckMap = map[string]interface{}{"UniqueA": model.UniqueA, "unique_b": model.UniqueB}
		dups, e = Utils().Model(&TestModel{}).CheckUniqueness(ctx, toCheckMap)
		g.Expect(errors.Is(e, data.ErrorDuplicateKey)).To(gomega.BeTrue(), "should return error on map check with duplicate keys")
		g.Expect(dups).To(gomega.HaveLen(2), "duplicates shouldn't be empty when uniqueness check fails")

		// map check succeed
		toCheckMap = map[string]interface{}{"UniqueA": "doesn't matter", "unique_b": model.UniqueB}
		dups, e = Utils().Model(&TestModel{}).CheckUniqueness(ctx, toChecks)
		g.Expect(e).To(gomega.Succeed(), "should not return error on map check without duplicate keys")

		// invalid checks
		dups, e = Utils().CheckUniqueness(ctx, map[string]interface{}{})
		g.Expect(e).To(gomega.HaveOccurred(), "should return error of unsupported values")

		toCheck = TestModel{OneToOneKey: "whatever", UniqueA: model.UniqueA, UniqueB: "whatever"}
		dups, e = Utils().CheckUniqueness(ctx, &toCheck, []string{"Invalid", "UniqueB"})
		g.Expect(e).To(gomega.HaveOccurred(), "should return error of invalid field/column name values")

		// all zero value
		dups, e = Utils().CheckUniqueness(ctx, &TestModel{})
		g.Expect(e).To(gomega.HaveOccurred(), "should return error of all zero values")

		// nil value
		dups, e = Utils().CheckUniqueness(ctx, nil)
		g.Expect(e).To(gomega.HaveOccurred(), "should return error of all zero values")

		// different options
		toCheck = TestModel{OneToOneKey: "whatever", UniqueA: model.UniqueA, UniqueB: "whatever"}
		dups, e = Utils(&gorm.Session{NewDB: true}).CheckUniqueness(ctx, &toCheck)
		g.Expect(e).To(gomega.Succeed(), "should return error of all zero values")
	}
}

/*************************
	Helpers
 *************************/

func assertFullyFetchedTestModel(model *TestModel, g *gomega.WithT, target string) {
	g.Expect(model).To(gomega.Not(gomega.BeNil()), "assertFullyFetchedTestModel shouldn't get nil model")
	g.Expect(model.ID).To(gomega.Not(gomega.BeZero()), "%s should populate models", target)
	g.Expect(model.OneToOne).To(gomega.Not(gomega.BeNil()), "%s should populate OneToOne models", target)
	g.Expect(model.ManyToOne).To(gomega.Not(gomega.BeNil()), "%s should populate ManyToOne models", target)
	g.Expect(model.ManyToOne.RelatedMTMModels).To(gomega.Not(gomega.HaveLen(0)),
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
		g.Expect(rs.Error).To(gomega.Succeed(), "truncating table of %s should return no error", table)
	}

	// prepare models
	mtoModels := prepareMTOModels()
	mtmModels, relations := prepareMTMModels()
	models, otoModels := prepareMainModels()
	for _, list := range []interface{}{mtoModels, mtmModels, relations, otoModels, models} {
		rs = db.Create(list)
		g.Expect(rs.Error).To(gomega.Succeed(), "create models shouldn't fail")
		g.Expect(rs.RowsAffected).To(gomega.BeNumerically(">", 0), "create models should create correct number of rows")
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
		g.Expect(rs.Error).To(gomega.Succeed(), "create table if not exists shouldn't fail")
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
