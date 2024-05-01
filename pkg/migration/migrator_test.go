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
	"embed"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/data/postgresql/cockroach"
	"github.com/cisco-open/go-lanai/pkg/migration"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*************************
	Setup Test
 *************************/

//go:embed testdata/*.sql
var TestStepsFS embed.FS

const TestTableName = `migration_migrator_test`

/*************************
	Tests
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type TestMigrateDI struct {
	fx.In
	dbtest.DI
}

func TestMigrate(t *testing.T) {
	di := TestMigrateDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithFxOptions(
			fx.Provide(cockroach.NewGormDbCreator),
			fx.Provide(
				migration.DefaultGormConfigurerProvider(),
			),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupDropMigrationTable(&di.DI)),
		test.GomegaSubTest(SubTestMigrateSuccess(&di), "TestMigrateSuccess"),
		test.GomegaSubTest(SubTestMigrateFailAndResume(&di), "TestMigrateFailAndResume"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupDropMigrationTable(di *dbtest.DI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		if e := DropMigrationTable(di.DB); e != nil {
			return ctx, e
		}
		rs := di.DB.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s";`, TestTableName))
		return ctx, rs.Error
	}
}

func SubTestMigrateSuccess(di *TestMigrateDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		reg := migration.NewRegistrar()
		ver := migration.NewGormVersioner(di.DB)
		reg.AddMigrations(
			migration.WithVersion("1.0.0").Dot(1).WithTag(migration.TagPreUpgrade).
				WithDesc("Step 1 - Create table from SQL file").WithFile(TestStepsFS, "testdata/test.sql", di.DB),
			migration.WithVersion("1.0.0").Dot(2).WithTag(migration.TagPostUpgrade).
				WithDesc("Step 2 - Seed some data").
				WithFunc(func(ctx context.Context) error {
					rs := di.DB.Exec(fmt.Sprintf(`INSERT INTO "%s" ("id") VALUES ('first record')`, TestTableName))
					return rs.Error
				}),
		)
		e := migration.Migrate(ctx, reg, ver)
		g.Expect(e).To(Succeed(), "migration should not fail")
		AssertMigrationResult(g, di.DB, "1.0.0.1", true)
		AssertMigrationResult(g, di.DB, "1.0.0.2", true)
		var rs *gorm.DB
		var count int64
		rs = di.DB.Model(&TestModel{}).Count(&count)
		g.Expect(rs.Error).To(Succeed(), "selecting from created table should not fail")
		g.Expect(count).To(BeEquivalentTo(1), "created table should have correct records")
	}
}

func SubTestMigrateFailAndResume(di *TestMigrateDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var migrationErr = fmt.Errorf("oops")
		reg := migration.NewRegistrar()
		ver := migration.NewGormVersioner(di.DB)
		reg.AddMigrations(
			migration.WithVersion("1.0.0").Dot(1).WithTag(migration.TagPreUpgrade).
				WithDesc("Step 1 - Create table from SQL file").WithFile(TestStepsFS, "testdata/test.sql", di.DB),
			migration.WithVersion("1.0.0").Dot(2).WithTag(migration.TagPostUpgrade).
				WithDesc("Step 2 - Seed some data").
				WithFunc(func(ctx context.Context) error {
					return migrationErr
				}),
		)
		// first pass
		e := migration.Migrate(ctx, reg, ver)
		g.Expect(e).To(HaveOccurred(), "migration should fail")
		AssertMigrationResult(g, di.DB, "1.0.0.1", true)
		AssertMigrationResult(g, di.DB, "1.0.0.2", false)
		var rs *gorm.DB
		var count int64
		rs = di.DB.Model(&TestModel{}).Count(&count)
		g.Expect(rs.Error).To(Succeed(), "selecting from created table should not fail")
		g.Expect(count).To(BeEquivalentTo(0), "created table should be empty")

		// 2nd pass
		reg = migration.NewRegistrar()
		reg.AddMigrations(
			migration.WithVersion("1.0.0").Dot(1).WithTag(migration.TagPreUpgrade).
				WithDesc("Step 1 - Create table from SQL file").WithFile(TestStepsFS, "testdata/test.sql", di.DB),
			migration.WithVersion("1.0.0").Dot(2).WithTag(migration.TagPostUpgrade).
				WithDesc("Step 2 - Seed some data").
				WithFunc(func(ctx context.Context) error {
					rs := di.DB.Exec(fmt.Sprintf(`INSERT INTO "%s" ("id") VALUES ('first record')`, TestTableName))
					return rs.Error
				}),
		)

		e = migration.Migrate(ctx, reg, ver)
		g.Expect(e).To(HaveOccurred(), "migration should fail again without clearing failed record")
		AssertMigrationResult(g, di.DB, "1.0.0.1", true)
		AssertMigrationResult(g, di.DB, "1.0.0.2", false)
		rs = di.DB.Model(&TestModel{}).Count(&count)
		g.Expect(rs.Error).To(Succeed(), "selecting from created table should not fail")
		g.Expect(count).To(BeEquivalentTo(0), "created table should still be empty")

		// manual clear and 3rd pass
		rs = di.DB.Delete(&migration.MigrationVersion{Version: []int{1, 0, 0, 2}})
		g.Expect(rs.Error).To(Succeed(), "clearing failed step should not fail")
		e = migration.Migrate(ctx, reg, ver)
		g.Expect(e).To(Succeed(), "migration should not faile after clearing failed record")
		AssertMigrationResult(g, di.DB, "1.0.0.1", true)
		AssertMigrationResult(g, di.DB, "1.0.0.2", true)
		rs = di.DB.Model(&TestModel{}).Count(&count)
		g.Expect(rs.Error).To(Succeed(), "selecting from created table should not fail")
		g.Expect(count).To(BeEquivalentTo(1), "created table should be correct")
	}
}

/*************************
	Helpers
 *************************/

func AssertMigrationResult(g *gomega.WithT, db *gorm.DB, version string, success bool) {
	var record migration.MigrationVersion
	rs := db.Model(&migration.MigrationVersion{}).Take(&record, "Version = ?", version)
	g.Expect(rs.Error).To(Succeed(), "migration version [%s] should exists", version)
	g.Expect(record.GetVersion().String()).To(Equal(version), "migration version should be correct")
	g.Expect(record.IsSuccess()).To(Equal(success), "migration result with version [%s] should be correct", version)
	g.Expect(record.GetDescription()).ToNot(BeEmpty(), "migration step description of [%s] should be correct", version)
	g.Expect(record.GetInstalledOn()).ToNot(BeZero(), "migration step time of [%s] should be correct", version)
}

type TestModel struct {
	ID string `gorm:"primaryKey"`
}

func (TestModel) TableName() string {
	return TestTableName
}
