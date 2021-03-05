package vault

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

//TODO: register health
//TODO: tracing instrumentation?

var logger = log.New("vault")

var Module = &bootstrap.Module {
	Name: "vault",
	Precedence: bootstrap.VaultPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(newConnectionProperties, newClientAuthentication, NewConnection),
	},
	/*
	Options: []fx.Option{
		fx.Invoke(),
	},
	 */
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func newConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) *ConnectionProperties {
	c := &ConnectionProperties{
		Authentication: TokenAuthentication,
	}
	bootstrapConfig.Bind(c, PropertyPrefix)
	return c
}

func newClientAuthentication(p *ConnectionProperties) ClientAuthentication {
	var clientAuthentication ClientAuthentication
	switch p.Authentication {
	case TokenAuthentication:
		clientAuthentication = TokenClientAuthentication(p.Token)
	default:
		clientAuthentication = TokenClientAuthentication(p.Token)
	}
	return clientAuthentication
}