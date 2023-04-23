package swagger

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/embedded"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		embedded.Redis(),
	)
}

func TestSwaggerDocSecurityDisabledWithMockedServer(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(
			resserver.Module,
			redis.Module,
			access.Module,
			errorhandling.Module,
		),
		apptest.WithProperties("swagger.security.secure-docs=false", "swagger.spec: testdata/api-docs-v3.yml"),
		apptest.WithFxOptions(
			fx.Provide(
				NewResServerConfigurer,
				bindSwaggerProperties,
			),
			fx.Invoke(
				initialize,
				configureSecurity,
			),
		),
		test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *gomega.WithT) {
			var req *http.Request
			var resp *http.Response
			uri := fmt.Sprintf("http://cisco.com/test/v3/api-docs")
			req = webtest.NewRequest(ctx, http.MethodGet, uri, nil, func(req *http.Request) {
				req.Header.Add("content-type", "application/json")
			})
			resp = webtest.MustExec(ctx, req).Response
			fmt.Printf("%v\n", resp)
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}, "TestSwaggerDocApiSecurityDisabled"),
	)
}

func TestSwaggerDocSecurityEnabledWithMockedServer(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(
			resserver.Module,
			redis.Module,
			access.Module,
			errorhandling.Module,
		),
		apptest.WithProperties("swagger.security.secure-docs=true", "swagger.spec: testdata/api-docs-v3.yml"),
		apptest.WithFxOptions(
			fx.Provide(
				NewResServerConfigurer,
				bindSwaggerProperties,
			),
			fx.Invoke(
				initialize,
				configureSecurity,
			),
		),
		test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *gomega.WithT) {
			var req *http.Request
			var resp *http.Response
			uri := fmt.Sprintf("http://cisco.com/test/v3/api-docs")
			req = webtest.NewRequest(ctx, http.MethodGet, uri, nil, func(req *http.Request) {
				req.Header.Add("content-type", "application/json")
			})
			resp = webtest.MustExec(ctx, req).Response
			fmt.Printf("%v\n", resp)
			g.Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		}, "TestSwaggerDocApiSecurityEnabled"),
	)
}

func NewResServerConfigurer() resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		//do nothing
	}
}
