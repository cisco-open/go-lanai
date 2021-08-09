package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

//var logger = log.New("OAuth2.Token")

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "oauth2 resource server",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{},
}

func init() {
	bootstrap.Register(Module)
}

