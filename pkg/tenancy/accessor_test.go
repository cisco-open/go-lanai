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

package tenancy_test

import (
    "context"
    "embed"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/tenancy"
    th_loader "github.com/cisco-open/go-lanai/pkg/tenancy/loader"
    th_modifier "github.com/cisco-open/go-lanai/pkg/tenancy/modifier"
    "github.com/cisco-open/go-lanai/pkg/tenancy/testdata"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/embedded"
    "github.com/google/uuid"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "testing"
)

/*************************
	Setup Test
 *************************/

const (
	TenantRoot = `root`
	TenantA    = `A`
	TenantB    = `B`

	TenantA1 = `A-1`
	TenantA2 = `A-2`
	TenantB1 = `B-1`
	TenantB2 = `B-2`

	TenantA11 = `A-1-1`
	TenantA12 = `A-1-2`
	TenantA21 = `A-2-1`
	TenantA22 = `A-2-2`
	TenantB11 = `B-1-1`
	TenantB12 = `B-1-2`
	TenantB21 = `B-2-1`
	TenantB22 = `B-2-2`
)

//go:embed testdata/mock_tenants.yml
var TenantsSourceFS embed.FS

func ProvideTestTenantStore() (*testdata.TestTenantStore, th_loader.TenantHierarchyStore) {
	store := &testdata.TestTenantStore{
		SourceFS:   TenantsSourceFS,
		SourcePath: "testdata/mock_tenants.yml",
	}
	return store, store
}

/*************************
	Tests
 *************************/

type TestAccessorDI struct {
	fx.In
	AppContext      *bootstrap.ApplicationContext
	TestTenantStore *testdata.TestTenantStore
	Modifier        th_modifier.Modifier
}

func TestTenancyAccessor(t *testing.T) {
	di := TestAccessorDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		embedded.WithRedis(),
		apptest.WithModules(tenancy.Module, th_modifier.Module, th_loader.Module, redis.Module),
		apptest.WithFxOptions(
			fx.Provide(ProvideTestTenantStore),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestTraceBack(&di), "TestTraceBack"),
		test.GomegaSubTest(SubTestTraceForward(&di), "TestTraceForward"),
		test.GomegaSubTest(SubTestAnyHas(&di), "TestAnyHas"),
		test.GomegaSubTest(SubTestTenancyModification(&di), "TestTenancyModification"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestTraceBack(di *TestAccessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		loaded := tenancy.IsLoaded(ctx)
		g.Expect(loaded).To(BeTrue(), "tenancy should be loaded")

		var e error
		var v string
		v, e = tenancy.GetRoot(ctx)
		g.Expect(e).To(Succeed(), "GetRoot should not fail")
		g.Expect(v).To(Equal(IDOf(di, TenantRoot)), "root should be correct")

		v, e = tenancy.GetParent(ctx, IDOf(di, TenantRoot))
		g.Expect(e).To(Succeed(), "GetParent should not fail")
		g.Expect(v).To(BeEmpty(), "root should not have parent")

		v, e = tenancy.GetParent(ctx, IDOf(di, TenantA12))
		g.Expect(e).To(Succeed(), "GetParent should not fail")
		g.Expect(v).To(Equal(IDOf(di, TenantA1)), "parent should be correct")

		var multiV []string
		multiV, e = tenancy.GetAncestors(ctx, IDOf(di, TenantRoot))
		g.Expect(e).To(Succeed(), "GetAncestors should not fail")
		g.Expect(multiV).To(BeEmpty(), "ancestors of root should be correct")

		multiV, e = tenancy.GetAncestors(ctx, IDOf(di, TenantB22))
		g.Expect(e).To(Succeed(), "GetAncestors should not fail")
		g.Expect(multiV).To(HaveLen(3), "ancestors should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB2)), "ancestors should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB)), "ancestors should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantRoot)), "ancestors should be correct")

		var uuids []uuid.UUID
		uuids, e = tenancy.GetTenancyPath(ctx, IDOf(di, TenantRoot))
		g.Expect(e).To(Succeed(), "GetTenancyPath should not fail")
		g.Expect(uuids).To(HaveLen(1), "tenant path of root should be correct")
		g.Expect(uuids[0]).To(Equal(UUIDOf(di, TenantRoot)), "tenant path of root should be correct")

		uuids, e = tenancy.GetTenancyPath(ctx, IDOf(di, TenantB21))
		g.Expect(e).To(Succeed(), "GetTenancyPath should not fail")
		g.Expect(uuids).To(HaveLen(4), "tenant path should be correct")
		g.Expect(uuids[0]).To(Equal(UUIDOf(di, TenantRoot)), "tenant path should be correct")
		g.Expect(uuids[1]).To(Equal(UUIDOf(di, TenantB)), "tenant path should be correct")
		g.Expect(uuids[2]).To(Equal(UUIDOf(di, TenantB2)), "tenant path should be correct")
		g.Expect(uuids[3]).To(Equal(UUIDOf(di, TenantB21)), "tenant path should be correct")
	}
}

func SubTestTraceForward(di *TestAccessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		loaded := tenancy.IsLoaded(ctx)
		g.Expect(loaded).To(BeTrue(), "tenancy should be loaded")

		var e error
		var multiV []string
		multiV, e = tenancy.GetChildren(ctx, IDOf(di, TenantA21))
		g.Expect(e).To(Succeed(), "GetChildren should not fail")
		g.Expect(multiV).To(BeEmpty(), "children of leaf tenant should be correct")

		multiV, e = tenancy.GetChildren(ctx, IDOf(di, TenantB1))
		g.Expect(e).To(Succeed(), "GetChildren should not fail")
		g.Expect(multiV).To(HaveLen(2), "children should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB11)), "children should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB12)), "children should be correct")

		multiV, e = tenancy.GetDescendants(ctx, IDOf(di, TenantA22))
		g.Expect(e).To(Succeed(), "GetDescendants should not fail")
		g.Expect(multiV).To(BeEmpty(), "descendants of leaf tenant should be correct")

		multiV, e = tenancy.GetDescendants(ctx, IDOf(di, TenantB))
		g.Expect(e).To(Succeed(), "GetDescendants should not fail")
		g.Expect(multiV).To(HaveLen(6), "descendants should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB1)), "descendants should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB2)), "descendants should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB11)), "descendants should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB12)), "descendants should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB21)), "descendants should be correct")
		g.Expect(multiV).To(ContainElement(IDOf(di, TenantB22)), "descendants should be correct")
	}
}

func SubTestAnyHas(di *TestAccessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		loaded := tenancy.IsLoaded(ctx)
		g.Expect(loaded).To(BeTrue(), "tenancy should be loaded")

		var ids utils.StringSet
		var rs bool
		ids = utils.NewStringSet(IDOf(di, TenantA1), IDOf(di, TenantB1))
		rs = tenancy.AnyHasDescendant(ctx, ids, IDOf(di, TenantA))
		g.Expect(rs).To(BeFalse(), "AnyHasDescendant should be correct")

		ids = utils.NewStringSet(IDOf(di, TenantA1), IDOf(di, TenantB1))
		rs = tenancy.AnyHasDescendant(ctx, ids, IDOf(di, TenantA21))
		g.Expect(rs).To(BeFalse(), "AnyHasDescendant should be correct")

		// TODO is "self" count as descendent?
		ids = utils.NewStringSet(IDOf(di, TenantA1), IDOf(di, TenantB1))
		rs = tenancy.AnyHasDescendant(ctx, ids, IDOf(di, TenantA1))
		g.Expect(rs).To(BeTrue(), "AnyHasDescendant should be correct")

		ids = utils.NewStringSet(IDOf(di, TenantA1), IDOf(di, TenantB1))
		rs = tenancy.AnyHasDescendant(ctx, ids, IDOf(di, TenantB12))
		g.Expect(rs).To(BeTrue(), "AnyHasDescendant should be correct")

		rs = tenancy.AnyHasDescendant(ctx, nil, IDOf(di, TenantB12))
		g.Expect(rs).To(BeFalse(), "AnyHasDescendant should be correct")

		rs = tenancy.AnyHasDescendant(ctx, ids, "")
		g.Expect(rs).To(BeFalse(), "AnyHasDescendant should be correct")
	}
}

func SubTestTenancyModification(di *TestAccessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		loaded := tenancy.IsLoaded(ctx)
		g.Expect(loaded).To(BeTrue(), "tenancy should be loaded")

		var e error
		var v string
		var multiV []string
		// remove
		e = th_modifier.RemoveTenant(ctx, IDOf(di, TenantA2))
		g.Expect(e).To(HaveOccurred(), "removing non-leaf tenant should fail")

		e = th_modifier.RemoveTenant(ctx, IDOf(di, TenantA21))
		g.Expect(e).To(Succeed(), "removing leaf tenant should fail")
		v, e = tenancy.GetParent(ctx, IDOf(di, TenantA21))
		g.Expect(e).To(Succeed(), "GetParent of removed tenant should not fail")
		g.Expect(v).To(BeEmpty(), "GetParent of removed tenant should be correct")
		multiV, e = tenancy.GetChildren(ctx, IDOf(di, TenantA2))
		g.Expect(e).To(Succeed(), "GetChildren of removed tenant should not fail")
		g.Expect(multiV).To(HaveLen(1), "GetChildren of removed tenant should be correct")

		e = th_modifier.RemoveTenant(ctx, IDOf(di, TenantA21))
		g.Expect(e).To(HaveOccurred(), "removing non-existing tenant should fail")

		// add
		e = th_modifier.AddTenant(ctx, IDOf(di, TenantA21), IDOf(di, TenantB2))
		g.Expect(e).To(Succeed(), "add tenant should not fail")
		v, e = tenancy.GetParent(ctx, IDOf(di, TenantA21))
		g.Expect(e).To(Succeed(), "GetParent of added tenant should not fail")
		g.Expect(v).To(Equal(IDOf(di, TenantB2)), "GetParent of added tenant should be correct")
		multiV, e = tenancy.GetChildren(ctx, IDOf(di, TenantB2))
		g.Expect(e).To(Succeed(), "GetChildren of added tenant's parent should not fail")
		g.Expect(multiV).To(HaveLen(3), "GetChildren of added tenant's parent should be correct")

		// TODO following test should not fail, maybe it's a bug?
		e = di.Modifier.AddTenant(ctx, IDOf(di, TenantA21), IDOf(di, TenantA2))
		g.Expect(e).To(HaveOccurred(), "relocating leaf tenant should fail")
	}
}

/*************************
	Helpers
 *************************/

func IDOf(di *TestAccessorDI, tenant string) string {
	return di.TestTenantStore.IDof(tenant)
}

func UUIDOf(di *TestAccessorDI, tenant string) uuid.UUID {
	id, e := uuid.Parse(di.TestTenantStore.IDof(tenant))
	if e != nil {
		return uuid.Nil
	}
	return id
}
