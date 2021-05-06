package appconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/vaultprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"fmt"
	"go.uber.org/fx"
)

type vaultDi struct {
	fx.In
	BootstrapConfig       *appconfig.BootstrapConfig
	VaultConfigProperties *vaultprovider.KvConfigProperties
	VaultClient           *vault.Client `optional:"true"`
}

func newVaultDefaultContextProviderGroup(di vaultDi) appConfigProvidersOut {
	if !di.VaultConfigProperties.Enabled || di.VaultClient == nil{
		return appConfigProvidersOut{}
	}

	kvSecretEngine, err := vaultprovider.NewKvSecretEngine(
		di.VaultConfigProperties.BackendVersion, di.VaultConfigProperties.Backend, di.VaultClient)

	if err != nil {
		panic(err)
	}

	group := appconfig.NewProfileBasedProviderGroup(externalDefaultContextPrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return di.VaultConfigProperties.DefaultContext
		}
		return fmt.Sprintf("%s%s%s", di.VaultConfigProperties.DefaultContext, di.VaultConfigProperties.ProfileSeparator, profile)
	}
	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		return vaultprovider.NewVaultKvProvider(order, name, kvSecretEngine)
	}

	return appConfigProvidersOut {
		ProviderGroup: group,
	}
}

func newVaultAppContextProviderGroup(di vaultDi) appConfigProvidersOut {
	if !di.VaultConfigProperties.Enabled || di.VaultClient == nil{
		return appConfigProvidersOut{}
	}

	kvSecretEngine, err := vaultprovider.NewKvSecretEngine(
		di.VaultConfigProperties.BackendVersion, di.VaultConfigProperties.Backend, di.VaultClient)

	if err != nil {
		panic(err)
	}

	appName := di.BootstrapConfig.Value(bootstrap.PropertyKeyApplicationName)

	group := appconfig.NewProfileBasedProviderGroup(externalAppContextPrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return fmt.Sprintf("%s", appName)
		}
		return fmt.Sprintf("%s%s%s", appName, di.VaultConfigProperties.ProfileSeparator, profile)
	}
	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		ptr := vaultprovider.NewVaultKvProvider(order, name, kvSecretEngine)
		if ptr == nil {
			return nil
		}
		return ptr
	}

	return appConfigProvidersOut {
		ProviderGroup: group,
	}
}
