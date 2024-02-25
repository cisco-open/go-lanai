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
	"testing"
)

/*************************
	Test
 *************************/

func TestGenericAllow(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		opatest.WithBundles(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestGenericBaseline(), "TestGenericBaseline"),
		test.GomegaSubTest(SubTestAllowGenericWithAuth(), "TestAllowGenericWithAuth"),
		test.GomegaSubTest(SubTestAllowGenericWithoutAuth(), "TestAllowGenericWithoutAuth"),
		test.GomegaSubTest(SubTestGenericInvalidInputCustomizer(di), "TestGenericInvalidInputCustomizer"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestGenericBaseline() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = opa.Allow(ctx, func(q *opa.Query) {
			q.Policy = "baseline/allow"
			q.RawInput = map[string]interface{}{
				"just_data": "data",
			}
		}, opa.SilentQuery())
		g.Expect(e).To(Succeed())
	}
}

func SubTestAllowGenericWithAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = opa.Allow(ctx, opa.QueryWithPolicy("baseline/allow_custom"))
		g.Expect(e).To(HaveOccurred())

		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		e = opa.Allow(ctx, opa.QueryWithPolicy("baseline/allow_custom"))
		g.Expect(e).To(Succeed())
	}
}

func SubTestAllowGenericWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = opa.Allow(ctx,
			opa.QueryWithPolicy("baseline/allow_custom"),
			opa.QueryWithInputCustomizer(func(ctx context.Context, input *opa.Input) error {
				input.ExtraData["allow_no_auth"] = true
				return nil
			}),
		)
		g.Expect(e).To(Succeed())
	}
}

func SubTestGenericInvalidInputCustomizer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		e = opa.Allow(ctx,
			opa.QueryWithPolicy("baseline/allow_custom"),
			opa.QueryWithInputCustomizer(func(ctx context.Context, input *opa.Input) error {
				return errors.New("oops")
			}),
		)
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, opa.ErrInternal)).To(BeTrue(), "error should be ErrInternal")
	}
}
