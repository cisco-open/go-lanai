package vault

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"go.uber.org/fx"
)

var logger = log.New("vault")

var Module = &bootstrap.Module {
	Name: "vault",
	Precedence: bootstrap.VaultPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(newConnectionProperties, vault.NewClient),
	},
	Options: []fx.Option{
		fx.Invoke(setupRenewal, registerHealth),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func newConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) *vault.ConnectionProperties {
	c := &vault.ConnectionProperties{
		Authentication: vault.Token,
	}
	if e := bootstrapConfig.Bind(c, vault.PropertyPrefix); e != nil {
		panic(e)
	}
	return c
}

func newClient(p *vault.ConnectionProperties) *vault.Client {
	c, err := vault.NewClient(p)
	if err != nil {
		panic(err)
	}
	return c
}

type renewDi struct {
	fx.In
	VaultClient     *vault.Client `optional:"true"`
}

func setupRenewal(lc fx.Lifecycle, di renewDi) {
	if di.VaultClient == nil {
		return
	}
	client := di.VaultClient

	renewer, err := client.GetClientTokenRenewer()

	if err != nil {
		panic("cannot create renewer for vault token")
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			//r.Renew() starts a blocking process to periodically renew the token. Therefore we run it as a go routine
			go renewer.Renew()
			//this starts a background process to log the renewal events. These two go routine exits when the renewer is stopped
			go client.MonitorRenew(ctx, renewer, "vault apiClient token")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			renewer.Stop()
			return nil
		},
	})
}

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	VaultClient     *vault.Client `optional:"true"`
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil || di.VaultClient == nil {
		return
	}
	di.HealthRegistrar.Register(&vault.VaultHealthIndicator{
		Client: di.VaultClient,
	})
}