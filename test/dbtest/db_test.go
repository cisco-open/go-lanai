package dbtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*************************
	Models
 *************************/

type Client struct {
	ID                          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	OAuthClientId               string    `gorm:"uniqueIndex;column:oauth_client_id;not null;"`
}

func (Client) TableName() string {
	return "security_clients"
}

/*************************
	Test
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		EnableDBRecordMode(),
//	)
//}

type testDI struct {
	fx.In
	DB *gorm.DB        `optional:"true"`
}

func TestDBPlayback(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDBPlayback("testdb"),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestExampleSelect(di), "Select"),
		test.GomegaSubTest(SubTestExampleTxSave(di), "TransactionalSave"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleSelect(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.DB).To(Not(BeNil()), "injected gorm.DB should not be nil")

		// select one
		v := Client{}
		r := di.DB.WithContext(ctx).Model(&Client{}).First(&v)
		g.Expect(r.Error).To(Succeed(), "recorded SQL shouldn't introduce error")
		g.Expect(v.ID).To(Not(Equal(uuid.Invalid)), "model should be loaded by First()")

		// select all
		s := make([]*Client, 0)
		r = di.DB.WithContext(ctx).Model(&Client{}).Find(&s)
		g.Expect(r.Error).To(Succeed(), "recorded SQL shouldn't introduce error")
		g.Expect(s).To(Not(BeEmpty()), "slice should not be empty after Find")
	}
}

func SubTestExampleTxSave(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.DB).To(Not(BeNil()), "injected gorm.DB should not be nil")

		e := tx.Transaction(ctx, func(ctx context.Context) error {
			// select one
			v := Client{}
			r := di.DB.WithContext(ctx).Model(&Client{}).First(&v)
			g.Expect(r.Error).To(Succeed(), "select SQL shouldn't introduce error")
			g.Expect(v.ID).To(Not(Equal(uuid.Invalid)), "model should be loaded by First()")

			// save one
			r = di.DB.WithContext(ctx).Save(&v)
			g.Expect(r.Error).To(Succeed(), "save operation shouldn't introduce error")
			return r.Error
		})
		g.Expect(e).To(Succeed(), "tx.Transaction should return no error")
	}
}
