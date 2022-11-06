package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/web_test/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Tests
 *************************/

func TestMiddlewareRegistration(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestWithMWMapping(&di), "TestWithMWMapping"),
		test.GomegaSubTest(SubTestWithConditionalMW(&di), "TestWithConditionalMW"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestWithMWMapping(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		mw := NewTestMW()
		route := matcher.RouteWithPrefix("/mw")
		mappings := []interface{}{
			middleware.NewBuilder("mw").ApplyTo(route).With(mw).Build(),
			middleware.NewBuilder("mw-web").ApplyTo(route).Use(mw.HandlerFunc()).Build(),
			middleware.NewBuilder("mw-http").ApplyTo(route).Use(mw.HttpHandlerFunc()).Build(),
			middleware.NewBuilder("mw-gin").ApplyTo(route).Use(mw.GinHandlerFunc()).Build(),
		}
		WebInit(ctx, t, g, di,
			registerSuccessEndpoint(http.MethodPut, "/mw/:var"),
			registerSuccessEndpoint(http.MethodPut, "/no-mw/:var"),
			func(reg *web.Registrar) {
				e := reg.Register(mappings)
				g.Expect(e).To(Succeed(), "register controller should success")
			},
		)
		testEndpoint(ctx, t, g, http.MethodPut, "/mw/var-value")
		testEndpoint(ctx, t, g, http.MethodPut, "/no-mw/var-value")
		assertMW(t, g, mw, mwExpectCount(4))
	}
}

func SubTestWithConditionalMW(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		mw := NewTestMW()
		condMW := NewTestMW()
		route := matcher.RouteWithPattern("/mw/**")
		cond := matcher.RequestWithMethods(http.MethodPost)
		mappings := []interface{}{
			middleware.NewBuilder("mw").ApplyTo(route).With(mw).Build(),
			middleware.NewBuilder("mw-cond").ApplyTo(route).WithCondition(cond).With(condMW).Build(),
			middleware.NewBuilder("mw-web").ApplyTo(route).Use(mw.HandlerFunc()).Build(),
			middleware.NewBuilder("mw-web-cond").ApplyTo(route).WithCondition(cond).Use(condMW.HandlerFunc()).Build(),
			middleware.NewBuilder("mw-http").ApplyTo(route).Use(mw.HttpHandlerFunc()).Build(),
			middleware.NewBuilder("mw-http-cond").ApplyTo(route).WithCondition(cond).Use(condMW.HttpHandlerFunc()).Build(),
			middleware.NewBuilder("mw-gin").ApplyTo(route).Use(mw.GinHandlerFunc()).Build(),
			middleware.NewBuilder("mw-gin-cond").ApplyTo(route).WithCondition(cond).Use(condMW.GinHandlerFunc()).Build(),
		}
		WebInit(ctx, t, g, di,
			registerSuccessEndpoint(http.MethodPut, "/mw/:var"),
			registerSuccessEndpoint(http.MethodPost, "/mw/:var"),
			registerSuccessEndpoint(http.MethodPut, "/no-mw/:var"),
			registerSuccessEndpoint(http.MethodPost, "/no-mw/:var"),
			func(reg *web.Registrar) {
				e := reg.Register(mappings)
				g.Expect(e).To(Succeed(), "register controller should success")
			},
		)
		testEndpoint(ctx, t, g, http.MethodPut, "/mw/var-value")
		testEndpoint(ctx, t, g, http.MethodPut, "/no-mw/var-value")
		assertMW(t, g, mw, mwExpectCount(4))
		assertMW(t, g, condMW, mwExpectCount(0))

		mw.Reset()
		testEndpoint(ctx, t, g, http.MethodPost, "/mw/var-value")
		testEndpoint(ctx, t, g, http.MethodPost, "/no-mw/var-value")
		assertMW(t, g, mw, mwExpectCount(4))
		assertMW(t, g, condMW, mwExpectCount(4))
	}
}

/*************************
	Helper
 *************************/

func registerSuccessEndpoint(method, path string) WebInitFunc {
	return func(reg *web.Registrar) {
		reg.MustRegister(rest.New(path).
			Method(method).
			Path(path).
			EndpointFunc(testdata.StructPtr200).
			Build())
	}
}

