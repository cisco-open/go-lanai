package tenant_hierarchy_accessor

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"go.uber.org/fx"
)

var RedisClient redis.Client

var Module = &bootstrap.Module{
	Name: "tenant-hierarchy",
	Precedence: bootstrap.TenantHierarchyAccessorPrecedence,
	Options: []fx.Option{
		fx.Provide(BindCacheProperties),
		fx.Invoke(setupRedisClient),
	},
}

//make this a set effect because when tenant hierarchy loader import this, we automatically want the accessor to work
func init() {
	bootstrap.Register(Module)
}

func setupRedisClient(ctx *bootstrap.ApplicationContext, cf redis.ClientFactory, prop CacheProperties) error {
	var e error
	RedisClient, e = cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = prop.DbIndex
	})
	if e != nil {
		return e
	}
	return nil
}
