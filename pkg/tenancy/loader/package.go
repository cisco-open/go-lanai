package th_loader

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"github.com/cenkalti/backoff/v4"
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

func initializeTenantHierarchy (ctx *bootstrap.ApplicationContext, lc fx.Lifecycle, loader Loader) error {
	go loadTenantHierarchyWithRetry(ctx, loader)
	return nil
}

func loadTenantHierarchyWithRetry(ctx context.Context, loader Loader) {
	f := func() error {
		err := loader.LoadTenantHierarchy(ctx)
		if err != nil {
			logger.WithContext(ctx).Errorf("tenant hierarchy not loaded due to %v", err)
		} else {
			logger.WithContext(ctx).Infof("finished loading tenant hierarchy")
		}
		return err
	}
	logger.WithContext(ctx).Infof("started loading tenant hierarchy")
	expBackoff := backoff.NewExponentialBackOff()
	//continue trying until succeeds
	expBackoff.MaxElapsedTime = 0
	err := backoff.Retry(f, backoff.WithContext(expBackoff, ctx))
	if err != nil { // technically shouldn't enter this state because we don't return a backoff.PermanentError error
		logger.WithContext(ctx).Errorf("tenant hierarchy loading failed and won't be retried %s", err)
	}
}