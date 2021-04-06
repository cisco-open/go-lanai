package tenancy

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"go.uber.org/fx"
)

var internalAccessor Accessor

var Module = &bootstrap.Module{
	Name: "tenant-hierarchy",
	Precedence: bootstrap.TenantHierarchyAccessorPrecedence,
	Options: []fx.Option{
		fx.Provide(bindCacheProperties),
		fx.Provide(provideAccessor),
		fx.Invoke(setup),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func provideAccessor(ctx *bootstrap.ApplicationContext, cf redis.ClientFactory, prop CacheProperties) Accessor {
	rc, e := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	return newAccessor(rc)
}

func setup(a Accessor) {
	internalAccessor = a
}
