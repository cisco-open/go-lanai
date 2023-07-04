package policy_filter_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	opadata "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data/policy_filter_tests/testdata"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
	"time"
)

/*************************
	Test
 *************************/

func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		dbtest.EnableDBRecordMode(),
	)
}

type OwnerTestDI struct {
	fx.In
	dbtest.DI
}

func TestOPAFilterWithOwnership(t *testing.T) {
	di := &OwnerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(10*time.Minute),
		dbtest.WithDBPlayback("testdb"),
		opatest.WithBundles(opatest.DefaultBundleFS, testdata.ModelBBundleFS),
		apptest.WithConfigFS(testdata.ConfigFS),
		apptest.WithFxOptions(),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareModelB(&di.DI)),
		test.GomegaSubTest(SubTestModelListWithoutTenancy(di), "TestModelListWithoutTenancy"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestPrepareModelB(di *dbtest.DI) test.SetupFunc {
	var models []*ModelB
	var share []*Shared
	closure := func(ctx context.Context, db *gorm.DB) {
		const more = 9
		extra := make([]*ModelB, 0, len(models)*more)
		for _, m := range models {
			key := LookupKey{Owner: m.OwnerID}
			prepareIdLookup(m.ID, key)
			for i := 0; i < more; i++ {
				newM := *m
				newM.ID = uuid.New()
				prepareIdLookup(m.ID, key)
				extra = append(extra, &newM)
			}
		}
		db.WithContext(ctx).CreateInBatches(extra, 50)
	}
	// We use special DB scope to prepare data, to by-pass policy filtering
	return dbtest.PrepareDataWithScope(di,
		dbtest.SetupWithGormScopes(opadata.SkipPolicyFiltering()),
		dbtest.SetupUsingSQLFile(testdata.ModelADataFS, "create_table_b.sql"),
		dbtest.SetupTruncateTables(ModelB{}.TableName(), Shared{}.TableName()),
		dbtest.SetupUsingModelSeedFile(testdata.ModelADataFS, &models, "model_b.yml", closure),
		dbtest.SetupUsingModelSeedFile(testdata.ModelADataFS, &share, "model_b_share.yml"),
	)
}

func SubTestModelListWithoutTenancy(di *OwnerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var models []*ModelB
		var rs *gorm.DB
		// user1
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User1SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(50), "user1 should see %d models", 50)

		// user2
		ctx = testdata.ContextWithSecurityMock(ctx, testdata.User2SecurityOptions())
		rs = di.DB.WithContext(ctx).Model(&ModelB{}).Find(&models)
		g.Expect(rs).To(Not(BeNil()))
		g.Expect(rs.Error).To(Succeed(), "list models should return no error")
		g.Expect(models).To(HaveLen(50), "user1 should see %d models", 50)
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
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value     string
	OwnerName string
	OwnerID   uuid.UUID    `gorm:"type:KeyID;not null" opa:"field:owner_id"`
	Sharing   pqx.JsonbMap `gorm:"type:jsonb;not null" opa:"field:share"`
	Shared    []*Shared    `gorm:"foreignKey:ResID;references:ID" opa:"field:shared"`
	//ShareToMe contstrants.ShareToMe
	OPAPolicyFilter opadata.PolicyFilter `gorm:"-" opa:"type:model"`
	types.Audit
}

func (ModelB) TableName() string {
	return "test_opa_model_b"
}

type Shared struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	ResID      uuid.UUID
	UserID     uuid.UUID
	Operations pq.StringArray `gorm:"type:string[];" opa:"field:operations"`
	Username   string
}

func (Shared) TableName() string {
	return "test_opa_model_b_share"
}
