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

// Package healthep
// Contains implementation of health endpoint as a separate package to avoid cyclic package dependency.
//
// Implementations in this package cannot be moved to package "actuator/health", otherwise, it could create
// cyclic package dependency as following:
// 		actuator/health -> actuator -> security -> tenancy -> redis -> actuator/health
//
// Therefore, any implementations involves package mentioned above should be moved here
package healthep

import (
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "actuator-health-ep",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func Register() {
	health.Use()
	bootstrap.Register(Module)
}

type regDI struct {
	fx.In
	Properties      health.HealthProperties
	HealthRegistrar health.Registrar
	Registrar       *actuator.Registrar           `optional:"true"`
	MgtProperties   actuator.ManagementProperties `optional:"true"`
}

func register(di regDI) {
	// Note: when actuator.Registrar is nil, we don't need to anything
	if di.Registrar == nil {
		return
	}
	healthReg := di.HealthRegistrar.(*health.SystemHealthRegistrar)
	endpoint, e := newEndpoint(func(opt *EndpointOption) {
		opt.MgtProperties = di.MgtProperties
		opt.Contributor = healthReg.Indicator
		opt.Properties = di.Properties
		opt.DetailsControl = healthReg.DetailsDisclosure
		opt.ComponentsControl = healthReg.ComponentsDisclosure
	})
	if e != nil {
		panic(e)
	}

	di.Registrar.MustRegister(endpoint)
}

