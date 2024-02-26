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

package actuator

import (
    "embed"
    "github.com/cisco-open/go-lanai/pkg/actuator"
    "github.com/cisco-open/go-lanai/pkg/actuator/alive"
    "github.com/cisco-open/go-lanai/pkg/actuator/apilist"
    "github.com/cisco-open/go-lanai/pkg/actuator/env"
    health "github.com/cisco-open/go-lanai/pkg/actuator/health/endpoint"
    "github.com/cisco-open/go-lanai/pkg/actuator/info"
    "github.com/cisco-open/go-lanai/pkg/actuator/loggers"
    appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "go.uber.org/fx"
)

//go:embed defaults-actuator.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "actuate-config",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
	},
}

func Use() {
	bootstrap.Register(actuator.Module)
	bootstrap.Register(Module)
	info.Register()
	health.Register()
	env.Register()
	alive.Register()
	apilist.Register()
	loggers.Register()
}

/**************************
	Initialize
***************************/
