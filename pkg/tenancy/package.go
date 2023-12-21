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

package tenancy

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"errors"
	"go.uber.org/fx"
)

var internalAccessor Accessor

var Module = &bootstrap.Module{
	Name:       "tenant-hierarchy",
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
	Cf                     redis.ClientFactory `optional:"true"`
	Prop                   CacheProperties     `optional:"true"`
	UnnamedTenancyAccessor Accessor            `optional:"true"`
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
	// keep it as default for now
}
