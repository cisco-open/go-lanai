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

package redis

import (
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/certs"
    "github.com/cisco-open/go-lanai/pkg/log"
    "go.uber.org/fx"
)

var logger = log.New("Redis")

var Module = &bootstrap.Module{
	Precedence: bootstrap.RedisPrecedence,
	Options: []fx.Option{
		fx.Provide(BindRedisProperties),
		fx.Provide(provideClientFactory),
		fx.Provide(provideDefaultClient),
		fx.Invoke(registerHealth),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

type factoryDI struct {
	fx.In
	Props       RedisProperties
	CertManager certs.Manager `optional:"true"`
}

func provideClientFactory(di factoryDI) ClientFactory {
	return NewClientFactory(func(opt *FactoryOption) {
		opt.Properties = di.Props
		opt.TLSCertsManager = di.CertManager
	})
}

type clientDI struct {
	fx.In
	AppCtx             *bootstrap.ApplicationContext
	Factory            ClientFactory
	Properties         RedisProperties
}

func provideDefaultClient(di clientDI) Client {
	c, e := di.Factory.New(di.AppCtx, func(opt *ClientOption) {
		opt.DbIndex = di.Properties.DB
	})

	if e != nil {
		panic(e)
	}
	return c
}
