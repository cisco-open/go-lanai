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
	"embed"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
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
		fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

/**************************
	Initialize
***************************/

type initDI struct {
	fx.In
	Lifecycle   fx.Lifecycle
	AppCtx      *bootstrap.ApplicationContext
	Manager     SyncManager   `optional:"true"`
	TestManager []SyncManager `group:"test"`
}

func initialize(di initDI) error {
	// set global variable
	syncManager = di.Manager
	if len(di.TestManager) != 0 {
		syncManager = di.TestManager[0]
	}
	if syncManager == nil {
		return ErrFailedInitialization.WithMessage(`unable to initialize distributed lock system and leadership lock. ` +
			`Hint: provide a dsync.SyncManager with 'consuldsync.Use()' or 'redisdsync.Use()' or with your own implementation `)
	}
	syncLc, ok := syncManager.(SyncManagerLifecycle)

	// start/stop hooks
	di.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if ok {
				if e := syncLc.Start(ctx); e != nil {
					return ErrFailedInitialization.WithCause(e)
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
	return nil
}
