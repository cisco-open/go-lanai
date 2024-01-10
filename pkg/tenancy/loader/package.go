package th_loader

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"go.uber.org/fx"
)

var logger = log.New("Tenancy.Load")

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

type loaderDI struct {
	fx.In
	Ctx *bootstrap.ApplicationContext
	Store TenantHierarchyStore
	Cf redis.ClientFactory
	Prop tenancy.CacheProperties
	Accessor tenancy.Accessor `name:"tenancy/accessor"`
}

func provideLoader(di loaderDI) Loader {
	rc, e := di.Cf.New(di.Ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = di.Prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	internalLoader = NewLoader(rc, di.Store, di.Accessor)
	return internalLoader
}

func initializeTenantHierarchy (ctx *bootstrap.ApplicationContext, loader Loader) error {
	logger.WithContext(ctx).Infof("started loading tenant hierarchy")
	internalLoader = loader
	err := LoadTenantHierarchy(ctx)
	if err != nil {
		logger.WithContext(ctx).Errorf("tenant hierarchy not loaded due to %v", err)
	} else {
		logger.WithContext(ctx).Infof("finished loading tenant hierarchy")
	}
	return err
}
