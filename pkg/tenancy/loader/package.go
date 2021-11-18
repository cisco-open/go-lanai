package th_loader

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/scheduler"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("tenancy-loader")

var internalLoader Loader
var retryInterval = 10 * time.Second

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

func initializeTenantHierarchy (appCtx *bootstrap.ApplicationContext, lc fx.Lifecycle, loader Loader) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var canceller scheduler.TaskCanceller
			var e error

			fn := func(ctx context.Context) error {
				err := loader.LoadTenantHierarchy(ctx)
				if err != nil {
					logger.WithContext(ctx).Errorf("tenant hierarchy loaded failed due to %v. will be retried in %v", err, 10 * time.Second)
				} else {
					logger.WithContext(ctx).Infof("finished loading tenant hierarchy")
					canceller.Cancel()
					<-canceller.Cancelled()
					logger.WithContext(ctx).Infof("stopped tenant hierarchy load task")
				}
				return err
			}

			canceller, e = scheduler.Repeat(fn, scheduler.AtRate(retryInterval))
			return e
		},
	})
}