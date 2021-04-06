package th_loader

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"go.uber.org/fx"
)

var logger = log.New("tenancy-loader")

var internalLoader Loader

var Module = &bootstrap.Module{
	Name: "tenancy-loader",
	Precedence: bootstrap.TenantHierarchyLoaderPrecedence,
	Options: []fx.Option{
		fx.Provide(provideLoader),
		fx.Invoke(initializeTenantHierarchy),
	},
}

func Use() {
	tenancy.Use()
	bootstrap.Register(Module)
}

func provideLoader(ctx *bootstrap.ApplicationContext, store TenantHierarchyStore, cf redis.ClientFactory, prop tenancy.CacheProperties, accessor tenancy.Accessor) Loader {
	rc, e := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	return NewLoader(rc, store, accessor)
}

func initializeTenantHierarchy (ctx *bootstrap.ApplicationContext, loader Loader) error {
	internalLoader = loader
	err := internalLoader.LoadTenantHierarchy(ctx)
	if err != nil {
		logger.Errorf("tenant hierarchy not loaded due to %v", err)
	}
	return err
}
