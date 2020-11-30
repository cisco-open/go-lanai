package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/route"
	"go.uber.org/fx"
)

func init() {
	bootstrap.AddOptions(fx.Invoke(configureSecurity))
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func configureSecurity(init security.Registrar, store security.AccountStore) {
	init.Register(&TestSecurityConfigurer {
		accountStore: store,
	})

	init.Register(&AnotherSecurityConfigurer { })
	init.Register(&ErrorPageSecurityConfigurer{})
}

type TestSecurityConfigurer struct {
	accountStore security.AccountStore
}

func (c *TestSecurityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(route.WithPattern("/api/**")).
		Condition(matcher.RequestWithHost("localhost:8080")).
		With(passwd.New().
			AccountStore(c.accountStore).
			PasswordEncoder(passwd.NewNoopPasswordEncoder()),
		).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(basicauth.New()).
		With(errorhandling.New())
}


type AnotherSecurityConfigurer struct {
}

func (c *AnotherSecurityConfigurer) Configure(ws security.WebSecurity) {

	// For Page
	handler := errorhandling.NewRedirectWithRelativePath("/error")

	ws.Route(route.WithPattern("/page/**")).
		Condition(matcher.RequestWithHost("localhost:8080")).
		With(basicauth.New()).
		With(session.New()).
		With(passwd.New()).
		With(access.New().
			Request(matcher.RequestWithPattern("/**/page/public")).PermitAll().
			Request(matcher.AnyRequest()).HasPermissions("welcomed"),
		).
		With(errorhandling.New().
			AuthenticationEntryPoint(handler).
			AccessDeniedHandler(handler),
		)
}

type ErrorPageSecurityConfigurer struct {
}

func (c *ErrorPageSecurityConfigurer) Configure(ws security.WebSecurity) {

	ws.Route(route.WithPattern("/error")).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).PermitAll(),
		)
}