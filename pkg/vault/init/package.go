package vault

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	appconfigInit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	vaulthealth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/health"
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
		fx.Invoke(vaulthealth.Register, manageClientLifecycle),
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
	opts := append([]vault.Options{
		vault.WithProperties(di.Props),
	}, di.Customizers...)
	client, err := vault.New(opts...)
	if err != nil {
		panic(err)
	}
	return client
}

type lcDI struct {
	fx.In
	AppCtx      *bootstrap.ApplicationContext
	Lifecycle   fx.Lifecycle
	VaultClient *vault.Client `optional:"true"`
}

func manageClientLifecycle(di lcDI) {
	if di.VaultClient == nil {
		return
	}
	di.Lifecycle.Append(fx.StartHook(func(_ context.Context) {
		//nolint:contextcheck // Non-inherited new context - intentional. Start hook context expires when startup finishes
		di.VaultClient.AutoRenewToken(di.AppCtx)
	}))
	di.Lifecycle.Append(fx.StopHook(func(_ context.Context) error {
		return di.VaultClient.Close()
	}))
}

type healthDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	VaultClient     *vault.Client    `optional:"true"`
}

func registerHealth(di healthDI) {
	if di.HealthRegistrar == nil || di.VaultClient == nil {
		return
	}
	di.HealthRegistrar.MustRegister(vaulthealth.New(di.VaultClient))
}
