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

package data

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	GormDB *gorm.DB `optional:"true"`
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil || di.GormDB == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&DbHealthIndicator{
		db: di.GormDB,
	})
}

// DbHealthIndicator
// Note: we currently only support one database
type DbHealthIndicator struct {
	db *gorm.DB
}

func (i *DbHealthIndicator) Name() string {
	return "database"
}

func (i *DbHealthIndicator) Health(c context.Context, options health.Options) health.Health {
	if sqldb, e := i.db.DB(); e != nil {
		return health.NewDetailedHealth(health.StatusUnknown, "database ping is not available", nil)
	} else {
		if e := sqldb.Ping(); e != nil {
			return health.NewDetailedHealth(health.StatusDown, "database ping failed", nil)
		} else {
			return health.NewDetailedHealth(health.StatusUp, "database ping succeeded", nil)
		}
	}
}


