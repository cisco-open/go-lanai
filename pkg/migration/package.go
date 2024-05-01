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
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const (
	TagPreUpgrade  = "pre_upgrade"
	TagPostUpgrade = "post_upgrade"
)

var logger = log.New("Migration")

var filterFlag string
var allowOutOfOrderFlag bool

var Module = &bootstrap.Module{
	Name:       "migration",
	Precedence: bootstrap.CommandLineRunnerPrecedence,
	Options: []fx.Option{
		fx.Provide(NewRegistrar),
		fx.Provide(NewGormVersioner),
		fx.Provide(provideMigrationRunner()),
	},
}

func Use() {
	bootstrap.AddStringFlag(&filterFlag, "filter", "", fmt.Sprintf("filter the migration steps by tag value. supports %s or %s", TagPreUpgrade, TagPostUpgrade))
	bootstrap.AddBoolFlag(&allowOutOfOrderFlag, "allow_out_of_order", false, fmt.Sprintf("allow migration steps to execute out of order"))
	bootstrap.Register(Module)
	// Note: migration CliRunner is provided in Module
	bootstrap.EnableCliRunnerMode()
}

func provideMigrationRunner() fx.Annotated {
	return fx.Annotated{
		Group:  bootstrap.FxCliRunnerGroup,
		Target: newMigrationRunner,
	}
}

type migrationRunnerIn struct {
	fx.In
	R          *Registrar
	V          Versioner
	DB         *gorm.DB
	DbCreators []data.DbCreator `group:"db_creator"`
}

func newMigrationRunner(di migrationRunnerIn) bootstrap.CliRunner {
	return func(ctx context.Context) error {
		if len(di.DbCreators) > 0 {
			order.SortStable(di.DbCreators, order.OrderedFirstCompare)
			dbCreator := di.DbCreators[0]
			if err := dbCreator.CreateDatabaseIfNotExist(ctx, di.DB); err != nil {
				return err
			}
		}
		return Migrate(ctx, di.R, di.V)
	}
}
