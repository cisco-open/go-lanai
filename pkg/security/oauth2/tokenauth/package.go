package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "oauth2 resource server",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{},
}

var logger = log.GetNamedLogger("OAuth2Resource")

func init() {
	bootstrap.Register(Module)
}

