package migration

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"sort"
	"time"
)

const (
	TagPreUpgrade = "pre_upgrade"
	TagPostUpgrade = "post_upgrade"
)

var logger = log.New("migration")

var filterFlag string

var Module = &bootstrap.Module{
	Name: "migration",
	Precedence: bootstrap.MigrationPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(newRegistrar),
		fx.Provide(newVersioner),
		fx.Invoke(applyMigrations),
	},
}

func init() {
	bootstrap.AddStringFlag(&filterFlag, "filter", "", fmt.Sprintf("filter the migration steps by tag value. supports %s or %s", TagPreUpgrade, TagPostUpgrade))
	bootstrap.Register(Module)
}

func Use() {

}

func newRegistrar() *Registrar {
	return &Registrar{
		errs: make([]error, 0),
	}
}

func newVersioner() Versioner {
	return &CockroachDbVersioner{}
}

func applyMigrations(lc fx.Lifecycle, c *bootstrap.ApplicationContext, r *Registrar, v Versioner, shutdowner fx.Shutdowner) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infof("filter %s", filterFlag)

			v.CreateVersionTableIfNotExist()
			appliedMigrations := v.GetAppliedMigrations()
			var lastAppliedMigration AppliedMigration

			if len(appliedMigrations) > 0 {
				lastAppliedMigration = appliedMigrations[len(appliedMigrations)-1]
			}

			logger.Infof("applying migrations")
			if len(r.errs) != 0 {
				logger.Errorf("encountered error when registering migration steps")
				for _, e := range r.errs {
					logger.Errorf("error: %v", e)
				}
				return errors.New("migrations not applied because there were error with migration steps.")
			}

			sort.Slice(r.migrationSteps, func(i, j int) bool {return r.migrationSteps[i].Version.Lt(r.migrationSteps[j].Version)})

			for _, s := range r.migrationSteps {
				if filterFlag != "" && !s.Tags.Has(filterFlag) {
					continue
				}

				if lastAppliedMigration.Version.Lt(s.Version) {
					err := s.Func(c)
					applied := AppliedMigration{
						Version: s.Version,
						Description: s.Description,
						InstalledOn: time.Now(),
					}
					if err != nil {
						applied.Success = false
						v.RecordAppliedMigration(applied)
						return errors.New(fmt.Sprintf("migration stoped because error at step %v", s.Version))
					} else {
						applied.Success = true
						v.RecordAppliedMigration(applied)
					}
				}
			}
			return shutdowner.Shutdown()
		},
		OnStop: nil,
	})
}