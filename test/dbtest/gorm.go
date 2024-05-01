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

package dbtest

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/pkg/data/postgresql"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormtest "gorm.io/gorm/utils/tests"
	"strings"
)

/*****************************
	gorm postgres Dialetor
 *****************************/

type dialectorDI struct {
	fx.In
}

func testGormDialectorProvider(opt *DBOption) func(di dialectorDI) gorm.Dialector {
	return func(di dialectorDI) gorm.Dialector {
		ssl := "disable"
		if opt.SSL {
			ssl = "enable"
		}
		options := map[string]interface{}{
			dsKeyHost:    opt.Host,
			dsKeyPort:    opt.Port,
			dsKeyDB:      opt.DBName,
			dsKeySslMode: ssl,
		}

		if opt.Username != "" {
			options[dsKeyUsername] = opt.Username
			options[dsKeyPassword] = opt.Password
		}

		config := postgres.Config{
			DriverName: "copyist_postgres",
			DSN:        toDSN(options),
		}
		return postgresql.NewGormDialectorWithConfig(config)
	}
}

func toDSN(options map[string]interface{}) string {
	opts := make([]string, 0)
	for k, v := range options {
		opt := fmt.Sprintf("%s=%v", k, v)
		opts = append(opts, opt)
	}
	return strings.Join(opts, " ")
}

/****************************
	gorm Noop Dialector
 ****************************/

type noopGormDialector struct {
	gormtest.DummyDialector
}

func provideNoopGormDialector() gorm.Dialector {
	return noopGormDialector{gormtest.DummyDialector{}}
}

func (d noopGormDialector) SavePoint(_ *gorm.DB, _ string) error {
	return nil
}

func (d noopGormDialector) RollbackTo(_ *gorm.DB, _ string) error {
	return nil
}

/*****************************
	gorm cockroach error
 *****************************/

func pqErrorTranslatorProvider() fx.Annotated {
	return fx.Annotated{
		Group: data.GormConfigurerGroup,
		Target: func() data.ErrorTranslator {
			return postgresql.PostgresErrorTranslator{}
		},
	}
}

/*****************************
	gorm dry run
 *****************************/

func enableGormDryRun(db *gorm.DB) {
	db.DryRun = true
	db.SkipDefaultTransaction = true
}
