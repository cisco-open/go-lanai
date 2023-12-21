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

package monitor

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"embed"
	"go.uber.org/fx"
)

var logger = log.New("PProf.Charts")

//go:embed static/*
var Content embed.FS

var Module = &bootstrap.Module{
	Precedence: bootstrap.DebugPrecedence,
	Options: []fx.Option{
		fx.Provide(provideDataStorage, NewDataCollector),
		fx.Invoke(initialize),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

type storageDI struct {
	fx.In
	AppCtx       *bootstrap.ApplicationContext
	RedisFactory redis.ClientFactory `optional:"true"`
}

func provideDataStorage(di storageDI) DataStorage {
	if di.RedisFactory != nil {
		return NewRedisDataStorage(di.AppCtx, di.RedisFactory)
	}
	return nil // TODO: in-memory storage as fallback
}

type initDI struct {
	fx.In
	LC        fx.Lifecycle
	AppCtx    *bootstrap.ApplicationContext
	Registrar *web.Registrar `optional:"true"`
	Collector *dataCollector
}

func initialize(di initDI) {
	if di.Registrar != nil {
		di.Registrar.MustRegister(Content)
		di.Registrar.MustRegister(NewChartController(di.Collector.storage, di.Collector))
	}

	di.Collector.Start(di.AppCtx)

	di.LC.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			di.Collector.Stop()
			return nil
		},
	})
}
