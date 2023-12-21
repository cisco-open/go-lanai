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

package vaulthealth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"go.uber.org/fx"
)

type HealthRegDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	VaultClient     *vault.Client    `optional:"true"`
}

func Register(di HealthRegDI) error {
	if di.HealthRegistrar == nil || di.VaultClient == nil {
		return nil
	}
	return di.HealthRegistrar.Register(New(di.VaultClient))
}

func New(client *vault.Client) *HealthIndicator {
	return &HealthIndicator{Client: client}
}

type HealthIndicator struct {
	Client *vault.Client
}

func (i *HealthIndicator) Name() string {
	return "vault"
}

func (i *HealthIndicator) Health(c context.Context, options health.Options) health.Health {
	if _, e := i.Client.Sys(c).Health(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "vault /v1/sys/health failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "vault /v1/sys/health succeeded", nil)
	}
}

