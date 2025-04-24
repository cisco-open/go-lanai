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

package th_loader

import (
	"context"
	"embed"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/tenancy"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/embedded"
	"github.com/cisco-open/go-lanai/test/mocks"
	"github.com/cisco-open/go-lanai/test/suitetest"
	"github.com/ghodss/yaml"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"io"
	"io/fs"
	"testing"
)

/*************************
	Test Setup
 *************************/

//go:embed testdata/mock_tenants.yml
var TenantsSourceFS embed.FS

func ProvideTestTenantStore() (*TestTenantStore, TenantHierarchyStore) {
	store := &TestTenantStore{
		SourceFS:   TenantsSourceFS,
		SourcePath: "testdata/mock_tenants.yml",
	}
	return store, store
}

type TAOut struct {
	fx.Out
	Accessor tenancy.Accessor `name:"tenancy/accessor"`
}

func ProvideTestTenantAccessor() TAOut {
	return TAOut{
		Accessor: &mocks.MockTenancyAccessor{
			Isloaded: false,
		},
	}
}

func ProvideTestCacheProperties() tenancy.CacheProperties {
	return tenancy.CacheProperties{DbIndex: 0}
}

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
	DB              *gorm.DB `optional:"true"`
	InternalLoader  Loader   `name:"tenant_hierarchy/loader"`
	ClientFactory   redis.ClientFactory
	AppCtx          *bootstrap.ApplicationContext
	TestTenantStore *TestTenantStore
}

func TestLoader(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(redis.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(
				defaultLoader(),
				ProvideTestTenantStore,
				ProvideTestTenantAccessor,
				ProvideTestCacheProperties,
			),
		),
		test.GomegaSubTest(SubTestLoadTenantHierarchy(di), "SubTestLoadTenantHierarchy"),
		test.GomegaSubTest(SubTestLoadTenantHierarchyFailure(di), "SubTestLoadTenantHierarchyFailure"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLoadTenantHierarchy(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := initializeTenantHierarchy(initDi{
			AppCtx:          di.AppCtx,
			EffectiveLoader: di.InternalLoader,
		})
		g.Expect(e).To(Succeed())
	}
}

func SubTestLoadTenantHierarchyFailure(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.TestTenantStore.Reset(nil, "")
		e := initializeTenantHierarchy(initDi{
			AppCtx:          di.AppCtx,
			EffectiveLoader: di.InternalLoader,
		})
		g.Expect(e).To(HaveOccurred())
	}
}

/*************************
	Mocks
 *************************/

type TestData struct {
	Tenants []TestTenant `json:"tenants"`
}

type TestTenant struct {
	ID       string `json:"id"`
	ParentID string `json:"parent-id"`
}

func (t TestTenant) GetId() string {
	return t.ID
}

func (t TestTenant) GetParentId() string {
	return t.ParentID
}

type TestTenantStore struct {
	TestData
	SourceFS   fs.FS
	SourcePath string
}

func (s *TestTenantStore) Reset(srcFS fs.FS, srcPath string) {
	s.Tenants = nil
	s.SourceFS = srcFS
	s.SourcePath = srcPath
}

func (s *TestTenantStore) GetIterator(_ context.Context) (TenantIterator, error) {
	if len(s.SourcePath) == 0 || s.SourceFS == nil {
		return &TestTenantIterator{Tenants: []TestTenant{}}, nil
	}

	if len(s.Tenants) == 0 {
		data, e := fs.ReadFile(s.SourceFS, s.SourcePath)
		if e != nil {
			return nil, fmt.Errorf("unable to load test tenants file: %v", e)
		}
		if e := yaml.Unmarshal(data, &s.TestData); e != nil {
			return nil, fmt.Errorf("unable to parse test tenants file: %v", e)
		}
	}
	return &TestTenantIterator{Tenants: s.Tenants}, nil
}

type TestTenantIterator struct {
	Tenants []TestTenant
}

func (i *TestTenantIterator) Next() bool {
	return len(i.Tenants) != 0
}

func (i *TestTenantIterator) Scan(_ context.Context) (Tenant, error) {
	if len(i.Tenants) == 0 {
		return nil, io.EOF
	}
	defer func() {
		i.Tenants = i.Tenants[1:]
	}()
	return i.Tenants[0], nil
}

func (i *TestTenantIterator) Close() error {
	return nil
}

func (i *TestTenantIterator) Err() error {
	return nil
}
