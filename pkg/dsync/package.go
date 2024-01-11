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

package dsync

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"fmt"
	"go.uber.org/fx"
)

//go:embed defaults-dsync.yml
var defaultConfigFS embed.FS

var logger = log.New("DSync")

var syncManager SyncManager

var Module = &bootstrap.Module{
	Name:       "distributed",
	Precedence: bootstrap.DistributedLockPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(provideSyncManager),
		fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

type syncDI struct {
	fx.In
	AppCtx          *bootstrap.ApplicationContext
	Conn            *consul.Connection `optional:"true"`
	TestSyncManager []SyncManager      `group:"test"`
}

func provideSyncManager(di syncDI) (SyncManager, error) {
	if len(di.TestSyncManager) != 0 {
		return di.TestSyncManager[0], nil
	}
	if di.Conn == nil {
		return nil, fmt.Errorf("*consul.Connection is required for 'dsync' package")
	}
	return NewConsulLockManager(di.AppCtx, di.Conn), nil
}

/**************************
	Initialize
***************************/

type initDI struct {
	fx.In
	Lifecycle fx.Lifecycle
	AppCtx    *bootstrap.ApplicationContext
	Manager   SyncManager
}

func initialize(di initDI) {
	// set global variable
	syncManager = di.Manager
	syncLc, ok := di.Manager.(SyncManagerLifecycle)
	di.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if ok {
				if e := syncLc.Start(ctx); e != nil {
					return e
				}
			}
			// start leader election lock
			return startLeadershipLock(ctx, di)
		},
		OnStop: func(ctx context.Context) error {
			if ok {
				return syncLc.Stop(ctx)
			}
			return nil
		},
	})
}
