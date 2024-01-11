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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
	"sort"
	"time"
)

func Migrate(ctx context.Context, r *Registrar, v Versioner) error {
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
	sort.SliceStable(appliedMigrations, func (i, j int) bool {return appliedMigrations[i].GetVersion().Lt(appliedMigrations[j].GetVersion())})

	for _, a := range appliedMigrations {
		if !a.IsSuccess() {
			return errors.New(fmt.Sprintf("stopping migration because there is a failed migration step: %s", a.GetVersion().String()))
		}
	}

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
			migrationErr := s.Func(ctx) //TODO: manual rollback function?
			finishTime := time.Now()
			duration := finishTime.Sub(startTime)
			if migrationErr != nil {
				err = v.RecordAppliedMigration(ctx, s.Version, s.Description, false, finishTime, duration)
				if err != nil {
					logger.Errorf("error recording failed migration version due to %v", err)
				}
				err = errors.New(fmt.Sprintf("migration stopped at step %v because of error: %v", s.Version, migrationErr))
				logger.Errorf("%v", err)
				return err
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
