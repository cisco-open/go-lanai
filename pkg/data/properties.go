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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	ManagementPropertiesPrefix = "data"
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
	if err := ctx.Config().Bind(props, ManagementPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DataProperties"))
	}
	return *props
}
