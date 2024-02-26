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

package consulhealth

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"go.uber.org/fx"
)

type HealthIndicator struct {
	conn *consul.Connection
}

type HealthRegDI struct {
	fx.In
	HealthRegistrar health.Registrar   `optional:"true"`
	ConsulClient    *consul.Connection `optional:"true"`
}

func Register(di HealthRegDI) error {
	if di.HealthRegistrar == nil || di.ConsulClient == nil {
		return nil
	}
	return di.HealthRegistrar.Register(New(di.ConsulClient))
}

func New(conn *consul.Connection) *HealthIndicator {
	return &HealthIndicator{
		conn: conn,
	}
}

func (i *HealthIndicator) Name() string {
	return "consul"
}

func (i *HealthIndicator) Health(c context.Context, options health.Options) health.Health {

	if _, e := i.conn.Client().Status().Leader(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "consul leader status failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "consul leader status succeeded", nil)
	}
}
