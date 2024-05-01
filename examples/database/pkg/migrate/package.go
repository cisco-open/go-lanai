package migrate

import (
	"embed"
	"github.com/cisco-open/go-lanai/examples/skeleton-service/pkg/migrate/migration_v1"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/data"
	datainit "github.com/cisco-open/go-lanai/pkg/data/init"
	"github.com/cisco-open/go-lanai/pkg/data/postgresql/cockroach"
	"github.com/cisco-open/go-lanai/pkg/migration"
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
