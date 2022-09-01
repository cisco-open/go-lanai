// Examples. The reason this file exists is to work around an issue existed since go 1.3:
// https://github.com/golang/go/issues/8279
// Note: this issue has been fixed in 1.17

package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"fmt"
	"net/http"
)

/*************************
	Examples Setup
 *************************/

type TestTarget struct{}

func (t *TestTarget) DoSomethingWithinSecurityScope(ctx context.Context) error {
	e := scope.Do(ctx, func(scopedCtx context.Context) {
		// scopedCtx contains switched security context
		// do something with scopedCtx...
		_ = t.DoSomethingRequiringSecurity(scopedCtx)
	}, scope.UseSystemAccount())
	return e
}

func (t *TestTarget) DoSomethingRequiringSecurity(ctx context.Context) error {
	auth := security.Get(ctx)
	if !security.IsFullyAuthenticated(auth) {
		return fmt.Errorf("not authenticated")
	}
	return nil
}

const (
	TestSecuredURL    = "/api/v1/secured"
	TestEntryPointURL = "/login"
)

type TestController struct{}

func registerTestController(reg *web.Registrar) {
	reg.MustRegister(&TestController{})
}

func (c *TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("secured-get").Get(TestSecuredURL).
			EndpointFunc(c.Secured).Build(),
		rest.New("secured-post").Post(TestSecuredURL).
			EndpointFunc(c.Secured).Build(),
	}
}

func (c *TestController) Secured(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		"Message": "Yes",
	}, nil
}

type TestSecConfigurer struct{}

func (c *TestSecConfigurer) Configure(ws security.WebSecurity) {
	ws.Route(matcher.RouteWithPattern("/api/**")).
		With(
			basicauth.New().EntryPoint(redirect.NewRedirectWithRelativePath(TestEntryPointURL, false)),
		).
		With(access.New().Request(matcher.AnyRequest()).Authenticated()).
		With(errorhandling.New())
}

func registerTestSecurity(registrar security.Registrar) {
	cfg := TestSecConfigurer{}
	registrar.Register(&cfg)
}