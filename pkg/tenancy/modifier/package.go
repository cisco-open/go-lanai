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

type modifierDI struct {
	fx.In
	Ctx *bootstrap.ApplicationContext
	Cf redis.ClientFactory
	Prop tenancy.CacheProperties
	Accessor tenancy.Accessor `name:"tenancy/tenancyAccessor"`
}

func provideModifier(di modifierDI) Modifier {
	rc, e := di.Cf.New(di.Ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = di.Prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	internaModifier = newModifier(rc, di.Accessor)
	return internaModifier
}

func setup(_ Modifier) {
}