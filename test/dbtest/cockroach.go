package dbtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/cockroach"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"flag"
	"fmt"
	"github.com/cockroachdb/copyist"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"strings"
	"testing"
)

const (
	ctxKeyCopyistCloser   = "copyist-closer"
	flagCopyistRecordMode = "record"
)

const (
	dsKeyHost     = "host"
	dsKeyPort     = "port"
	dsKeyDB       = "dbname"
	dsKeySslMode  = "sslmode"
	dsKeyUsername = "user"
	dsKeyPassword = "password"
)

type DBOptions func(opt *DBOption)
type DBOption struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
	SSL      bool
}

const (
	ModeAuto Mode = iota
	ModePlayback
	ModeRecord
)

type Mode int

// CockroachDB start do necessary setup to start mocked CockroachDB
// Note: DBName(...) option is required for ModeRecord
func CockroachDB(mode Mode, dbName string, opts ...DBOptions) suitetest.PackageOptions {
	defer func() { dbInitialized = true }()
	switch mode {
	case ModePlayback:
		mustSetFlag(flagCopyistRecordMode, "false")
	case ModeRecord:
		mustSetFlag(flagCopyistRecordMode, "true")
	}

	// prepare options
	opt := DBOption{
		Host:     "127.0.0.1",
		Port:     26257,
		DBName:   dbName,
		Username: "root",
	}
	for _, fn := range opts {
		fn(&opt)
	}

	return suitetest.WithOptions(
		suitetest.SetupWithOrder(orderCockroachMock, initializePostgresMock()),
		suitetest.TestOptions(
			test.Setup(openCopyistConn(&opt)),
			test.Teardown(closeCopyistConn()),
			apptest.WithFxOptions(
				fx.Provide(cockroach.BindCockroachProperties),
				fx.Provide(testGormDialetorProvider(&opt)),
			),
			apptest.WithProperties(
				fmt.Sprintf("data.cockroach.host: %s", opt.Host),
				fmt.Sprintf("data.cockroach.port: %d", opt.Port),
				fmt.Sprintf("data.cockroach.database: %s", opt.DBName),
				fmt.Sprintf("data.cockroach.username: %s", opt.Username),
				fmt.Sprintf("data.cockroach.password: %s", opt.Password),
			),
		),
	)
}

func DBName(db string) DBOptions {
	return func(opt *DBOption) {
		opt.DBName = db
	}
}

func DBCredentials(user, password string) DBOptions {
	return func(opt *DBOption) {
		opt.Username = user
		opt.Password = password
	}
}

func DBPort(port int) DBOptions {
	return func(opt *DBOption) {
		opt.Port = port
	}
}

func DBHost(host string) DBOptions {
	return func(opt *DBOption) {
		opt.Host = host
	}
}

func initializePostgresMock() suitetest.SetupFunc {
	return func() error {
		copyist.Register("postgres")
		return nil
	}
}

func openCopyistConn(opt *DBOption) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		if opt.DBName == "" {
			return nil, fmt.Errorf("DBName is required for recording mode")
		}

		closer := copyist.Open(t)
		if closer == nil {
			return ctx, nil
		}
		return context.WithValue(ctx, ctxKeyCopyistCloser, closer), nil
	}
}

func closeCopyistConn() test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		switch v := ctx.Value(ctxKeyCopyistCloser).(type) {
		case io.Closer:
			return v.Close()
		}
		return nil
	}
}

type mockDI struct {
	fx.In
}

func testGormDialetorProvider(opt *DBOption) func(di mockDI) gorm.Dialector {
	return func(di mockDI) gorm.Dialector {
		//"host=localhost user=root password=root dbname=idm port=26257 ssl=disable"
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
		return cockroach.NewGormDialectorWithConfig(config)
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

func mustSetFlag(name, value string) {
	e := flag.Set(name, value)
	if e != nil {
		panic(e)
	}
}
