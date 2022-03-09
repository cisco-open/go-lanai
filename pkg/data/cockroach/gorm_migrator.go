package cockroach

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

/*************************
	Custom Migrator
 *************************/

// GormMigrator
// Inverted index support:
// for now, use PostgreSQL-compatible syntax: https://www.cockroachlabs.com/docs/v20.2/inverted-indexes#creation
type GormMigrator struct {
	postgres.Migrator
}

func NewGormMigrator(db *gorm.DB, dialector gorm.Dialector) *GormMigrator {
	return &GormMigrator{
		Migrator: postgres.Migrator{
			Migrator: migrator.Migrator{
				Config: migrator.Config{
					DB:                          db,
					Dialector:                   dialector,
					CreateIndexAfterCreateTable: true,
				},
			},
		},
	}
}

