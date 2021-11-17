package dbtest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/cockroach"
	"fmt"
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
			return cockroach.PqErrorTranslator{}
		},
	}
}

func gormErrTranslatorProvider() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: func() data.ErrorTranslator {
			return data.NewGormErrorTranslator()
		},
	}
}


