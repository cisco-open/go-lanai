package migration

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type MigrationFunc func(ctx context.Context) error

type Registrar struct {
	migrationSteps []*Migration
	errs 		   []error
}

func (r *Registrar) AddMigration(version string, description string, migrationFunc MigrationFunc, tags... string) {
	v, err := NewVersion(version)

	if err != nil {
		r.errs = append(r.errs, err)
	}

	m := &Migration{
		Version:     v,
		Description: description,
		Func:        migrationFunc,
		Tags: utils.NewStringSet(tags...),
	}
	r.migrationSteps = append(r.migrationSteps, m)
}

type Migration struct {
	Version     Version
	Description string
	Func		MigrationFunc
	Tags        utils.StringSet
}

