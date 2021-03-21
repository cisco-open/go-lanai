package cockroach

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"fmt"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
)

const (
	dsKeyHost     = "host"
	dsKeyPort     = "port"
	dsKeyDB       = "dbname"
	dsKeySslMode  = "sslmode"
	dsKeyUsername = "user"
	dsKeyPassword = "password"
)

type initDI struct {
	fx.In
	AppContext *bootstrap.ApplicationContext
	Properties CockroachProperties
}

func NewGormDialetor(di initDI) gorm.Dialector {
	//"host=localhost user=root password=root dbname=idm port=26257 sslmode=disable"
	options := map[string]interface{}{
		dsKeyHost: di.Properties.Host,
		dsKeyPort: di.Properties.Port,
		dsKeyDB: di.Properties.Database,
		dsKeySslMode: "disable",
	}

	if di.Properties.Username != "" {
		options[dsKeyUsername] = di.Properties.Username
		options[dsKeyPassword] = di.Properties.Password
	}

	config := postgres.Config{
		//DriverName:           "postgres",
		DSN:                  toDSN(options),
	}
	return NewGormDialectorWithConfig(config)
}

func toDSN(options map[string]interface{}) string {
	opts := []string{}
	for k, v := range options {
		opt := fmt.Sprintf("%s=%v", k, v)
		opts = append(opts, opt)
	}
	return strings.Join(opts, " ")
}

