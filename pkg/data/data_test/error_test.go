package data_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

var (
	TestModelID1 = uuid.MustParse("92d22359-6e61-4407-adf1-cee2ae8b8262")
	TestModelID2 = uuid.MustParse("63299139-748d-44bc-bf9a-cdd79389ad68")
	TestModelID3 = uuid.MustParse("3468d223-11ed-4dfc-9e50-fd385dc57099")
	PreparedModels = map[string]uuid.UUID{
		"Model-1": TestModelID1,
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

type errTestDI struct {
	fx.In
	DB           *gorm.DB
}

func TestGormModel(t *testing.T) {
	di := &errTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithFxOptions(
			//fx.Provide(provideMockedTenancyAccessor),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupWithTable(di)),
		test.SubTestTeardown(TeardownWithTruncateTable(di)),
		test.GomegaSubTest(SubTestServerSideErrorTranslation(di), "ServerSideErrorTranslation"),
	)
}

/*************************
	Test
 *************************/

func SetupWithTable(di *errTestDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		r := di.DB.Exec(tableSQL)
		g.Expect(r.Error).To(Succeed(), "truncate table shouldn't fail")
		for k, v := range PreparedModels {
			m := ErrorTestModel{
				ID:        v,
				UniqueKey: k,
				Value:     fmt.Sprintf("Value of %s", k),
			}
			r = di.DB.Create(&m)
			g.Expect(r.Error).To(Succeed(), "create model [%s] shouldn't fail", k)
		}
		return ctx, nil
	}
}

func TeardownWithTruncateTable(di *errTestDI) test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		return di.DB.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" RESTRICT`, ErrorTestModel{}.TableName())).Error
	}
}

func SubTestServerSideErrorTranslation(di *errTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// duplicated key
		m := ErrorTestModel{
			ID:        TestModelID2,
			UniqueKey: "Model-1", // duplicated key
			Value:     "what ever",
		}
		expected := data.NewDuplicateKeyError("mocked error")
		r := di.DB.Save(&m)
		g.Expect(r.Error).To(HaveOccurred(), "create model [%s] should fail", m.UniqueKey)
		g.Expect(r.Error).To(BeAssignableToTypeOf(expected), "error should be data.DataError type")
		g.Expect(errors.Is(r.Error, expected)).To(BeTrue(), "error should match DuplicateKeyError")
		g.Expect(r.Error.(data.DataError).Cause()).To(BeAssignableToTypeOf(&pq.Error{}), "error should have cause with pq.Error type")
	}
}

/*************************
	Mocks
 *************************/

type ErrorTestModel struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	UniqueKey string    `gorm:"uniqueIndex;column:uk"`
	Value     string
}

func (ErrorTestModel) TableName() string {
	return "test_errors"
}

const tableSQL = `
CREATE TABLE IF NOT EXISTS public.test_errors (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"uk" STRING NOT NULL,
	"value" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	UNIQUE INDEX idx_unique_key (uk ASC),
	FAMILY "primary" (id, uk, value)
);`
