package migration

import (
	"context"
	"time"
)

type AppliedMigration interface {
	GetVersion()     Version
	GetDescription() string
	IsSuccess() bool
	GetInstalledOn() time.Time //TODO: other information
}

type Versioner interface {
	CreateVersionTableIfNotExist(ctx context.Context) error
	GetAppliedMigrations(ctx context.Context) ([]AppliedMigration, error)
	RecordAppliedMigration(ctx context.Context, version Version, description string, success bool, installedOn time.Time, executionTime time.Duration) error
}





