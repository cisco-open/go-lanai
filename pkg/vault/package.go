package vault

import (
	"context"
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
	Options: []fx.Option{
		fx.Invoke(setupRenewal),
	},
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

func setupRenewal(lc fx.Lifecycle, conn *Connection) {
	renewer, err := conn.GetClientTokenRenewer()

	if err != nil {
		panic("cannot create renewer for vault token")
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go renewer.Renew() //r.Renew() starts a blocking process to periodically renew the token. Therefore we run it as a go routine
			go conn.monitorRenew(renewer, "vault client token") //this starts a background process to log the renewal events. These two go routine exits when the renewer is stopped
			return nil
		},
		OnStop: func(ctx context.Context) error {
			renewer.Stop()
			return nil
		},
	})
}