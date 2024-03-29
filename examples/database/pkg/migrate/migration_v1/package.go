package migration_v1

import (
    "embed"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/migration"
    "go.uber.org/fx"
    "gorm.io/gorm"
)

//go:embed *.sql
var migrateFS embed.FS

func Use() {
	bootstrap.AddOptions(
		fx.Invoke(registerMigrations),
	)
}

func registerMigrations(r *migration.Registrar, db *gorm.DB, appCtx *bootstrap.ApplicationContext) {
	r.AddMigrations(
		migration.WithVersion("1.0.0.1").WithTag(migration.TagPreUpgrade).
			WithFile(migrateFS, "seed_initial_database.sql", db).
			WithDesc("Create DB schema"),
	)
}
