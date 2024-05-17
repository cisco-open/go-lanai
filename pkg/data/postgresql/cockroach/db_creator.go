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
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/pkg/data/postgresql"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type GormDbCreator struct {
	dbName string
}

func (g *GormDbCreator) Order() int {
	return postgresql.DBCreatorPostgresOrder - 1
}

func (g *GormDbCreator) CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error {
	result := db.WithContext(ctx).Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", db.Statement.Quote(g.dbName)))
	if result.Error != nil && errors.Is(result.Error, data.ErrorInsufficientPrivilege) {
		logger.Warnf("Skipped creating database because %v", result.Error)
		return nil
	}
	return result.Error
}

func NewGormDbCreator(properties data.DataProperties) data.DbCreator {
	return &GormDbCreator{
		dbName: properties.DB.Database,
	}
}

func newAnnotatedGormDbCreator() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: NewGormDbCreator,
	}
}
