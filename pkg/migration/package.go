package migration

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const (
	TagPreUpgrade = "pre_upgrade"
	TagPostUpgrade = "post_upgrade"
)

var logger = log.New("migration")

var filterFlag string
var allowOutOfOrderFlag bool

var Module = &bootstrap.Module{
	Name: "migration",
	Precedence: bootstrap.MigrationPrecedence,
	Options: []fx.Option{
		fx.Provide(newRegistrar),
		fx.Provide(newVersioner),
		fx.Invoke(applyMigrations),
	},
}

func Use() {
	bootstrap.AddStringFlag(&filterFlag, "filter", "", fmt.Sprintf("filter the migration steps by tag value. supports %s or %s", TagPreUpgrade, TagPostUpgrade))
	bootstrap.AddBoolFlag(&allowOutOfOrderFlag, "allow_out_of_order", false, fmt.Sprintf("allow migration steps to execute out of order"))
	bootstrap.Register(Module)
}

func newRegistrar() *Registrar {
	return &Registrar{}
}

func newVersioner(db *gorm.DB) Versioner {
	return &GormVersioner{
		db: db,
	}
}

func applyMigrations(lc fx.Lifecycle, r *Registrar, v Versioner, db *gorm.DB, dbCreator data.DbCreator, shutdowner fx.Shutdowner) {
	var err error
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err = dbCreator.CreateDatabaseIfNotExist(ctx, db)
			if err != nil {
				return shutdowner.Shutdown()
			}
			err = migrate(ctx, r, v)
			return shutdowner.Shutdown()
		},
		OnStop:  func(ctx context.Context) error {
			return err
		},
	})
}