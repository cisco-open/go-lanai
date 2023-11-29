package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

func TestClientUserTenantResolver(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(2*time.Minute),
		apptest.WithModules(
			tenancy.Module,
		),
		apptest.WithFxOptions(
			fx.Provide(ProvideMockedTenancyAccessor),
		),
		test.GomegaSubTest(SubTestClientOnly(), "TestClientOnly"),
		test.GomegaSubTest(SubTestClientAndUser(), "TestClientAndUser"),
	)
}

func SubTestClientOnly() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := sectest.MockedClient{
			MockedClientProperties: sectest.MockedClientProperties{
				ClientID: "test-client",
				AssignedTenantIds: []string{
					"id-tenant-1",
				},
			},
		}
		defaultTenantId, tenantIds, e := ResolveClientUserTenants(ctx, nil, client)
		g.Expect(e).To(Not(HaveOccurred()))
		g.Expect(defaultTenantId).To(Equal("id-tenant-1"))
		g.Expect(tenantIds).To(Equal([]string{"id-tenant-1"}))

		client = sectest.MockedClient{
			MockedClientProperties: sectest.MockedClientProperties{
				ClientID: "test-client",
				AssignedTenantIds: []string{
					"id-tenant-1",
					"id-tenant-2",
				},
			},
		}
		defaultTenantId, tenantIds, e = ResolveClientUserTenants(ctx, nil, client)
		g.Expect(e).To(Not(HaveOccurred()))
		g.Expect(defaultTenantId).To(Equal("")) // this client has multiple tenants, so we can't pick a default
		g.Expect(len(tenantIds)).To(Equal(2))
		g.Expect(utils.NewStringSet(tenantIds...)).To(Equal(utils.NewStringSet("id-tenant-1", "id-tenant-2")))
	}
}

func SubTestClientAndUser() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := sectest.MockedClient{
			MockedClientProperties: sectest.MockedClientProperties{
				ClientID: "test-client",
				AssignedTenantIds: []string{
					"id-tenant-1",
					"id-tenant-2-b",
				},
			},
		}

		user := sectest.MockedAccount{
			MockedAccountDetails: sectest.MockedAccountDetails{
				UserId:          "test-user",
				AssignedTenants: utils.NewStringSet("id-tenant-2", "id-tenant-1-a"),
				DefaultTenant:   "id-tenant-1-a",
			},
		}

		defaultTenantId, tenantIds, e := ResolveClientUserTenants(ctx, user, client)
		g.Expect(e).To(Not(HaveOccurred()))
		g.Expect(defaultTenantId).To(Equal("id-tenant-1-a")) // this client has multiple tenants, so we can't pick a default
		g.Expect(len(tenantIds)).To(Equal(2))
		g.Expect(utils.NewStringSet(tenantIds...)).To(Equal(utils.NewStringSet("id-tenant-1-a", "id-tenant-2-b")))

		user = sectest.MockedAccount{
			MockedAccountDetails: sectest.MockedAccountDetails{
				UserId:          "test-user",
				AssignedTenants: utils.NewStringSet("id-tenant-2", "id-tenant-1-a"),
				DefaultTenant:   "id-tenant-2",
			},
		}

		defaultTenantId, tenantIds, e = ResolveClientUserTenants(ctx, user, client)
		g.Expect(e).To(Not(HaveOccurred()))
		g.Expect(defaultTenantId).To(Equal("")) // this client has multiple tenants, so we can't pick a default
		g.Expect(len(tenantIds)).To(Equal(2))
		g.Expect(utils.NewStringSet(tenantIds...)).To(Equal(utils.NewStringSet("id-tenant-1-a", "id-tenant-2-b")))

		user = sectest.MockedAccount{
			MockedAccountDetails: sectest.MockedAccountDetails{
				UserId:          "test-user",
				AssignedTenants: utils.NewStringSet("id-tenant-1", "id-tenant-2-b"),
				DefaultTenant:   "id-tenant-1",
			},
		}

		defaultTenantId, tenantIds, e = ResolveClientUserTenants(ctx, user, client)
		g.Expect(e).To(Not(HaveOccurred()))
		g.Expect(defaultTenantId).To(Equal("id-tenant-1")) // this client has multiple tenants, so we can't pick a default
		g.Expect(len(tenantIds)).To(Equal(2))
		g.Expect(utils.NewStringSet(tenantIds...)).To(Equal(utils.NewStringSet("id-tenant-1", "id-tenant-2-b")))
	}
}

func ProvideMockedTenancyAccessor() tenancy.Accessor {
	mockTenantAccessor := mocks.NewMockTenancyAccessorUsingStrIds([]mocks.TenancyRelationWithStrId{
		{ParentId: "id-tenant-root", ChildId: "id-tenant-1"},
		{ParentId: "id-tenant-root", ChildId: "id-tenant-2"},
		{ParentId: "id-tenant-1", ChildId: "id-tenant-1-a"},
		{ParentId: "id-tenant-1", ChildId: "id-tenant-1-b"},
		{ParentId: "id-tenant-2", ChildId: "id-tenant-2-a"},
		{ParentId: "id-tenant-2", ChildId: "id-tenant-2-b"},
	}, "id-tenant-root")
	return mockTenantAccessor
}
