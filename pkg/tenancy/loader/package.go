// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package th_loader

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/tenancy"
	"go.uber.org/fx"
)

var logger = log.New("Tenancy.Load")

var internalLoader Loader

var Module = &bootstrap.Module{
	Name:       "tenancy-loader",
	Precedence: bootstrap.TenantHierarchyLoaderPrecedence,
	Options: []fx.Option{
		fx.Provide(defaultLoader()),
		fx.Invoke(initializeTenantHierarchy),
	},
}

const (
	fxNameLoader = "tenant-hierarchy/loader"
)

func Use() {
	tenancy.Use()
	bootstrap.Register(Module)
}

type loaderDI struct {
	fx.In
	Ctx           *bootstrap.ApplicationContext
	Store         TenantHierarchyStore
	Cf            redis.ClientFactory
	Prop          tenancy.CacheProperties
	Accessor      tenancy.Accessor `name:"tenancy/accessor"`
	UnnamedLoader Loader           `optional:"true"`
}

func defaultLoader() fx.Annotated {
	return fx.Annotated{
		Name:   fxNameLoader,
		Target: provideLoader,
	}
}

func provideLoader(di loaderDI) Loader {
	if di.UnnamedLoader != nil {
		internalLoader = di.UnnamedLoader
		return di.UnnamedLoader
	}
	rc, e := di.Cf.New(di.Ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = di.Prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	internalLoader = NewLoader(rc, di.Store, di.Accessor)
	return internalLoader
}

type initDi struct {
	fx.In
	AppCtx          *bootstrap.ApplicationContext
	EffectiveLoader Loader `name:"tenant-hierarchy/loader"`
}

func initializeTenantHierarchy(di initDi) error {
	ctx := di.AppCtx
	logger.WithContext(ctx).Infof("started loading tenant hierarchy")
	internalLoader = di.EffectiveLoader
	err := LoadTenantHierarchy(ctx)
	if err != nil {
		logger.WithContext(ctx).Errorf("tenant hierarchy not loaded due to %v", err)
	} else {
		logger.WithContext(ctx).Infof("finished loading tenant hierarchy")
	}
	return err

}
