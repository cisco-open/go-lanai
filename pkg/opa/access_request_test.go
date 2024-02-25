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
	"net/http"
	"testing"
)

/*************************
	Test Setup
 *************************/

/*************************
	Test
 *************************/

func TestAllowRequest(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		opatest.WithBundles(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestRequestBaseline(di), "TestRequestBaseline"),
		test.GomegaSubTest(SubTestRequestWithPermission(di), "TestRequestWithPermission"),
		test.GomegaSubTest(SubTestRequestWithoutPermission(di), "TestRequestWithoutPermission"),
		test.GomegaSubTest(SubTestRequestWithoutPolicy(di), "TestRequestWithoutPolicy"),
		test.GomegaSubTest(SubTestRequestInvalidInputCustomizer(di), "TestRequestInvalidInputCustomizer"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestRequestBaseline(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var e error
		req = MockRequest(ctx, http.MethodGet, "/doesnt/matter")
		e = opa.AllowRequest(ctx, req, opa.RequestQueryWithPolicy("baseline/allow"), func(opt *opa.RequestQuery) {
			opt.RawInput = map[string]interface{}{
				"just_data": "data",
			}
		}, opa.SilentRequestQuery())
		g.Expect(e).To(Succeed())
	}
}

func SubTestRequestWithPermission(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var e error
		// admin - can read
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		req = MockRequest(ctx, http.MethodGet, "/test/api/get")
		e = opa.AllowRequest(ctx, req, opa.RequestQueryWithPolicy("testservice/allow_api"))
		g.Expect(e).To(Succeed(), "API access should be granted")

		// user - can read
		ctx = sectest.ContextWithSecurity(ctx, MemberNonOwnerOptions())
		req = MockRequest(ctx, http.MethodGet, "/test/api/get")
		e = opa.AllowRequest(ctx, req, opa.RequestQueryWithPolicy("testservice/allow_api"))
		g.Expect(e).To(Succeed(), "API access should be granted")
	}
}

func SubTestRequestWithoutPermission(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var e error
		// user - cannot write
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		req = MockRequest(ctx, http.MethodPost, "/test/api/post")
		e = opa.AllowRequest(ctx, req, func(opt *opa.RequestQuery) {
			opt.Policy = "testservice/allow_api"
		})
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, opa.ErrAccessDenied)).To(BeTrue(), "error should be ErrAccessDenied")
	}
}

func SubTestRequestWithoutPolicy(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var e error
		// user - cannot write
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		req = MockRequest(ctx, http.MethodPost, "/test/api/post")
		e = opa.AllowRequest(ctx, req, opa.RequestQueryWithPolicy("testservice/unknown_policy"))
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, opa.ErrAccessDenied)).To(BeTrue(), "error should be ErrAccessDenied")
	}
}

func SubTestRequestInvalidInputCustomizer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var e error
		// user - cannot write
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		req = MockRequest(ctx, http.MethodPost, "/test/api/post")
		e = opa.AllowRequest(ctx, req, func(opt *opa.RequestQuery) {
			opt.InputCustomizers = append(opt.InputCustomizers, opa.InputCustomizerFunc(func(ctx context.Context, input *opa.Input) error {
				return errors.New("oops")
			}))
		})
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, opa.ErrInternal)).To(BeTrue(), "error should be ErrInternal")
	}
}

/*************************
	Helpers
 *************************/

func MockRequest(ctx context.Context, method, path string) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, method, path, nil)
	return req
}

