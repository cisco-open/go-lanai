package dbtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/repo"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"embed"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"sync"
	"testing"
	"time"
)

/*************************
	Setup
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		EnableDBRecordMode(),
//	)
//}

/***************************
	Data to Test
 ***************************/

type TestCat struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Name string    `gorm:"not null"`
	// Relationships many2many
	Toys []*TestToy `gorm:"many2many:test_cat_toys;joinForeignKey:cat_id;joinReferences:toy_id;constraint:OnDelete:CASCADE"`
}

type TestToy struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	What string    `gorm:"not null"`
	// Relationships many2many
	Owners []*TestCat `gorm:"many2many:test_cat_toys;joinForeignKey:toy_id;joinReferences:cat_id"`
}

//go:embed testdata/*.yml testdata/*.sql
var testDataFS embed.FS

var testCatIDs []uuid.UUID

func DropAllTables() DataSetupStep {
	return SetupDropTables("test_cat_toys", "test_cats", "test_toys")
}

func CreateAllTables() DataSetupStep {
	return SetupUsingSQLFile(testDataFS, "testdata/tables.sql")
}

func TruncateAllTables() DataSetupStep {
	return SetupTruncateTables("test_cat_toys", "test_cats", "test_toys")
}

func SeedCats() DataSetupStep {
	var cats []*TestCat
	return SetupUsingModelSeedFile(testDataFS, &cats, "testdata/model_cats.yml", func(ctx context.Context, db *gorm.DB) {
		testCatIDs = make([]uuid.UUID, len(cats))
		for i, c := range cats {
			testCatIDs[i] = c.ID
		}
	})
}

func SeedToys() DataSetupStep {
	var toys []*TestToy
	return SetupUsingModelSeedFile(testDataFS, &toys, "testdata/model_toys.yml")
}

func SeedCatToysRelations() DataSetupStep {
	return SetupUsingSQLFile(testDataFS, "testdata/relation_cat_toys.sql")
}

func SetupScopeTestPrepareTables(di *DI) test.SetupFunc {
	var once sync.Once
	return PrepareData(di,
		SetupOnce(&once,
			DropAllTables(),
			CreateAllTables(),
		),
		TruncateAllTables(),
	)
}

func SetupScopeTestPrepareData(di *DI) test.SetupFunc {
	return PrepareDataWithScope(di,
		SetupWithGormScopes(func(db *gorm.DB) *gorm.DB {
			return db.Set("scopeKey", "scopeValue")
		}),
		SeedCats(), SeedToys(), SeedCatToysRelations(),
	)
}

/*************************
	Tests
 *************************/

type dataUtilsTestDI struct {
	fx.In
	DI
}

func TestScopeController(t *testing.T) {
	di := &dataUtilsTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDBPlayback("testdb"),
		apptest.WithModules(repo.Module),
		apptest.WithTimeout(time.Minute),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupScopeTestPrepareTables(&di.DI)),
		test.SubTestSetup(SetupScopeTestPrepareData(&di.DI)),
		test.GomegaSubTest(SubTestCats(di), "TestCats"),
		test.GomegaSubTest(SubTestToys(di), "TestToys"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestCats(di *dataUtilsTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var cats []*TestCat
		var rs *gorm.DB
		rs = di.DB.WithContext(ctx).Preload("Toys").Find(&cats)
		g.Expect(rs.Error).To(Succeed(), "Find []*TestCat shouldn't return error")
		g.Expect(cats).To(HaveLen(3), "Should find 3 cats")
		for _, cat := range cats {
			g.Expect(cat.ID).To(Not(BeZero()), "each cat should have valid ID")
			g.Expect(cat.Name).To(Not(BeEmpty()), "each cat should have valid Name")
			g.Expect(cat.Toys).To(HaveLen(2), "each cat should have 2 toys")
		}
	}
}

func SubTestToys(di *dataUtilsTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var toys []*TestToy
		var rs *gorm.DB
		rs = di.DB.WithContext(ctx).Preload("Owners").Find(&toys)
		g.Expect(rs.Error).To(Succeed(), "Find []*TestToy shouldn't return error")
		g.Expect(toys).To(HaveLen(3), "Should find 3 toys")
		for _, toy := range toys {
			g.Expect(toy.ID).To(Not(BeZero()), "each toy should have valid ID")
			g.Expect(toy.What).To(Not(BeEmpty()), "each toy should have valid description")
			g.Expect(toy.Owners).To(HaveLen(2), "each toy should have 2 owners")
		}
	}
}

/*************************
	Setups
 *************************/
