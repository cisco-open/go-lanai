package service

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Example")

// Maker func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.AddOptions(
		fx.Provide(BindAccountsProperties),
		fx.Provide(BindAccountPoliciesProperties),
		fx.Provide(BindTenantProperties),
		fx.Provide(BindClientsProperties),
		fx.Provide(NewInMemoryAccountStore),
		fx.Provide(NewInMemoryFederatedAccountStore),
		fx.Provide(NewInMemoryClientStore),
		fx.Provide(NewTenantStore),
		fx.Provide(NewProviderStore),
		fx.Provide(NewInMemoryIdpManager),
		fx.Provide(NewInMemSpManager),
	)
}
