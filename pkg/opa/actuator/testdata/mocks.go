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

package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
)

const SpecialScopeAdmin = `admin`

type MockedHealthIndicator struct {
	Status      health.Status
	Description string
	Details     map[string]interface{}
}

func NewMockedHealthIndicator() *MockedHealthIndicator {
	return &MockedHealthIndicator{
		Status: health.StatusUp,
		Description: "mocked",
		Details:     map[string]interface{}{
			"key": "value",
		},
	}
}

func (i *MockedHealthIndicator) Name() string {
	return "test"
}

func (i *MockedHealthIndicator) Health(_ context.Context, opts health.Options) health.Health {
	ret := health.CompositeHealth{
		SimpleHealth: health.SimpleHealth{
			Stat: i.Status,
			Desc: i.Description,
		},
	}
	if opts.ShowComponents {
		detailed := health.DetailedHealth{
			SimpleHealth: health.SimpleHealth{
				Stat: i.Status,
				Desc: "mocked detailed",
			},
		}
		if opts.ShowDetails {
			detailed.Details = i.Details
		}

		ret.Components = map[string]health.Health{
			"simple": health.SimpleHealth{
				Stat: i.Status,
				Desc: "mocked simple",
			},
			"detailed": detailed,
		}
	}
	return ret
}
