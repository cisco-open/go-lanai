package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/authconfig"
	"go.uber.org/fx"
)

func init() {
	bootstrap.AddOptions(
		fx.Invoke(configureSecurity),
	)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {
	authconfig.Use()
}

func configureSecurity(init security.Registrar, store security.AccountStore) {
	init.Register(&TokenEndpointSecurityConfigurer {})
}
