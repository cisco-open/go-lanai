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

package migration_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/data/postgresql/cockroach"
	"github.com/cisco-open/go-lanai/pkg/migration"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*
************************

		Setup Test
	 ************************
*/
func ProvideMigrationTableDropper() fx.Annotated {
	return fx.Annotated{
		Group: bootstrap.FxCliRunnerGroup,
		Target: func(db *gorm.DB) bootstrap.OrderedCliRunner {
			return bootstrap.OrderedCliRunner{
				Precedence: order.Highest, // run before migration
				CliRunner: func(ctx context.Context) error {
					return DropMigrationTable(db)
				},
			}
		},
	}
}

func DropMigrationTable(db *gorm.DB) error {
	rs := db.Exec(`DROP TABLE IF EXISTS "migration_versions";`)
	return rs.Error
}

const packageTestSQL = `create table if not exists migration_package_test(id uuid default gen_random_uuid() not null primary key);`

func RegisterSimpleMigrationStep(reg *migration.Registrar, db *gorm.DB) {
	reg.AddMigrations(migration.WithVersion("1.0.0").Dot(1).WithTag(migration.TagPreUpgrade).
		WithDesc("A test migration step").
		WithFunc(func(ctx context.Context) error {
			rs := db.Exec(packageTestSQL)
			return rs.Error
		}),
	)
}

/*************************
	Tests
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type TestPackageDI struct {
	fx.In
	dbtest.DI
}

func TestModuleInit(t *testing.T) {
	di := TestPackageDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(migration.Module),
		apptest.WithFxOptions(
			fx.Provide(cockroach.NewGormDbCreator),
			fx.Provide(migration.DefaultGormConfigurerProvider()),
			fx.Provide(ProvideMigrationTableDropper()),
			fx.Invoke(RegisterSimpleMigrationStep),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestMigratorAutoRun(&di), "TestMigratorAutoRun"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestMigratorAutoRun(di *TestPackageDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var records []*migration.MigrationVersion
		var rs *gorm.DB
		rs = di.DB.Model(&migration.MigrationVersion{}).Order("version ASC").Find(&records)
		g.Expect(rs.Error).To(Succeed(), "getting migration versions should not fail")
		g.Expect(records).To(HaveLen(1), "migration versions should have correct count")
		g.Expect(records[0].GetVersion().String()).To(Equal("1.0.0.1"), "migration version should be correct")
		g.Expect(records[0].IsSuccess()).To(BeTrue(), "migration should succeeded")

		rs = di.DB.Exec("SELECT * FROM public.migration_package_test;")
		g.Expect(rs.Error).To(Succeed(), "selecting from created table should not fail")
	}
}

/*************************
	Helpers
 *************************/
