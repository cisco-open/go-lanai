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

package acm

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/acm"
    "github.com/cisco-open/go-lanai/pkg/actuator/health"
    "go.uber.org/fx"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	AcmClient       *acm.Client
}

func RegisterHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&HealthIndicator{
		AcmClient: di.AcmClient,
	})
}

// HealthIndicator monitor ACM client status
type HealthIndicator struct {
	AcmClient *acm.Client
}

func (i *HealthIndicator) Name() string {
	return "aws.acm"
}

func (i *HealthIndicator) Health(ctx context.Context, _ health.Options) health.Health {
	input := &acm.GetAccountConfigurationInput{}
	if _, e := i.AcmClient.GetAccountConfiguration(ctx, input); e != nil {
		logger.WithContext(ctx).Warnf("AWS ACM connection not available or identity invalid: %v", e)
		return health.NewDetailedHealth(health.StatusUnknown, "AWS ACM connection not available or identity invalid", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "aws connect succeeded", nil)
	}
}
