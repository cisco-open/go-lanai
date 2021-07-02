package tenancy

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"errors"
	"go.uber.org/fx"
)

var internalAccessor Accessor

var Module = &bootstrap.Module{
	Name: "tenant-hierarchy",
	Precedence: bootstrap.TenantHierarchyAccessorPrecedence,
	Options: []fx.Option{
		fx.Provide(bindCacheProperties),
		fx.Provide(defaultTenancyAccessorProvider()),
		fx.Invoke(setup),
	},
}

const (
	fxNameAccessor = "tenancy/accessor"
)

func Use() {
	bootstrap.Register(Module)
}

type defaultDI struct {
	fx.In
	Ctx                    *bootstrap.ApplicationContext
	Cf                     redis.ClientFactory           `optional:"true"`
	Prop                   CacheProperties               `optional:"true"`
	UnnamedTenancyAccessor Accessor                      `optional:"true"`
}

func defaultTenancyAccessorProvider() fx.Annotated {
	return fx.Annotated{
		Name:   fxNameAccessor,
		Target: provideAccessor,
	}
}

func provideAccessor(di defaultDI) Accessor {
	if di.UnnamedTenancyAccessor != nil {
		internalAccessor = di.UnnamedTenancyAccessor
		return di.UnnamedTenancyAccessor
	}

	if di.Cf == nil {
		panic(errors.New("redis client factory is required"))
	}

	rc, e := di.Cf.New(di.Ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = di.Prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	internalAccessor = newAccessor(rc)
	return internalAccessor
}

type setupDI struct {
	fx.In
	EffectiveAccessor Accessor `name:"tenancy/accessor"`
}

func setup(_ setupDI) {
}
