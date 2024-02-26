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

package web

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/pkg/errors"
)

/***********************
	Server
************************/

const (
	ServerPropertiesPrefix = "server"
)

type ServerProperties struct {
	Port        int               `json:"port"`
	ContextPath string            `json:"context-path"`
	Logging     LoggingProperties `json:"logging"`
}

type LoggingProperties struct {
	Enabled      bool                              `json:"enabled"`
	DefaultLevel log.LoggingLevel                  `json:"default-level"`
	Levels       map[string]LoggingLevelProperties `json:"levels"`
}

// LoggingLevelProperties is used to override logging level on particular set of paths
// the LoggingProperties.Pattern support wildcard and should not include "context-path"
// the LoggingProperties.Method is space separated values. If left blank or contains "*", it matches all methods
type LoggingLevelProperties struct {
	Method  string           `json:"method"`
	Pattern string           `json:"pattern"`
	Level   log.LoggingLevel `json:"level"`
}

// NewServerProperties create a ServerProperties with default values
func NewServerProperties() *ServerProperties {
	return &ServerProperties{
		Port:        -1,
		ContextPath: "/",
		Logging: LoggingProperties{
			Enabled:      true,
			DefaultLevel: log.LevelDebug,
			Levels:       map[string]LoggingLevelProperties{},
		},
	}
}

// BindServerProperties create and bind a ServerProperties using default prefix
func BindServerProperties(ctx *bootstrap.ApplicationContext) ServerProperties {
	props := NewServerProperties()
	if err := ctx.Config().Bind(props, ServerPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ServerProperties"))
	}
	return *props
}
