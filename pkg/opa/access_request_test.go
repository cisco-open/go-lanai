package opa_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"errors"
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

