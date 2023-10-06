package dbtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/cockroach"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*************************
	Tests
 *************************/

type gormDI struct {
	fx.In
	GormDB *gorm.DB
}

func TestGormWithDBPlayback(t *testing.T) {
	di := gormDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDBPlayback("testdb"),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestGormDialetorValidation(&di, &cockroach.GormDialector{}), "GormDialetorValidation"),
	)
}

func TestNoopGorm(t *testing.T) {
	di := gormDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithNoopMocks(),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestGormDialetorValidation(&di, noopGormDialector{}), "GormDialetorValidation"),
		test.GomegaSubTest(SubTestGormDryRun(&di), "TestGormDryRun"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestGormDialetorValidation(di *gormDI, expected interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.GormDB).To(Not(BeNil()), "*gorm.DB should not be nil")
		g.Expect(di.GormDB.Dialector).To(Not(BeNil()), "Dialector should not be nil")
		g.Expect(di.GormDB.Dialector).To(BeAssignableToTypeOf(expected), "Dialector should be expected type")
	}
}

func SubTestGormDryRun(di *gormDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		rs = di.GormDB.Create(&Model{Value: "doesn't matter"})
		g.Expect(rs.Error).To(Succeed(), "create should succeed")

		var models []*Model
		rs = di.GormDB.Find(&models)
		g.Expect(rs.Error).To(Succeed(), "find should succeed")
	}
}

type Model struct {
	ID    string `gorm:"primaryKey;"`
	Value string
}
