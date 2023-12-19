package vault

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	appconfigInit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"embed"
	"go.uber.org/fx"
)

//go:embed defaults-vault.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "vault",
	Precedence: bootstrap.VaultPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(BindConnectionProperties, ProvideDefaultClient),
	},
	Options: []fx.Option{
		appconfigInit.FxEmbeddedDefaults(defaultConfigFS),
		fx.Invoke(setupRenewal, registerHealth),
	},
}

// Use func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

func BindConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) (vault.ConnectionProperties, error) {
	c := vault.ConnectionProperties{
		Host:           "localhost",
		Port:           8200,
		Scheme:         "http",
		Authentication: vault.Token,
		Token:          "replace_with_token_value",
	}
	if e := bootstrapConfig.Bind(&c, vault.PropertyPrefix); e != nil {
		return c, e
	}
	return c, nil
}

type clientDI struct {
	fx.In
	Props       vault.ConnectionProperties
	Customizers []vault.Options `group:"vault"`
}

func ProvideDefaultClient(di clientDI) *vault.Client {
	opts := append([]vault.Options{vault.WithProperties(di.Props)}, di.Customizers...)
	c, err := vault.New(opts...)
	if err != nil {
		panic(err)
	}
	return c
}

type renewDi struct {
	fx.In
	AppContext  *bootstrap.ApplicationContext
	VaultClient *vault.Client `optional:"true"`
}

func setupRenewal(lc fx.Lifecycle, di renewDi) {
	if di.VaultClient == nil {
		return
	}
	client := di.VaultClient
	refresher := vault.NewTokenRefresher(client)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			//nolint:contextcheck // intended, we don't use passed in context, refresher will depend on application context
			refresher.Start(di.AppContext)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			refresher.Stop()
			return nil
		},
	})
}

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	VaultClient     *vault.Client    `optional:"true"`
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil || di.VaultClient == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&vault.VaultHealthIndicator{
		Client: di.VaultClient,
	})
}
