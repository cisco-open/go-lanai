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
	"gorm.io/gorm"
	"io/fs"
)

type MigrationFunc func(ctx context.Context) error

type Registrar struct {
	migrationSteps []*Migration
	versions 	   utils.StringSet
}

func NewRegistrar() *Registrar {
	return &Registrar{}
}

func (r *Registrar) AddMigrations(m... *Migration) {
	r.migrationSteps = append(r.migrationSteps, m...)
}

type Migration struct {
	Version     Version
	Description string
	Func		MigrationFunc
	Tags        utils.StringSet
}

func WithVersion(version string) *Migration {
	v, err := fromString(version)
	if err != nil {
		panic(err)
	}
	return &Migration{
		Version: v,
	}
}

func (m *Migration) Dot(i int) *Migration {
	m.Version = append(m.Version, i)
	return m
}

func (m *Migration) WithTag(tags...string) *Migration {
	if m.Tags == nil {
		m.Tags = utils.NewStringSet(tags...)
	} else {
		m.Tags.Add(tags...)
	}
	return m
}

func (m *Migration) WithFile(fs fs.FS, filePath string, db *gorm.DB) *Migration {
	m.Func = migrationFuncFromTextFile(fs, filePath, db)
	return m
}

func (m *Migration) WithFunc(f MigrationFunc) *Migration {
	m.Func = f
	return m
}

func (m *Migration) WithDesc(d string) *Migration {
	m.Description = d
	return m
}