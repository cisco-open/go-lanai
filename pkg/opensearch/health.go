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

package opensearch

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
)

type HealthIndicator struct {
	client OpenClient
}

func (i *HealthIndicator) Name() string {
	return "opensearch"
}

func NewHealthIndicator(client OpenClient) *HealthIndicator {
	return &HealthIndicator{
		client: client,
	}
}

func (i *HealthIndicator) Health(c context.Context, options health.Options) health.Health {
	resp, err := i.client.Ping(c)
	if err != nil {
		logger.WithContext(c).Errorf("unable to ping opensearch: %v", err)
		return health.NewDetailedHealth(health.StatusDown, "opensearch ping failed", nil)
	}
	if resp.IsError() {
		logger.WithContext(c).Errorf("unable to ping opensearch: %v", resp.String())
		return health.NewDetailedHealth(health.StatusDown, "opensearch ping failed", nil)
	}
	return health.NewDetailedHealth(health.StatusUp, "opensearch ping succeeded", nil)
}
