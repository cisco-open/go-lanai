package cockroach

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*************************
	Custom GormDialector
 *************************/
type GormDialector struct {
	postgres.Dialector
}

func NewGormDialectorWithConfig(config postgres.Config) *GormDialector {
	return &GormDialector{
		Dialector: *postgres.New(config).(*postgres.Dialector),
	}
}

func (d GormDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return NewGormMigrator(db, d)
}


