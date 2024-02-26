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

package security

import (
    "context"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/tenancy"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/mocks"
    "github.com/google/uuid"
    "github.com/onsi/gomega"
    "go.uber.org/fx"
    "testing"
    "time"
)

var (
	MockedRootTenantId = uuid.MustParse("23967dfe-d90f-4e1b-9406-e2df6685f232")
	MockedTenantIdA    = uuid.MustParse("d8423acc-28cb-4209-95d6-089de7fb27ef")
)

// Uncomment this function to generate a new copyist sql file to test against - needed when expected db sql commands change

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type contextTestDI struct {
	fx.In
	TA tenancy.Accessor
}

func provideMockedTenancyAccessor() tenancy.Accessor {
	tenancyRelationship := []mocks.TenancyRelation{
		{Parent: MockedRootTenantId, Child: MockedTenantIdA},
	}
	return mocks.NewMockTenancyAccessor(tenancyRelationship, MockedRootTenantId)
}

func TestContext(t *testing.T) {
	di := &contextTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(tenancy.Module),
		apptest.WithTimeout(time.Minute),
		apptest.WithFxOptions(
			fx.Provide(provideMockedTenancyAccessor),
		),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestHasErrorAccessingTenant(di), "SubTestHasErrorAccessingTenant"),
	)
}

func SubTestHasErrorAccessingTenant(di *contextTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tests := []struct {
			name              string
			tenantId          string
			permission        string
			hasDescendant     bool
			assignedTenantIds utils.StringSet
			expectedErr       error
		}{
			{
				name:          "test invalid tenant id",
				tenantId:      uuid.New().String(),
				permission:    SpecialPermissionAPIAdmin,
				hasDescendant: false,
				expectedErr:   ErrorInvalidTenantId,
			},
			{
				name:              "test has access to all",
				tenantId:          MockedRootTenantId.String(),
				assignedTenantIds: utils.NewStringSet(SpecialTenantIdWildcard),
				hasDescendant:     false,
				expectedErr:       nil,
			},
			{
				name:              "test has access to tenant",
				tenantId:          MockedTenantIdA.String(),
				permission:        SpecialPermissionAPIAdmin,
				hasDescendant:     true,
				assignedTenantIds: utils.NewStringSet(MockedTenantIdA.String()),
				expectedErr:       nil,
			},
			{
				name:              "test no access",
				tenantId:          MockedTenantIdA.String(),
				permission:        SpecialPermissionAPIAdmin,
				hasDescendant:     true,
				assignedTenantIds: utils.NewStringSet(),
				expectedErr:       ErrorTenantAccessDenied,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mockedAuth := &MockedAccountAuth{
					permissions: map[string]interface{}{tt.permission: struct{}{}},
					details: &MockedUserDetails{
						userId:            uuid.New().String(),
						username:          "test user",
						assignedTenantIds: tt.assignedTenantIds,
					},
				}
				//ctx = context.WithValue(ctx, contextKeySecurity, mockedAuth)
				ctx := utils.MakeMutableContext(ctx)
				MustSet(ctx, mockedAuth)
				err := HasErrorAccessingTenant(ctx, tt.tenantId)
				g.Expect(errors.Is(err, tt.expectedErr)).To(gomega.BeTrue())
			})
		}
	}
}
