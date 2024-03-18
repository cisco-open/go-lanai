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

package tracing

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	TracingPropertiesPrefix = "tracing"
)

type TracingProperties struct {
	Enabled bool              `json:"enabled"`
	Jaeger  JaegerProperties  `json:"jaeger"`
	Zipkin  ZipkinProperties  `json:"zipkin"`
	Sampler SamplerProperties `json:"sampler"`
}

type JaegerProperties struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

type ZipkinProperties struct {
	Enabled bool `json:"enabled"`
}

type SamplerProperties struct {
	Enabled     bool    `json:"enabled"`
	RateLimit   float64 `json:"limit-per-second"`
	Probability float64 `json:"probability"`
	LowestRate  float64 `json:"lowest-per-second"`
}

//NewSessionProperties create a SessionProperties with default values
func NewTracingProperties() *TracingProperties {
	return &TracingProperties{
		Enabled: true,
		Jaeger: JaegerProperties{
			Enabled: true,
		},
		Zipkin: ZipkinProperties{},
		Sampler: SamplerProperties{
			Enabled:   false,
			RateLimit: 10.0,
		},
	}
}

// BindTracingProperties create and bind SessionProperties, with a optional prefix
func BindTracingProperties(ctx *bootstrap.ApplicationContext) TracingProperties {
	props := NewTracingProperties()
	if err := ctx.Config().Bind(props, TracingPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind TracingProperties"))
	}
	return *props
}
