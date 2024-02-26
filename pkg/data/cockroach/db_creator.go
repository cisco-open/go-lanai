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

package cockroach

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/data"
    "gorm.io/gorm"
)

type GormDbCreator struct {
	dbUser string
	dbName string
}

func NewGormDbCreator(properties CockroachProperties) data.DbCreator {
	return &GormDbCreator{
		dbUser: properties.Username,
		dbName: properties.Database,
	}
}

func (g *GormDbCreator) CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error {
	if g.dbUser != "root" {
		logger.WithContext(ctx).Info("db user is not a privileged account, skipped db creation.")
		return nil
	}
	result := db.WithContext(ctx).Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", g.dbName))
	return result.Error

}
