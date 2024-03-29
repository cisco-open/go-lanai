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

package opa_test

import (
	"context"
	"errors"
	"github.com/cisco-open/go-lanai/pkg/opa"
	opatest "github.com/cisco-open/go-lanai/pkg/opa/test"
	. "github.com/cisco-open/go-lanai/pkg/opa/testdata"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test
 *************************/

type testDI struct {
	fx.In
}

func TestAllowResource(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		opatest.WithBundles(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestResourceBaseline(di), "TestResourceBaseline"),
		test.GomegaSubTest(SubTestMemberAdmin(di), "TestMemberAdmin"),
		test.GomegaSubTest(SubTestMemberOwner(di), "TestMemberOwner"),
		test.GomegaSubTest(SubTestMemberNonOwner(di), "TestMemberNonOwner"),
		test.GomegaSubTest(SubTestNonMember(di), "TestNonMember"),
		test.GomegaSubTest(SubTestSharedUser(di), "TestSharedUser"),
		test.GomegaSubTest(SubTestResourceWithoutPolicy(di), "TestResourceWithoutPolicy"),
		test.GomegaSubTest(SubTestResourceInvalidInputCustomizer(di), "TestResourceInvalidInputCustomizer"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestResourceBaseline(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = opa.AllowResource(ctx, "doesn't matter", "whatever", func(res *opa.ResourceQuery) {
			res.Policy = "baseline/allow"
			res.RawInput = map[string]interface{}{
				"just_data": "data",
			}
		}, opa.SilentResourceQuery())
		g.Expect(e).To(Succeed())
	}
}

func SubTestMemberAdmin(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		// member admin - can read
		e = opa.AllowResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.ExtraData["debug"] = "test"
		})
		g.Expect(e).To(Succeed())
		// member admin - can write
		e = opa.AllowResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())
	}
}

func SubTestMemberOwner(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// owner - can read
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		// member user - can read
		e = opa.AllowResource(ctx, "poc", opa.OpRead, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())

		// owner - can write
		e = opa.AllowResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())
	}
}

func SubTestMemberNonOwner(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// member user
		ctx = sectest.ContextWithSecurity(ctx, MemberNonOwnerOptions())
		// member user - can read
		e = opa.AllowResource(ctx, "poc", opa.OpRead, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(Succeed())

		// member user - cannot write
		e = opa.AllowResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(HaveOccurred())
	}
}

func SubTestNonMember(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// non-member admin - can't read
		ctx = sectest.ContextWithSecurity(ctx, NonMemberAdminOptions())
		e = opa.AllowResource(ctx, "poc", opa.OpRead, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
		})
		g.Expect(e).To(HaveOccurred())
	}
}

func SubTestSharedUser(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, MemberNonOwnerOptions())
		// non-member user but shared - cannot write if not allowed
		e = opa.AllowResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.Sharing = map[string][]opa.ResourceOperation{
				AnotherUserId: {"read"},
			}
		})
		g.Expect(e).To(HaveOccurred())

		// non-member user but shared - can write if allowed
		e = opa.AllowResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.Sharing = map[string][]opa.ResourceOperation{
				AnotherUserId: {"read", "write"},
			}
		})
		g.Expect(e).To(Succeed())
	}
}

func SubTestResourceWithoutPolicy(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		// member user - can read
		e = opa.AllowResource(ctx, "poc", opa.OpRead, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.Policy = "poc/unknown_policy"
		})
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, opa.ErrAccessDenied)).To(BeTrue(), "error should be ErrAccessDenied")
	}
}

func SubTestResourceInvalidInputCustomizer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		// member user - can read
		e = opa.AllowResource(ctx, "poc", opa.OpRead, func(res *opa.ResourceQuery) {
			res.TenantID = TenantId
			res.OwnerID = OwnerUserId
			res.TenantPath = []string{RootTenantId, TenantId}
			res.InputCustomizers = append(res.InputCustomizers, opa.InputCustomizerFunc(func(ctx context.Context, input *opa.Input) error {
				return errors.New("oops")
			}))
		})
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, opa.ErrInternal)).To(BeTrue(), "error should be ErrInternal")
	}
}