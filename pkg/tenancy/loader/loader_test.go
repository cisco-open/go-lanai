package th_loader

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/embedded"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"fmt"
	goredis "github.com/go-redis/redis/v8"
	"github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*************************
	Test
 *************************/

func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		embedded.Redis(),
	)
}

type testDI struct {
	fx.In
	DB             *gorm.DB `optional:"true"`
	InternalLoader TenancyLoader
}

type MockedTenantHierarchyStore struct {
}

type MockTenantIterator struct {
}

func (m MockTenantIterator) Next() bool {
	return false
}

func (m MockTenantIterator) Scan(ctx context.Context) (Tenant, error) {
	return nil, nil
}

func (m MockTenantIterator) Close() error {
	return nil
}

func (m MockTenantIterator) Err() error {
	return nil
}

func (m MockedTenantHierarchyStore) GetIterator(ctx context.Context) (TenantIterator, error) {
	return &MockTenantIterator{}, nil
}

func provideMockedTenancyLoader() TenancyLoader {
	return TenancyLoader{
		store: MockedTenantHierarchyStore{},
		accessor: &mocks.MockTenancyAccessor{
			Isloaded: false,
		},
	}
}

func TestLoader(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(
				provideMockedTenancyLoader,
			),
		),
		test.GomegaSubTest(SubTestLoadTenantHierarchy(di), "SubTestLoadTenantHierarchy"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLoadTenantHierarchy(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		universal := &goredis.UniversalOptions{}
		opts := universal.Simple()
		opts.Addr = fmt.Sprintf("127.0.0.1:%d", embedded.CurrentRedisPort())
		client := goredis.NewClient(opts)
		defer func() { _ = client.Close() }()

		di.InternalLoader.rc = client

		err := di.InternalLoader.LoadTenantHierarchy(ctx)
		g.Expect(err.Error()).To(gomega.Equal("redis: nil"))
	}
}
