package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"reflect"
	"testing"
)

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

func TestGormCRUDRepository(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(NewTestRepository),
		),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestCreateTables(di)),
		test.GomegaSubTest(SubTestSchemaResolver(di), "TestSchemaResolver"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestCreateTables(di *testDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		prepareTable(di.DB, g)
		return ctx, nil
	}
}

func SubTestSchemaResolver(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Repo.ModelType()).
			To(BeEquivalentTo(reflect.TypeOf(TestModel{})), "ModelType should be correct")
		g.Expect(di.Repo.Table()).
			To(Equal("test_repo_models"), "Table should be correct")
		g.Expect(di.Repo.ColumnName("Value")).
			To(Equal("value"), "ColumnName of direct field should be correct")
		g.Expect(di.Repo.ColumnDataType("Value")).
			To(Equal("string"), "ColumnDataType of direct field should be correct")

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

/*************************
	Repository
 *************************/

type TestRepository CrudRepository

func NewTestRepository(factory Factory) TestRepository {
	return factory.NewCRUD(&TestModel{})
}

/*************************
	Mocks
 *************************/

const tableSQL1 = `
CREATE TABLE IF NOT EXISTS public.test_repo_model2 (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"value" STRING,
	search STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	INDEX idx_search (search ASC),
	FAMILY "primary" (id, "value")
);`

const tableSQL2 = `
CREATE TABLE IF NOT EXISTS public.test_repo_model1 (
	ref_key STRING NOT NULL,
	"value" STRING,
	search STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (ref_key ASC),
	INDEX idx_search (search ASC),
	FAMILY "primary" (ref_key, "value")
);`

const tableSQL3 = `
CREATE TABLE IF NOT EXISTS public.test_repo_models (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"value" STRING,
	search STRING NOT NULL,
	one_to_one_key STRING NOT NULL,
	many_to_one_id UUID NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	CONSTRAINT fk_ont_to_one FOREIGN KEY (one_to_one_key) REFERENCES public.test_repo_model1(ref_key),
	CONSTRAINT fk_many_to_one FOREIGN KEY (many_to_one_id) REFERENCES public.test_repo_model2(id) ON DELETE SET NULL,
	UNIQUE INDEX idx_ont_to_one (one_to_one_key ASC),
	INDEX idx_many_to_one (many_to_one_id ASC),
	INDEX idx_search (search ASC),
	FAMILY "primary" (id, "value", one_to_one_key, many_to_one_id)
);`

func prepareTable(db *gorm.DB, g *gomega.WithT) {
	r := db.Exec(tableSQL1)
	g.Expect(r.Error).To(Succeed(), "create table 1 if not exists shouldn't fail")
	r = db.Exec(tableSQL2)
	g.Expect(r.Error).To(Succeed(), "create table 2 if not exists shouldn't fail")
	r = db.Exec(tableSQL3)
	g.Expect(r.Error).To(Succeed(), "create table 3 if not exists shouldn't fail")
}

type TestModel struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value       string
	Search      string
	OneToOneKey string
	ManyToOneID uuid.UUID
	OneToOne    *TestOTOModel `gorm:"foreignKey:RefKey;references:OneToOneKey;not null"`
	ManyToOne   *TestMTOModel `gorm:"foreignKey:ManyToOneID;"`
}

func (TestModel) TableName() string {
	return "test_repo_models"
}

type TestOTOModel struct {
	RefKey         uuid.UUID `gorm:"primary_key;column:ref_key;type:TEXT;"`
	RelationValue  string    `gorm:"column:value;"`
	RelationSearch string    `gorm:"column:search;"`
}

func (TestOTOModel) TableName() string {
	return "test_repo_model1"
}

type TestMTOModel struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	RelationValue  string    `gorm:"column:value;"`
	RelationSearch string    `gorm:"column:search;"`
}

func (TestMTOModel) TableName() string {
	return "test_repo_model2"
}
