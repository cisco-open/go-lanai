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
	"strings"
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

// SetupWithGormScopes returns a DataSetupScope that applies given gorm scopes
func SetupWithGormScopes(scopes ...func(*gorm.DB) *gorm.DB) DataSetupScope {
	return func(ctx context.Context, t *testing.T, db *gorm.DB) (context.Context, *gorm.DB) {
		return ctx, db.Scopes(scopes...)
	}
}

func execSqlFile(ctx context.Context, fsys fs.FS, db *gorm.DB, g *gomega.WithT, filename string) {
	file, e := fsys.Open(filename)
	g.Expect(e).To(Succeed(), "table preparation should be able to open SQL file '%s'", filename)

	queries, e := io.ReadAll(file)
	g.Expect(e).To(Succeed(), "table preparation should be able to read SQL file '%s'", filename)
	for _, q := range strings.Split(string(queries), ";") {
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