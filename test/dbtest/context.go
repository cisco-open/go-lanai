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

package dbtest

import (
	"context"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*************************
	Enums
 *************************/

const (
	modeAuto mode = iota
	modePlayback
	modeRecord
)

type mode int

/*************************
	DBOptions
 *************************/

type DBOptions func(opt *DBOption)
type DBOption struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
	SSL      bool
}

func DBName(db string) DBOptions {
	return func(opt *DBOption) {
		opt.DBName = db
	}
}

func DBCredentials(user, password string) DBOptions {
	return func(opt *DBOption) {
		opt.Username = user
		opt.Password = password
	}
}

func DBPort(port int) DBOptions {
	return func(opt *DBOption) {
		opt.Port = port
	}
}

func DBHost(host string) DBOptions {
	return func(opt *DBOption) {
		opt.Host = host
	}
}

/*************************
	TX context
 *************************/

type mockedTxContext struct {
	context.Context
}

func (c mockedTxContext) Parent() context.Context {
	return c.Context
}

type mockedGormContext struct {
	mockedTxContext
	db *gorm.DB
}

func (c mockedGormContext) DB() *gorm.DB {
	return c.db
}

/*************************
	Data Setup
 *************************/

type DI struct {
	fx.In
	DB *gorm.DB
}

type DataSetupStep func(ctx context.Context, t *testing.T, db *gorm.DB) context.Context
type DataSetupScope func(ctx context.Context, t *testing.T, db *gorm.DB) (context.Context, *gorm.DB)
