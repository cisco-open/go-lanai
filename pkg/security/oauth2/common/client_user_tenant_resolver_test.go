// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/tenancy"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/mocks"
	"github.com/cisco-open/go-lanai/test/sectest"
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
		g.Expect(defaultTenantId).To(Equal("id-tenant-1-a")) // default tenant is still within the set of resolved tenants, so keep it.
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
		g.Expect(defaultTenantId).To(Equal("")) //default tenant is no longer part of the assigned tenants, so it's empty.
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
		g.Expect(defaultTenantId).To(Equal("id-tenant-1")) //default tenant is still part of the result tenants, so keep it.
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
