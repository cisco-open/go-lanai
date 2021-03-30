package tenant_hierarchy

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	tenant_hierarchy_accessor "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenant_hierarchy/accessor"
	"go.uber.org/fx"
)

var logger = log.New("tenant_hierarchy_loader")

var messageListener redis.Client

var Module = &bootstrap.Module{
	Name: "tenant-hierarchy",
	Precedence: bootstrap.TenantHierarchyLoaderPrecedence,
	Options: []fx.Option{
		fx.Invoke(initializeTenantHierarchy),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func initializeTenantHierarchy (ctx *bootstrap.ApplicationContext, store TenantHierarchyStore, cf redis.ClientFactory, prop tenant_hierarchy_accessor.CacheProperties) {
	err := LoadTenantHierarchy(ctx, store, tenant_hierarchy_accessor.RedisClient)
	if err != nil {
		logger.Warnf("tenant hierarchy not loaded due to %v", err)
	}
}
