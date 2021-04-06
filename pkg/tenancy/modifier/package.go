package th_modifier

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"go.uber.org/fx"
)

var logger = log.New("tenant_hierarchy_modifier")

var internaModifier Modifier

var Module = &bootstrap.Module{
	Name: "tenancy-modifier",
	Precedence: bootstrap.TenantHierarchyModifierPrecedence,
	Options: []fx.Option{
		fx.Provide(provideModifier),
		fx.Invoke(setup),
	},
}

func Use() {
	tenancy.Use()
	bootstrap.Register(Module)
}

func provideModifier(ctx *bootstrap.ApplicationContext, cf redis.ClientFactory, prop tenancy.CacheProperties) Modifier {
	rc, e := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	return newModifier(rc)
}

func setup(ctx *bootstrap.ApplicationContext, modifier Modifier) {
	internaModifier = modifier
}