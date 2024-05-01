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

package data

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/certs"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	PropertiesPrefix         = "data"
	DatabasePropertiesPrefix = "data.db"
)

type DataProperties struct {
	Logging     LoggingProperties     `json:"logging"`
	Transaction TransactionProperties `json:"transaction"`
}

type TransactionProperties struct {
	MaxRetry int `json:"max-retry"`
}

type LoggingProperties struct {
	Level         log.LoggingLevel `json:"level"`
	SlowThreshold utils.Duration   `json:"slow-threshold"`
}

// NewDataProperties create a DataProperties with default values
func NewDataProperties() *DataProperties {
	return &DataProperties{
		Logging: LoggingProperties{
			Level:         log.LevelWarn,
			SlowThreshold: utils.Duration(15 * time.Second),
		},
		Transaction: TransactionProperties{
			MaxRetry: 5,
		},
	}
}

// BindDataProperties create and bind SessionProperties, with a optional prefix
func BindDataProperties(ctx *bootstrap.ApplicationContext) DataProperties {
	props := NewDataProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DataProperties"))
	}
	return *props
}

type DatabaseProperties struct {
	//Enabled       bool                               `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SslMode  string `json:"sslmode"`
	Tls      TLS    `json:"tls"`
}

type TLS struct {
	Enable bool                   `json:"enabled"`
	Certs  certs.SourceProperties `json:"certs"`
}

// NewDatabaseProperties create a DatabaseProperties with default values
func NewDatabaseProperties() *DatabaseProperties {
	return &DatabaseProperties{
		Host:     "localhost",
		Port:     26257,
		Username: "root",
		Password: "root",
		SslMode:  "disable",
	}
}

// BindDatabaseProperties create and bind DatabaseProperties
func BindDatabaseProperties(ctx *bootstrap.ApplicationContext) DatabaseProperties {
	props := NewDatabaseProperties()
	if err := ctx.Config().Bind(props, DatabasePropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DatabaseProperties"))
	}
	return *props
}
