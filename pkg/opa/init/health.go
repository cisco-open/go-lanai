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

package opainit

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	"github.com/cisco-open/go-lanai/pkg/opa"
	"go.uber.org/fx"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	OPAReady        opa.EmbeddedOPAReadyCH
}

func RegisterHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&HealthIndicator{
		ready: di.OPAReady,
	})
}

type HealthIndicator struct {
	ready opa.EmbeddedOPAReadyCH
}

func (i *HealthIndicator) Name() string {
	return "opa"
}

func (i *HealthIndicator) Health(_ context.Context, _ health.Options) health.Health {
	select {
	case <-i.ready:
		return health.NewDetailedHealth(health.StatusUp, "OPA engine is UP", nil)
	default:
		return health.NewDetailedHealth(health.StatusDown, "OPA engine is not ready", nil)
	}
}
