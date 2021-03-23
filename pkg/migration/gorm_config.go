package migration

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type gormConfigurer struct {}

func DefaultGormConfigurerProvider() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: newGormMigrationConfigurer,
	}
}

func newGormMigrationConfigurer() data.GormConfigurer {
	return &gormConfigurer{}
}

func (c gormConfigurer) Order() int {
	return 0
}

func (c gormConfigurer) Configure(config *gorm.Config) {
	config.DisableForeignKeyConstraintWhenMigrating = true
	config.FullSaveAssociations = false
	config.SkipDefaultTransaction = true
	config.CreateBatchSize = 1000
}