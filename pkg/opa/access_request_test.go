package opa

import (
	"context"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
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
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(BundleServerProvider()),
			fx.Invoke(opatestserver.InitializeBundleServer),
		),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestRequestWithPermission(di), "TestRequestWithPermission"),
		test.GomegaSubTest(SubTestRequestWithoutPermission(di), "TestRequestWithoutPermission"),
		test.GomegaSubTest(SubTestRequestWithoutPolicy(di), "TestRequestWithoutPolicy"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestRequestWithPermission(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var e error
		// admin - can read
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		req = MockRequest(ctx, http.MethodGet, "/test/api/get")
		e = AllowRequest(ctx, req, func(opt *RequestQueryOption) {
			opt.Policy = "testservice/allow_api"
		})
		g.Expect(e).To(Succeed(), "API access should be granted")

		// user - can read
		ctx = sectest.ContextWithSecurity(ctx, MemberNonOwnerOptions())
		req = MockRequest(ctx, http.MethodGet, "/test/api/get")
		e = AllowRequest(ctx, req, func(opt *RequestQueryOption) {
			opt.Policy = "testservice/allow_api"
		})
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
		e = AllowRequest(ctx, req, func(opt *RequestQueryOption) {
			opt.Policy = "testservice/allow_api"
		})
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, ErrAccessDenied)).To(BeTrue(), "error should be ErrAccessDenied")
	}
}

func SubTestRequestWithoutPolicy(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var e error
		// user - cannot write
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		req = MockRequest(ctx, http.MethodPost, "/test/api/post")
		e = AllowRequest(ctx, req, func(opt *RequestQueryOption) {
			opt.Policy = "testservice/unknown_policy"
		})
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, ErrAccessDenied)).To(BeTrue(), "error should be ErrAccessDenied")
	}
}

/*************************
	Helpers
 *************************/

func MockRequest(ctx context.Context, method, path string) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, method, path, nil)
	return req
}

