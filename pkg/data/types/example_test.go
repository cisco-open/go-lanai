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

package types

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/tenancy"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/dbtest"
    "github.com/cisco-open/go-lanai/test/sectest"
    "github.com/google/uuid"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "gorm.io/gorm"
    "testing"
)

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type exampleDI struct {
	fx.In
	MockAccessor tenancy.Accessor
	DB           *gorm.DB
}

func TestGormModel(t *testing.T) {
	di := &exampleDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(tenancy.Module),
		apptest.WithFxOptions(
			fx.Provide(provideMockedTenancyAccessor),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupWithMockedSecurity(di)),
		test.SubTestTeardown(TeardownWithTruncateTable(di)),
		test.GomegaSubTest(SubExampleModelCreateWithTenancy(di), "CreateModelWithTenancy"),
		test.GomegaSubTest(SubExampleModelUpdateWithTenancy(di), "UpdateModelWithTenancy"),
		test.GomegaSubTest(SubExampleModelDeleteWithTenancy(di), "DeleteModelWithTenancy"),
	)
}

func SetupWithMockedSecurity(di *exampleDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		_ = di.DB.Exec(exampleTableSQL)
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.Username = "any-username"
			d.UserId = "any-user-id"
			d.TenantExternalId = "any-tenant-ext-id"
			d.TenantId = MockedTenantIdA.String()
			d.Tenants = utils.NewStringSet(MockedTenantIdA.String())
			d.Permissions = utils.NewStringSet()
		}))
		return ctx, nil
	}
}

func TeardownWithTruncateTable(di *exampleDI) test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		return di.DB.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" RESTRICT`, ExampleModel{}.TableName())).Error
	}
}

func SubExampleModelCreateWithTenancy(di *exampleDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// Note: Current user should have access to [MockedTenantIdA, MockedTenantIdA1, MockedTenantIdA2],
		// 		 and have no access to [MockedRootTenantId, MockedTenantIdB, MockedTenantIdB1, MockedTenantIdB2].

		var m *ExampleModel
		var r *gorm.DB

		// with access
		m = &ExampleModel{
			ID:         uuid.New(),
			TenantName: "Tenant A-1",
			Value:      "Any",
			Tenancy: Tenancy{
				TenantID: MockedTenantIdA1,
			},
		}
		r = di.DB.WithContext(ctx).Create(m)
		g.Expect(r.Error).To(Succeed())
		g.Expect(m.TenantPath).To(HaveLen(3), "TenantPath should be populated")

		// without access
		m = &ExampleModel{
			ID:         uuid.New(),
			TenantName: "Tenant B-1",
			Value:      "Any",
			Tenancy: Tenancy{
				TenantID: MockedTenantIdB1,
			},
		}
		r = di.DB.WithContext(ctx).Create(m)
		g.Expect(r.Error).To(Not(Succeed()), "No access to B-1")
	}
}

func SubExampleModelUpdateWithTenancy(di *exampleDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// Note: Current user should have access to [MockedTenantIdA, MockedTenantIdA1, MockedTenantIdA2],
		// 		 and have no access to [MockedRootTenantId, MockedTenantIdB, MockedTenantIdB1, MockedTenantIdB2].
		var m *ExampleModel
		var r *gorm.DB
		// create some records (using access all tenants)
		rootAccessCtx := mockAccessAllTenants(ctx)
		m1 := newExampleModel(MockedTenantIdA1)
		_ = di.DB.WithContext(rootAccessCtx).Create(m1)
		m2 := newExampleModel(MockedTenantIdB1)
		_ = di.DB.WithContext(rootAccessCtx).Create(m2)

		// works with Save
		m = &ExampleModel{}
		*m = *m1
		m.TenantID = MockedTenantIdA // move to A
		r = di.DB.WithContext(ctx).Save(m)
		g.Expect(r.Error).To(Succeed())
		g.Expect(m.TenantPath).To(HaveLen(2), "TenantPath should be populated")

		m.TenantID = MockedTenantIdB // move to B
		r = di.DB.WithContext(ctx).Save(m)
		g.Expect(r.Error).To(Not(Succeed()), "No access to B")

		// Works with Updates by ID
		r = di.DB.WithContext(ctx).
			Model(&ExampleModel{ID: m1.ID}).
			Updates(map[string]interface{}{"TenantID": MockedTenantIdA2})
		g.Expect(r.Error).To(Succeed())
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1))

		r = di.DB.WithContext(ctx).
			Model(&ExampleModel{ID: m2.ID}).
			Updates(map[string]interface{}{"TenantID": MockedTenantIdA2})
		g.Expect(r.Error).To(Succeed())
		g.Expect(r.RowsAffected).To(BeEquivalentTo(0),
			"in case of ORIGINAL tenant is not accessible, Updates succeed with 0 rows affected")

		r = di.DB.WithContext(ctx).
			Model(&ExampleModel{ID: m1.ID}).
			Updates(map[string]interface{}{"TenantID": MockedTenantIdB1})
		g.Expect(r.Error).To(Not(Succeed()), "in case of TARGET tenant is not accessible, Updates fails with error")

		r = di.DB.WithContext(ctx).
			Model(&ExampleModel{ID: m1.ID}).
			Updates(&ExampleModel{Tenancy: Tenancy{TenantID: MockedTenantIdA}})
		g.Expect(r.Error).To(Succeed(), "also works with *ExampleModel as update target")
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1))

		// Works with Updates by WHERE clause
		r = di.DB.WithContext(ctx).
			Model(&ExampleModel{}).
			Where(&ExampleModel{ID: m1.ID}).Or(&ExampleModel{ID: m2.ID}).
			Updates(&ExampleModel{Tenancy: Tenancy{TenantID: MockedTenantIdA2}})
		g.Expect(r.Error).To(Succeed())
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1), "2 matched record, but only the accessible record got updated")
	}
}

func SubExampleModelDeleteWithTenancy(di *exampleDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// Note: Current user should have access to [MockedTenantIdA, MockedTenantIdA1, MockedTenantIdA2],
		// 		 and have no access to [MockedRootTenantId, MockedTenantIdB, MockedTenantIdB1, MockedTenantIdB2].
		var m *ExampleModel
		var r *gorm.DB

		// create some records (using access all tenants)
		rootAccessCtx := mockAccessAllTenants(ctx)
		m1 := newExampleModel(MockedTenantIdA1)
		_ = di.DB.WithContext(rootAccessCtx).Create(m1)
		m2 := newExampleModel(MockedTenantIdB1)
		_ = di.DB.WithContext(rootAccessCtx).Create(m2)
		for _, tid := range []uuid.UUID{MockedTenantIdA, MockedTenantIdA2, MockedTenantIdB, MockedTenantIdB2} {
			m = newExampleModel(tid)
			_ = di.DB.WithContext(rootAccessCtx).Create(m)
		}

		// Works with Delete by ID
		r = di.DB.WithContext(ctx).
			Delete(&ExampleModel{ID: m1.ID})
		g.Expect(r.Error).To(Succeed())
		g.Expect(r.RowsAffected).To(BeEquivalentTo(1))

		r = di.DB.WithContext(ctx).
			Delete(&ExampleModel{ID: m2.ID})
		g.Expect(r.Error).To(Succeed())
		g.Expect(r.RowsAffected).To(BeEquivalentTo(0),
			"to-be-deleted model is not accessible, Delete succeed with 0 rows affected")

		// Works with Updates by WHERE clause
		r = di.DB.WithContext(ctx).
			Where(&ExampleModel{Value: "Any"}).
			Delete(&ExampleModel{})
		g.Expect(r.Error).To(Succeed())
		g.Expect(r.RowsAffected).To(BeEquivalentTo(2), "5 matched record, but only the accessible records got deleted")
	}
}

/*************************
	Helpers
 *************************/

func newExampleModel(tenantIO uuid.UUID) *ExampleModel {
	return &ExampleModel{
		ID:    uuid.New(),
		Value: "Any",
		Tenancy: Tenancy{
			TenantID: tenantIO,
		},
	}
}

func mockAccessAllTenants(ctx context.Context) context.Context {
	return sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "any-username"
		d.UserId = "any-user-id"
		d.TenantExternalId = "any-tenant-ext-id"
		d.TenantId = MockedRootTenantId.String()
		d.Tenants = utils.NewStringSet(MockedRootTenantId.String())
	}))
}

/*************************
	Mocks
 *************************/

type ExampleModel struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	TenantName string
	Value      string
	Tenancy
	Audit
	SoftDelete
}

func (ExampleModel) TableName() string {
	return "model_example"
}

func (t *ExampleModel) BeforeUpdate(tx *gorm.DB) error {
	if security.HasPermissions(security.Get(tx.Statement.Context), specialPermissionSkipTenancyCheck) {
		t.SkipTenancyCheck(tx)
	}
	return t.Tenancy.BeforeUpdate(tx)
}

const exampleTableSQL = `
CREATE TABLE IF NOT EXISTS public.model_example (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"tenant_name" STRING NOT NULL,
	"value" STRING NOT NULL,
	tenant_id UUID NULL,
	tenant_path UUID[] NULL,
	created_at TIMESTAMPTZ NULL,
	updated_at TIMESTAMPTZ NULL,
	created_by UUID NULL,
	updated_by UUID NULL,
	deleted_at TIMESTAMPTZ NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	INVERTED INDEX idx_tenant_path (tenant_path),
	INDEX idx_tenant_name (tenant_name ASC),
	FAMILY "primary" (id, tenant_name, value, tenant_id, tenant_path, created_at, updated_at, created_by, updated_by, deleted_at)
);`
