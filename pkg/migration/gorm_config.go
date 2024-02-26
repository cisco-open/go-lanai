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
	"github.com/cisco-open/go-lanai/pkg/data"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type gormConfigurer struct {}

func DefaultGormConfigurerProvider() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: newGormMigrationConfigurer,
	}
}

func newGormMigrationConfigurer() data.GormConfigurer {
	return &gormConfigurer{}
}

func (c gormConfigurer) Order() int {
	return 0
}

func (c gormConfigurer) Configure(config *gorm.Config) {
	config.DisableForeignKeyConstraintWhenMigrating = true
	config.FullSaveAssociations = false
	config.SkipDefaultTransaction = true
	config.CreateBatchSize = 1000
}