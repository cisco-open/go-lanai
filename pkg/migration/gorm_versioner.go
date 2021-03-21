package migration

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type MigrationVersion struct {
	Version       Version `gorm:"primaryKey"`
	Description   string
	ExecutionTime time.Duration
	InstalledOn   time.Time
	Success       bool
}

func (v MigrationVersion) GetVersion() Version {
	return v.Version
}

func (v MigrationVersion) GetDescription() string {
	return v.Description
}

func (v MigrationVersion) IsSuccess() bool {
	return v.Success
}

func (v MigrationVersion) GetInstalledOn() time.Time {
	return v.InstalledOn
}


type GormVersioner struct {
	db          *gorm.DB
}

func (v *GormVersioner) CreateVersionTableIfNotExist(ctx context.Context) error {
	return v.db.WithContext(ctx).AutoMigrate(&MigrationVersion{})
}

func (v *GormVersioner) GetAppliedMigrations(ctx context.Context) ([]AppliedMigration, error) {
	versions := []MigrationVersion{}
	result := v.db.WithContext(ctx).Find(&versions)
	if result.Error != nil {
		return nil, result.Error
	}

	retVersions := []AppliedMigration{}
	for _, ver := range versions {
		retVersions = append(retVersions, ver)
	}
	return retVersions, nil
}

func (v *GormVersioner) RecordAppliedMigration(ctx context.Context, version Version, description string, success bool, installedOn time.Time, executionTime time.Duration) error {
	applied := &MigrationVersion{
		Version:       version,
		Description:   description,
		Success:       success,
		InstalledOn:   installedOn,
		ExecutionTime: executionTime,
	}
	result := v.db.WithContext(ctx).Save(applied)
	return result.Error
}