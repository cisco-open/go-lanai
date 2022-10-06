package extsamlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin"
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
	samllogin.Use() // samllogin enables External SAML IDP support
	bootstrap.Register(Module)
}

