package migration

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
	"sort"
	"time"
)

func migrate(ctx context.Context, r *Registrar, v Versioner) error {
	if len(r.errs) != 0 {
		logger.Errorf("encountered error when registering migration steps")
		for _, e := range r.errs {
			logger.Errorf("error: %v", e)
		}
		return errors.New("migrations not applied because there were error with migration steps.")
	}

	err := v.CreateVersionTableIfNotExist(ctx)
	if err != nil {
		return err
	}

	//sort registered migration steps
	sort.SliceStable(r.migrationSteps, func(i, j int) bool {return r.migrationSteps[i].Version.Lt(r.migrationSteps[j].Version)})

	appliedMigrations, err := v.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	//sort applied migration steps
	sort.SliceStable(appliedMigrations, func (i, j int) bool {return appliedMigrations[i].GetVersion().Lt(appliedMigrations[i].GetVersion())})

	var shouldExecuteMigration func(*Migration) bool

	if allowOutOfOrderFlag {
		appliedVersions := utils.NewStringSet()
		for _, a := range appliedMigrations {
			appliedVersions.Add(a.GetVersion().String())
		}
		shouldExecuteMigration = func(m *Migration) bool {
			return !appliedVersions.Has(m.Version.String())
		}
	} else {
		var lastAppliedMigration AppliedMigration
		if len(appliedMigrations) > 0 {
			lastAppliedMigration = appliedMigrations[len(appliedMigrations)-1]
		}
		shouldExecuteMigration = func(m *Migration) bool {
			return lastAppliedMigration == nil || lastAppliedMigration.GetVersion().Lt(m.Version)
		}
	}

	for _, s := range r.migrationSteps {
		if filterFlag != "" && !s.Tags.Has(filterFlag) {
			continue
		}
		//TODO: should the migration func and recording the version be put in one transaction?
		if shouldExecuteMigration(s) {
			logger.Infof("Executing migration step %s: %s", s.Version.String(), s.Description)
			startTime := time.Now()
			err = s.Func(ctx) //TODO: manual rollback function?
			finishTime := time.Now()
			duration := finishTime.Sub(startTime)
			if err != nil {
				err = v.RecordAppliedMigration(ctx, s.Version, s.Description, false, finishTime, duration)
				if err != nil {
					logger.Errorf("error recording failed migration version due to %v", err)
				}
				return errors.New(fmt.Sprintf("migration stoped because error at step %v", s.Version))
			} else {
				err = v.RecordAppliedMigration(ctx, s.Version, s.Description, true, finishTime, duration)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
