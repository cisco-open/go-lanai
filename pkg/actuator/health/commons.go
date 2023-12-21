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

package health

import (
	"context"
	"net/http"
)

/*******************************
	StaticStatusCodeMapper
********************************/

var DefaultStaticStatusCodeMapper = StaticStatusCodeMapper{
	StatusUp:           http.StatusOK,
	StatusDown:         http.StatusServiceUnavailable,
	StatusOutOfService: http.StatusServiceUnavailable,
	StatusUnknown:      http.StatusInternalServerError,
}

type StaticStatusCodeMapper map[Status]int

func (m StaticStatusCodeMapper) StatusCode(_ context.Context, status Status) int {
	if sc, ok := m[status]; ok {
		return sc
	}
	return http.StatusServiceUnavailable
}

/*******************************
	SimpleHealth
********************************/

// SimpleHealth implements Health
type SimpleHealth struct {
	Stat Status `json:"status"`
	Desc string `json:"description,omitempty"`
}

func (h SimpleHealth) Status() Status {
	return h.Stat
}

func (h SimpleHealth) Description() string {
	return h.Desc
}

/*******************************
	Composite
********************************/

// CompositeHealth implement Health
type CompositeHealth struct {
	SimpleHealth
	Components map[string]Health `json:"components,omitempty"`
}

func NewCompositeHealth(status Status, description string, components map[string]Health) *CompositeHealth {
	return &CompositeHealth{
		SimpleHealth: SimpleHealth{
			Stat: status,
			Desc: description,
		},
		Components: components,
	}
}

/*******************************
	DetailedHealth
********************************/

// DetailedHealth implement Health
type DetailedHealth struct {
	SimpleHealth
	Details map[string]interface{} `json:"details,omitempty"`
}

func NewDetailedHealth(status Status, description string, details map[string]interface{}) *DetailedHealth {
	return &DetailedHealth{
		SimpleHealth: SimpleHealth{
			Stat: status,
			Desc: description,
		},
		Details: details,
	}
}
