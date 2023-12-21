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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"io"
	"io/fs"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// PrepareData is a convenient function that returns a test.SetupFunc that executes given DataSetupStep in provided order
// Note: PrepareData accumulate all changes applied to context
func PrepareData(di *DI, steps ...DataSetupStep) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		for _, fn := range steps {
			ctx = fn(ctx, t, di.DB)
			if t.Failed() {
				return ctx, errors.New("test failed during data preparation")
			}
		}
		return ctx, nil
	}
}

// PrepareDataWithScope is similar to PrepareData, it applies given DataSetupScope before executing all DataSetupStep.
// DataSetupScope is used to prepare context and gorm.DB for all given DataSetupStep
// Note: Different from PrepareData, PrepareDataWithScope doesn't accumulate changes to context
func PrepareDataWithScope(di *DI, scope DataSetupScope, steps ...DataSetupStep) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		scopedCtx, db := scope(ctx, t, di.DB)
		for _, fn := range steps {
			scopedCtx = fn(scopedCtx, t, db)
			if t.Failed() {
				return ctx, errors.New("test failed during data preparation")
			}
		}
		return ctx, nil
	}
}

// SetupUsingSQLFile returns a DataSetupStep that execute the provided sql file in given FS.
func SetupUsingSQLFile(fsys fs.FS, filenames ...string) DataSetupStep {
	return func(ctx context.Context, t *testing.T, db *gorm.DB) context.Context {
		g := gomega.NewWithT(t)
		for _, filename := range filenames {
			execSqlFile(ctx, fsys, db, g, filename)
		}
		return ctx
	}
}

// SetupUsingSQLQueries returns a DataSetupStep that execute the provided sql queries.
func SetupUsingSQLQueries(queries ...string) DataSetupStep {
	return func(ctx context.Context, t *testing.T, db *gorm.DB) context.Context {
		g := gomega.NewWithT(t)
		for _, q := range queries {
			r := db.WithContext(ctx).Exec(q)
			g.Expect(r.Error).To(Succeed(), "table preparation should be able to run SQL '%s'", q)
			if t.Failed() {
				return ctx
			}
		}
		return ctx
	}
}

// SetupUsingModelSeedFile returns a DataSetupStep that load provided yaml file
// and parse it directly into provided model and save them.
// when "closures" is provided, it's invoked after seeding is done.
func SetupUsingModelSeedFile(fsys fs.FS, dest interface{}, filename string, closures...func(ctx context.Context, db *gorm.DB)) DataSetupStep {
	return func(ctx context.Context, t *testing.T, db *gorm.DB) context.Context {
		g := gomega.NewWithT(t)
		e := loadSeedData(fsys, dest, filename)
		g.Expect(e).To(Succeed(), "data preparation should be able to parse model's seed file")
		if t.Failed() {
			return ctx
		}
		tx := db.WithContext(ctx).CreateInBatches(dest, 100)
		g.Expect(tx.Error).To(Succeed(), "data preparation should be able to create models seed file")
		if t.Failed() {
			return ctx
		}
		for _, fn := range closures {
			fn(ctx, db)
		}
		return ctx
	}
}

// SetupTruncateTables returns a DataSetupStep that truncate given tables in the provided order
func SetupTruncateTables(tables ...string) DataSetupStep {
	sqls := make([]string, len(tables))
	for i, table := range tables {
		sqls[i] = truncateTableSql(table)
	}
	return SetupUsingSQLQueries(sqls...)
}

// SetupDropTables returns a DataSetupStep that truncate given tables in single DROP TABLE IF EXISTS
func SetupDropTables(tables ...string) DataSetupStep {
	tableLiterals := make([]string, len(tables))
	for i := range tables {
		tableLiterals[i] = fmt.Sprintf(`"%s"`, tables[i])
	}
	sql := fmt.Sprintf(`DROP TABLE IF EXISTS %s CASCADE;`, strings.Join(tableLiterals, ", "))
	return SetupUsingSQLQueries(sql)
}

// SetupOnce returns a DataSetupStep that run given DataSetupSteps within the given sync.Once.
// How sync.Once is scoped is up to caller. e.g. once per test, once per package execution, etc...
func SetupOnce(once *sync.Once, steps ...DataSetupStep) DataSetupStep {
	return func(ctx context.Context, t *testing.T, db *gorm.DB) context.Context {
		once.Do(func() {
			for _, step := range steps {
				ctx = step(ctx, t, db)
			}
		})
		return ctx
	}
}

// SetupWithGormScopes returns a DataSetupScope that applies given gorm scopes
func SetupWithGormScopes(scopes ...func(*gorm.DB) *gorm.DB) DataSetupScope {
	return func(ctx context.Context, t *testing.T, db *gorm.DB) (context.Context, *gorm.DB) {
		return ctx, db.Scopes(scopes...)
	}
}

var sqlStatementSep = regexp.MustCompile(`; *$`)

func execSqlFile(ctx context.Context, fsys fs.FS, db *gorm.DB, g *gomega.WithT, filename string) {
	file, e := fsys.Open(filename)
	g.Expect(e).To(Succeed(), "table preparation should be able to open SQL file '%s'", filename)

	queries, e := io.ReadAll(file)
	g.Expect(e).To(Succeed(), "table preparation should be able to read SQL file '%s'", filename)

	for _, q := range sqlStatementSep.Split(string(queries), -1) {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		r := db.WithContext(ctx).Exec(q)
		g.Expect(r.Error).To(Succeed(), "table preparation should be able to run SQL file '%s'", filename)
	}
}

func loadSeedData(fsys fs.FS, dest interface{}, filename string) (err error) {
	data, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, dest)
	return
}

func truncateTableSql(table string) string {
	return fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE;`, table)
}