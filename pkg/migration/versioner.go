package migration

import "time"

type AppliedMigration struct {
	Version     Version
	Description string
	Success bool
	InstalledOn time.Time
}

type Versioner interface {
	CreateVersionTableIfNotExist()
	GetAppliedMigrations() []AppliedMigration
	RecordAppliedMigration(migration AppliedMigration)
}

type CockroachDbVersioner struct {}

func (c *CockroachDbVersioner) CreateVersionTableIfNotExist() {
}

func (c *CockroachDbVersioner) GetAppliedMigrations() []AppliedMigration {
	return nil
}

func (c *CockroachDbVersioner) RecordAppliedMigration(migration AppliedMigration) {
}



