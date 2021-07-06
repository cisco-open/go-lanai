package dbtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
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

// TestMain is the only place we should kick off mocked CockroachDB
func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		//CockroachDB(ModeRecord, "usermanagement"),
		CockroachDB(ModePlayback, "usermanagement"),
	)
}

type testDI struct {
	fx.In
	DB *gorm.DB        `optional:"true"`
}

func TestDBRecord(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		Initialize(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestExampleTestSQL(di), "TestSQL"),
	)
}

func TestPlayback(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		Initialize(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestExampleTestSQL(di), "TestSQL"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleTestSQL(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.DB).To(Not(BeNil()), "injected gorm.DB should not be nil")

		// select one
		v := Client{}
		r := di.DB.WithContext(ctx).Model(&Client{}).First(&v)
		g.Expect(r.Error).To(Succeed(), "recorded SQL shouldn't introduce error")
		g.Expect(v.ID).To(Not(Equal(uuid.Invalid)), "model should be loaded by First()")

		// save one
		r = di.DB.WithContext(ctx).Save(&v)
		g.Expect(r.Error).To(Succeed(), "recorded SQL shouldn't introduce error")

		// select all
		s := make([]*Client, 0)
		r = di.DB.WithContext(ctx).Model(&Client{}).Find(&s)
		g.Expect(r.Error).To(Succeed(), "recorded SQL shouldn't introduce error")
		g.Expect(s).To(Not(BeEmpty()), "slice should not be empty after Find")
	}
}

