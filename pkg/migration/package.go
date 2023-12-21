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

package migration

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const (
	TagPreUpgrade = "pre_upgrade"
	TagPostUpgrade = "post_upgrade"
)

var logger = log.New("Migration")

var filterFlag string
var allowOutOfOrderFlag bool

var Module = &bootstrap.Module{
	Name: "migration",
	Precedence: bootstrap.MigrationPrecedence,
	Options: []fx.Option{
		fx.Provide(newRegistrar),
		fx.Provide(newVersioner),
		fx.Invoke(applyMigrations),
	},
}

func Use() {
	bootstrap.AddStringFlag(&filterFlag, "filter", "", fmt.Sprintf("filter the migration steps by tag value. supports %s or %s", TagPreUpgrade, TagPostUpgrade))
	bootstrap.AddBoolFlag(&allowOutOfOrderFlag, "allow_out_of_order", false, fmt.Sprintf("allow migration steps to execute out of order"))
	bootstrap.Register(Module)
}

func newRegistrar() *Registrar {
	return &Registrar{}
}

func newVersioner(db *gorm.DB) Versioner {
	return &GormVersioner{
		db: db,
	}
}

func applyMigrations(lc fx.Lifecycle, r *Registrar, v Versioner, db *gorm.DB, dbCreator data.DbCreator, shutdowner fx.Shutdowner) {
	var err error
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err = dbCreator.CreateDatabaseIfNotExist(ctx, db)
			if err != nil {
				return shutdowner.Shutdown()
			}
			err = migrate(ctx, r, v)
			return shutdowner.Shutdown()
		},
		OnStop:  func(ctx context.Context) error {
			return err
		},
	})
}