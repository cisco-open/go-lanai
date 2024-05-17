package cockroach

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

const dbName = "new_test_db"

func TestDBCreator(t *testing.T) {
	di := &dbtest.DI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithDI(di),
		test.SubTestTeardown(TearDownWithDropDatabase(di)),
		test.GomegaSubTest(SubTestCreateDB(di), "TestCreateDB"),
	)
}

func TearDownWithDropDatabase(di *dbtest.DI) test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		// Drop the table
		r := di.DB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", di.DB.Statement.Quote(dbName)))
		return r.Error
	}
}

func SubTestCreateDB(di *dbtest.DI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := data.DataProperties{
			DB: data.DatabaseProperties{
				Database: dbName,
			},
		}
		dbCreator := NewGormDbCreator(p)

		// Create the database
		err := dbCreator.CreateDatabaseIfNotExist(ctx, di.DB)
		g.Expect(err).To(Not(HaveOccurred()), "CreateDatabaseIfNotExist should not return error")

		r := di.DB.Raw("show databases;")
		g.Expect(r.Error).To(Not(HaveOccurred()), "show databases should not return error")

		found := false
		var records []DBRecord
		r.Scan(&records)
		for _, record := range records {
			if record.DbName == dbName {
				found = true
				break
			}
		}
		g.Expect(found).To(BeTrue(), "Database %s should be created", dbName)
	}
}

// This test expects the test database has a testuser with no create database permission
func TestDBCreatorWithoutCreateDBPermission(t *testing.T) {
	di := &dbtest.DI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb", dbtest.DBCredentials("testuser", "")),
		apptest.WithDI(di),
		apptest.WithFxOptions(),
		test.SubTestTeardown(TearDownWithDropDatabase(di)),
		test.GomegaSubTest(SubTestCreateDBIgnoreFailure(di), "TestCreateDBIgnoreFailure"),
	)
}

func SubTestCreateDBIgnoreFailure(di *dbtest.DI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := data.DataProperties{
			DB: data.DatabaseProperties{
				Database: dbName,
			},
		}
		dbCreator := NewGormDbCreator(p)

		// Create the database
		err := dbCreator.CreateDatabaseIfNotExist(ctx, di.DB)
		g.Expect(err).To(Not(HaveOccurred()), "CreateDatabaseIfNotExist should not return error")
	}
}

type DBRecord struct {
	DbName string `gorm:"column:database_name;type:text;not null;"`
}
