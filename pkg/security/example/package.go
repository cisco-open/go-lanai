package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	basic "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	session "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/init"
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

	ws.ApplyTo(route.WithPattern("/api/**"))

	passwd.Configure(ws).
		AccountStore(c.accountStore).
		PasswordEncoder(passwd.NewNoopPasswordEncoder())
	basic.Configure(ws)
}


type AnotherSecurityConfigurer struct {
}

func (c *AnotherSecurityConfigurer) Configure(ws security.WebSecurity) {

	ws.ApplyTo(route.WithPattern("/page/**"))

	session.Configure(ws)
	basic.Configure(ws)
	passwd.Configure(ws)
}