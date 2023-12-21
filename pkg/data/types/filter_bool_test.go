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

package types

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
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
)

var (
	FilterValue = map[bool]string{
		true:  "Filtered",
		false: "Not Filtered",
	}
	FilterSearch = map[bool]string{
		true:  "filtered",
		false: "not_filtered",
	}
)

var (
	MTOModelIdFiltered    = uuid.MustParse("7867eaf6-fdbe-4271-8f0b-bae68a0343ab")
	MTOModelIdNotFiltered = uuid.MustParse("26fbc037-6885-4c46-92d7-2bea5eefbf92")

	TestModelIDs = map[[3]bool]uuid.UUID{
		[3]bool{false, false, false}: uuid.MustParse("bf8e3769-cf07-4897-a982-8fa26d837df3"),
		[3]bool{false, true, false}:  uuid.MustParse("2c4e11dc-ac2c-4ed9-bcf7-f6508fd754bf"),
		[3]bool{false, false, true}:  uuid.MustParse("45e179a7-5856-4035-ad3a-df7b2be27c7b"),
		[3]bool{false, true, true}:   uuid.MustParse("090c7373-a146-450d-aa97-3db8593a6046"),
		[3]bool{true, false, false}:  uuid.MustParse("e27ac9ea-00cb-4427-bfdd-c488b1dbe616"),
		[3]bool{true, true, false}:   uuid.MustParse("1609e178-d647-4ebd-9753-a9096b4267b2"),
		[3]bool{true, false, true}:   uuid.MustParse("dcc796da-9e83-401b-9064-314ab6133a42"),
		[3]bool{true, true, true}:    uuid.MustParse("700f60e4-f81d-48b0-bf64-6877a99f150c"),
	}
	TestModelNoMTOIDs = map[[2]bool]uuid.UUID{
		[2]bool{false, false}: uuid.MustParse("3a822183-60be-4d1e-b4dd-79790294581a"),
		[2]bool{false, true}:  uuid.MustParse("c6337c29-30f2-4c1b-bad7-2a41a040f7cb"),
		[2]bool{true, false}:  uuid.MustParse("2e7ea821-45b5-490e-a611-9b1b4aacc5a2"),
		[2]bool{true, true}:   uuid.MustParse("1f68668a-9abf-45a9-9fd9-fe3d2ab9b13d"),
	}
)

/*************************
	Test
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type testBoolDI struct {
	fx.In
	DB *gorm.DB
}

func TestBoolFilter(t *testing.T) {
	di := &testBoolDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithFxOptions(
			fx.Provide(provideMockedTenancyAccessor),
		),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupBoolFilterTestPrepareData(di)),
		test.GomegaSubTest(SubTestFilterWithoutJoin(di), "DefaultFilterWithoutJoin"),
		test.GomegaSubTest(SubTestNegFilterWithoutJoin(di), "NegFilterWithoutJoin"),
		test.GomegaSubTest(SubTestOptInFilterWithoutJoin(di), "OptInFilterWithoutJoin"),
		test.GomegaSubTest(SubTestFilterWithOverride(di), "FilterWithOverride"),
		test.GomegaSubTest(SubTestFilterWithScope(di), "FilterWithAdditionalScope"),
		// Skipped, see sub-test's comment for reason
		//test.GomegaSubTest(SubTestFilterWithOneToOneJoin(di), "FilterWithOneToOneJoin"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupBoolFilterTestPrepareData(di *testBoolDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		prepareBoolFilterTables(di.DB, g)
		prepareTestModelData(di.DB, g)
		return ctx, nil
	}
}

func SubTestFilterWithoutJoin(di *testBoolDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var tx *gorm.DB
		var model TestModel
		var models []*TestModel
		fieldExtractor := func(m *TestModel) interface{} {
			return bool(m.Filtered)
		}
		// get by ID
		var id uuid.UUID
		id, _ = TestModelIDs[[3]bool{false, false, false}]
		tx = di.DB.Take(&model, id)
		g.Expect(tx.Error).To(Succeed(), "SELECT by ID without join shouldn't filter by relations")

		id, _ = TestModelIDs[[3]bool{true, true, true}]
		tx = di.DB.Take(&model, id)
		g.Expect(tx.Error).To(HaveOccurred(), "SELECT by ID without join should filter by its own field")

		// find all
		models = nil
		tx = di.DB.Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT * without join shouldn't return error")
		g.Expect(models).To(HaveLen(6), "SELECT * without join shouldn't filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(false)), "SELECT * without join should filter correctly")
		}

		// find with WHERE
		models = nil
		tx = di.DB.Where("many_to_one_id IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE without join shouldn't return error")
		g.Expect(models).To(HaveLen(4), "SELECT with WHERE without join shouldn't filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(false)), "SELECT * without join should filter correctly")
		}
	}
}

func SubTestNegFilterWithoutJoin(di *testBoolDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var tx *gorm.DB
		var models []*TestOTOModel
		fieldExtractor := func(m *TestOTOModel) interface{} {
			return bool(m.OTOEnabled)
		}
		// find all
		models = nil
		tx = di.DB.Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT * without join shouldn't return error")
		g.Expect(models).To(HaveLen(6), "SELECT * without join shouldn't filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(true)), "SELECT * without join should filter correctly")
		}

		// find with WHERE
		models = nil
		tx = di.DB.Where("value IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE without join shouldn't return error")
		g.Expect(models).To(HaveLen(6), "SELECT with WHERE without join shouldn't filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(true)), "SELECT * without join should filter correctly")
		}
	}
}

func SubTestOptInFilterWithoutJoin(di *testBoolDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var tx *gorm.DB
		var models []*TestMTOModel

		// find all
		models = nil
		tx = di.DB.Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT * with skipped filter without join shouldn't return error")
		g.Expect(models).To(HaveLen(2), "SELECT * with skipped filter without join shouldn't filter")

		// find with WHERE
		models = nil
		tx = di.DB.Where("value IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE with skipped filter without join shouldn't return error")
		g.Expect(models).To(HaveLen(2), "SELECT with WHERE with skipped filter without join shouldn't filter by relations")

		// find all opt-in filter
		models = nil
		tx = di.DB.Scopes(BoolFiltering(true)).Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT * with opt-in filter without join shouldn't return error")
		g.Expect(models).To(HaveLen(1), "SELECT * with opt-in filter without join shouldn't filter")


	}
}

func SubTestFilterWithOverride(di *testBoolDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var tx *gorm.DB
		var models []*TestModel
		fieldExtractor := func(m *TestModel) interface{} {
			return bool(m.Filtered)
		}

		// with field disabler
		models = nil
		tx = di.DB.Scopes(SkipBoolFilter("Filtered")).Where("many_to_one_id IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE without join shouldn't return error when filter is disabled on one field")
		g.Expect(models).To(HaveLen(8), "SELECT with WHERE without join shouldn't filter by relations when filter is disabled on one field")

		// with * disabler
		models = nil
		tx = di.DB.Scopes(SkipBoolFilter()).Where("many_to_one_id IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE without join shouldn't return error when filter is disabled on all fields")
		g.Expect(models).To(HaveLen(8), "SELECT with WHERE without join shouldn't filter by relations when filter is disabled on all fields")

		// with overridden behavior on one field
		models = nil
		tx = di.DB.Scopes(BoolFiltering(false, "Filtered")).Where("many_to_one_id IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE without join shouldn't return error when filter behavior is altered")
		g.Expect(models).To(HaveLen(4), "SELECT with WHERE without join should filter based on alternative behavior on specified field")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(true)),
				"SELECT * without join should filter correctly based on alternative behavior on specified field")
		}

		// with overridden behavior on all field
		models = nil
		tx = di.DB.Scopes(BoolFiltering(false)).Where("many_to_one_id IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE without join shouldn't return error when filter behavior is altered")
		g.Expect(models).To(HaveLen(4), "SELECT with WHERE without join should filter based on alternative behavior on all fields")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(true)),
				"SELECT * without join should filter correctly based on alternative behavior on all fields")
		}

	}
}

func SubTestFilterWithScope(di *testBoolDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var tx *gorm.DB
		var model TestModel
		var models []*TestModel
		fieldExtractor := func(m *TestModel) interface{} {
			return bool(m.Filtered)
		}
		noopScope := func(db *gorm.DB) *gorm.DB { return db }
		// get by ID
		var id uuid.UUID
		id, _ = TestModelIDs[[3]bool{false, false, false}]
		tx = di.DB.Scopes(noopScope).Take(&model, id)
		g.Expect(tx.Error).To(Succeed(), "SELECT by ID without join shouldn't filter by relations")

		id, _ = TestModelIDs[[3]bool{true, true, true}]
		tx = di.DB.Scopes(noopScope).Take(&model, id)
		g.Expect(tx.Error).To(HaveOccurred(), "SELECT by ID without join should filter by its own field")

		// find all
		models = nil
		tx = di.DB.Scopes(noopScope).Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT * without join shouldn't return error")
		g.Expect(models).To(HaveLen(6), "SELECT * without join shouldn't filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(false)), "SELECT * without join should filter correctly")
		}

		// find with WHERE
		models = nil
		tx = di.DB.Scopes(noopScope).Where("many_to_one_id IS NOT NULL").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE without join shouldn't return error")
		g.Expect(models).To(HaveLen(4), "SELECT with WHERE without join shouldn't filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(false)), "SELECT * without join should filter correctly")
		}
	}
}

// SubTestFilterWithOneToOneJoin Skipped
// As of gorm 1.21.11, gorm doesn't add schema.Schema.QueryClauses with joined models, so this test case is not
// possible to work.
func SubTestFilterWithOneToOneJoin(di *testBoolDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var tx *gorm.DB
		var model TestModel
		var models []*TestModel
		fieldExtractor := func(m *TestModel) interface{} {
			return bool(m.Filtered)
		}
		// get by ID
		var id uuid.UUID
		id, _ = TestModelIDs[[3]bool{false, false, false}]
		tx = di.DB.Joins("OneToOne").Take(&model, id)
		g.Expect(tx.Error).To(HaveOccurred(), "SELECT by ID with one-to-one join should also filter by relations")

		id, _ = TestModelIDs[[3]bool{true, true, true}]
		tx = di.DB.Joins("OneToOne").Take(&model, id)
		g.Expect(tx.Error).To(HaveOccurred(), "SELECT by ID with one-to-one join should filter by its own field")

		id, _ = TestModelIDs[[3]bool{false, false, false}]
		tx = di.DB.Joins("OneToOne").Take(&model, id)
		g.Expect(tx.Error).To(Succeed(), "SELECT by ID with one-to-one join should return unfiltered item")

		// find all
		models = nil
		tx = di.DB.Joins("OneToOne").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT * with one-to-one join shouldn't return error")
		g.Expect(models).To(HaveLen(3), "SELECT * with one-to-one join should also filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(false)), "SELECT * with one-to-one join should filter correctly")
		}

		// find with WHERE
		models = nil
		tx = di.DB.Where("many_to_one_id IS NOT NULL").Joins("OneToOne").Find(&models)
		g.Expect(tx.Error).To(Succeed(), "SELECT with WHERE with one-to-one join shouldn't return error")
		g.Expect(models).To(HaveLen(2), "SELECT with WHERE with one-to-one join shouldn also filter by relations")
		for _, m := range models {
			g.Expect(m).To(WithTransform(fieldExtractor, Equal(false)), "SELECT * with one-to-one join should filter correctly")
		}

		// TODO with disabled filter
		//models = nil
		//tx = di.DB.Find(&models)
		//g.Expect(tx.Error).To(Succeed(), "SELECT * without join shouldn't return error")
		//g.Expect(len(models)).To(Equal(6), "SELECT * without join shouldn't filter by relations")
	}
}

/*************************
	Data
 *************************/

func prepareTestModelData(db *gorm.DB, g *gomega.WithT) {
	// truncate table
	tables := []string{
		TestModel{}.TableName(),
		TestOTOModel{}.TableName(),
		TestMTOModel{}.TableName(),
	}
	for _, table := range tables {
		r := db.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, table))
		g.Expect(r.Error).To(Succeed(), "truncating table of %s should return no error", table)
	}
	// create many-to-one models
	createMTOModelRecord(MTOModelIdNotFiltered, false, db, g)
	createMTOModelRecord(MTOModelIdFiltered, true, db, g)

	// create test models and one-to-one models
	for k, id := range TestModelIDs {
		createTestModelRecordWithMTO(id, k[0], k[1], k[2], db, g)
	}

	// create test models and one-to-one models
	for k, id := range TestModelNoMTOIDs {
		createTestModelRecord(id, k[0], k[1], nil, db, g)
	}
}

func createMTOModelRecord(id uuid.UUID, filtered bool, db *gorm.DB, g *gomega.WithT) {
	m := TestMTOModel{
		ID:             id,
		RelationValue:  FilterValue[filtered],
		RelationSearch: FilterSearch[filtered],
		MTOEnabled:     FilterBool(filtered),
	}
	tx := db.Create(&m)
	g.Expect(tx.Error).To(Succeed(), "create %s of value %s shouldn't fail", m.TableName(), m.RelationValue)
}

func createOTOModelRecord(key string, filtered bool, db *gorm.DB, g *gomega.WithT) {
	m := TestOTOModel{
		RefKey:         key,
		RelationValue:  FilterValue[filtered],
		RelationSearch: FilterSearch[filtered],
		OTOEnabled:     FilterBool(filtered),
	}
	tx := db.Create(&m)
	g.Expect(tx.Error).To(Succeed(), "create %s of value %s shouldn't fail", m.TableName(), m.RelationValue)
}

func createTestModelRecordWithMTO(id uuid.UUID, filtered, otoFiltered, mtoFiltered bool, db *gorm.DB, g *gomega.WithT) {
	mtoId := MTOModelIdNotFiltered
	if mtoFiltered {
		mtoId = MTOModelIdFiltered
	}
	createTestModelRecord(id, filtered, otoFiltered, &mtoId, db, g)
}

func createTestModelRecord(id uuid.UUID, filtered, otoFiltered bool, mtoId *uuid.UUID, db *gorm.DB, g *gomega.WithT) {
	otoKey := utils.RandomString(8)
	createOTOModelRecord(otoKey, otoFiltered, db, g)
	m := TestModel{
		ID:          id,
		Value:       FilterValue[filtered],
		Search:      FilterSearch[filtered],
		Filtered:    FilterBool(filtered),
		OneToOneKey: otoKey,
		ManyToOneID: mtoId,
	}
	tx := db.Create(&m)
	g.Expect(tx.Error).To(Succeed(), "create %s of value %s shouldn't fail", m.TableName(), m.Value)
}

/*************************
	Mocks
 *************************/

const boolFilterTableSQL1 = `
CREATE TABLE IF NOT EXISTS public.test_bool_filter_model2 (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"value" STRING,
	search STRING NOT NULL,
	enabled BOOL NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	INDEX idx_search (search ASC),
	INDEX idx_filtered (enabled ASC),
	FAMILY "primary" (id, "value", search, enabled)
);`

const boolFilterTableSQL2 = `
CREATE TABLE IF NOT EXISTS public.test_bool_filter_model1 (
	ref_key STRING NOT NULL,
	"value" STRING,
	search STRING NOT NULL,
	enabled BOOL NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (ref_key ASC),
	INDEX idx_search (search ASC),
	INDEX idx_filtered (enabled ASC),
	FAMILY "primary" (ref_key, "value", search, enabled)
);`

const boolFilterTableSQL3 = `
CREATE TABLE IF NOT EXISTS public.test_bool_filter_models (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"value" STRING,
	search STRING NOT NULL,
	filtered BOOL NOT NULL,
	one_to_one_key STRING NOT NULL,
	many_to_one_id UUID NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	CONSTRAINT fk_ont_to_one FOREIGN KEY (one_to_one_key) REFERENCES public.test_bool_filter_model1(ref_key),
	CONSTRAINT fk_many_to_one FOREIGN KEY (many_to_one_id) REFERENCES public.test_bool_filter_model2(id) ON DELETE SET NULL,
	UNIQUE INDEX idx_ont_to_one (one_to_one_key ASC),
	INDEX idx_many_to_one (many_to_one_id ASC),
	INDEX idx_search (search ASC),
	INDEX idx_filtered (filtered ASC),
	FAMILY "primary" (id, "value", one_to_one_key, many_to_one_id, search, filtered)
);`

func prepareBoolFilterTables(db *gorm.DB, g *gomega.WithT) {
	r := db.Exec(boolFilterTableSQL1)
	g.Expect(r.Error).To(Succeed(), "create table 1 if not exists shouldn't fail")
	r = db.Exec(boolFilterTableSQL2)
	g.Expect(r.Error).To(Succeed(), "create table 2 if not exists shouldn't fail")
	r = db.Exec(boolFilterTableSQL3)
	g.Expect(r.Error).To(Succeed(), "create table 3 if not exists shouldn't fail")
}

type TestModel struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value       string
	Search      string
	Filtered    FilterBool
	OneToOneKey string
	ManyToOneID *uuid.UUID
	OneToOne    *TestOTOModel `gorm:"foreignKey:RefKey;references:OneToOneKey;not null"`
	ManyToOne   *TestMTOModel `gorm:"foreignKey:ManyToOneID;"`
}

func (TestModel) TableName() string {
	return "test_bool_filter_models"
}

type TestOTOModel struct {
	RefKey         string     `gorm:"primary_key;column:ref_key;type:TEXT;"`
	RelationValue  string     `gorm:"column:value;"`
	RelationSearch string     `gorm:"column:search;"`
	OTOEnabled     FilterBool `gorm:"column:enabled;" filter:"false"`
}

func (TestOTOModel) TableName() string {
	return "test_bool_filter_model1"
}

type TestMTOModel struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	RelationValue  string     `gorm:"column:value;"`
	RelationSearch string     `gorm:"column:search;"`
	MTOEnabled     FilterBool `gorm:"column:enabled;" filter:"-"`
}

func (TestMTOModel) TableName() string {
	return "test_bool_filter_model2"
}
