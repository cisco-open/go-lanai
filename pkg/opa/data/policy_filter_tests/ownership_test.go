package policy_filter_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	opadata "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data/policy_filter_tests/testdata"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
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

type OwnerTestDI struct {
	fx.In
	dbtest.DI
}

func TestOPAFilterWithOwnership(t *testing.T) {
	di := &OwnerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		opatest.WithBundles(),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithFxOptions(),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestCreateModels(&di.DI)),
		//test.GomegaSubTest(SubTestModelListWithoutTenancy(di), "TestModelListWithoutTenancy"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestModelListWithoutTenancy(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelB
		var rs *gorm.DB
		// user1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(5), "user1 should see %d models", 5)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(4), "user1 should see %d models", 4)
	}
}

/*************************
	Helpers
 *************************/




/*************************
	Models
 *************************/

// ModelB has no tenancy
type ModelB struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value           string
	TenantName      string
	OwnerName       string
	OwnerID         uuid.UUID    `gorm:"type:KeyID;not null" opa:"field:owner_id"`
	OPAPolicyFilter opadata.PolicyFilter `gorm:"-" opa:"type:poc"`
	types.Audit
}

func (ModelB) TableName() string {
	return "test_opa_model_a"
}

