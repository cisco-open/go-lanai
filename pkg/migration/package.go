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
	TagPreUpgrade  = "pre_upgrade"
	TagPostUpgrade = "post_upgrade"
)

var logger = log.New("Migration")

var filterFlag string
var allowOutOfOrderFlag bool

var Module = &bootstrap.Module{
	Name:       "migration",
	Precedence: bootstrap.CommandLineRunnerPrecedence,
	Options: []fx.Option{
		fx.Provide(NewRegistrar),
		fx.Provide(NewGormVersioner),
		fx.Provide(provideMigrationRunner()),
	},
}

func Use() {
	bootstrap.AddStringFlag(&filterFlag, "filter", "", fmt.Sprintf("filter the migration steps by tag value. supports %s or %s", TagPreUpgrade, TagPostUpgrade))
	bootstrap.AddBoolFlag(&allowOutOfOrderFlag, "allow_out_of_order", false, fmt.Sprintf("allow migration steps to execute out of order"))
	bootstrap.Register(Module)
	// Note: migration CliRunner is provided in Module
	bootstrap.EnableCliRunnerMode()
}

func provideMigrationRunner() fx.Annotated {
	return fx.Annotated{
		Group:  bootstrap.FxCliRunnerGroup,
		Target: newMigrationRunner,
	}
}

func newMigrationRunner(r *Registrar, v Versioner, db *gorm.DB, dbCreator data.DbCreator) bootstrap.CliRunner {
	return func(ctx context.Context) error {
		if err := dbCreator.CreateDatabaseIfNotExist(ctx, db); err != nil {
			return err
		}
		return Migrate(ctx, r, v)
	}
}
