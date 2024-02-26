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

package th_modifier

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/tenancy"
	"go.uber.org/fx"
)

var logger = log.New("Tenancy.Modify")

var internalModifier Modifier

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
	Accessor tenancy.Accessor `name:"tenancy/accessor"`
}

func provideModifier(di modifierDI) Modifier {
	rc, e := di.Cf.New(di.Ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = di.Prop.DbIndex
	})
	if e != nil {
		panic(e)
	}
	internalModifier = newModifier(rc, di.Accessor)
	return internalModifier
}

func setup(_ Modifier) {
	// currently, keep everything default
}