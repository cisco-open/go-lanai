package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	session2 "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/utils"
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
}

type TestSecurityConfigurer struct {
	accountStore security.AccountStore
}

func (c *TestSecurityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	ws.Route(route.WithPattern("/api/**")).
		Condition(utils.WithDomain("localhost:8080")).
		With(passwd.New().
			AccountStore(c.accountStore).
			PasswordEncoder(passwd.NewNoopPasswordEncoder()),
		).
		With(basicauth.New())
}


type AnotherSecurityConfigurer struct {
}

func (c *AnotherSecurityConfigurer) Configure(ws security.WebSecurity) {

	// non-DSL style example
	ws.Route(route.WithPattern("/page/**")).
		Condition(utils.WithDomain("localhost:8080"))

	session2.Configure(ws)
	basicauth.Configure(ws)
	passwd.Configure(ws)
}