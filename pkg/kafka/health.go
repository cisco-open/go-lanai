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

package kafka

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
)

type HealthIndicator struct {
	binder SaramaBinder
}

func NewHealthIndicator(binder Binder) *HealthIndicator {
	return &HealthIndicator{binder: binder.(SaramaBinder)}
}

func (i *HealthIndicator) Name() string {
	return "kafka"
}

func (i *HealthIndicator) Health(_ context.Context, opts health.Options) health.Health {
	topics := i.binder.ListTopics()

	client := i.binder.Client()
	if client == nil {
		return health.NewDetailedHealth(health.StatusUnknown, "kafka client not initialized yet", nil)
	}

	var details map[string]interface{}
	if opts.ShowDetails {
		details = map[string]interface{}{
			"topics": topics,
		}
	}

	if err := client.RefreshMetadata(topics...); err != nil {
		return health.NewDetailedHealth(health.StatusDown, "kafka refresh metadata failed", details)
	}
	return health.NewDetailedHealth(health.StatusUp, "kafka refresh metadata succeeded", details)
}
