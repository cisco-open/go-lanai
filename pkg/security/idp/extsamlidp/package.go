package extsamlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	samlsp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/sp"
	"go.uber.org/fx"
)

//var logger = log.New("SEC.SAML")

var Module = &bootstrap.Module{
	Name: "SAML IDP",
	Precedence: security.MaxSecurityPrecedence - 100,
	Options: []fx.Option {
		fx.Provide(BindSamlAuthProperties),
	},
}

func Use() {
	samlsp.Use() // samllogin enables External SAML IDP support
	bootstrap.Register(Module)
}

