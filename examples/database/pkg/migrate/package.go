package migrate

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/examples/skeleton-service/pkg/migrate/migration_v1"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/cockroach"
	datainit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/migration"
	"embed"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

//go:embed application-migrate.yml
var migrateConfigFS embed.FS

func Use() {
	appconfig.Use()
	//consul.Use()
	//vault.Use()
	datainit.Use()
	cockroach.Use()
	migration.Use()
	migration_v1.Use()
	bootstrap.AddOptions(
		appconfig.FxEmbeddedApplicationAdHoc(migrateConfigFS),
		fx.Provide(migrationGormConfigurer()),
	)
}

type gormConfigurer struct{}

func migrationGormConfigurer() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: newGormConfigurer,
	}
}

func newGormConfigurer() data.GormConfigurer {
	return &gormConfigurer{}
}

func (c gormConfigurer) Configure(config *gorm.Config) {
	config.CreateBatchSize = 250
}
