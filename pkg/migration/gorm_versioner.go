// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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

func NewGormVersioner(db *gorm.DB) Versioner {
	return &GormVersioner{
		db: db,
	}
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