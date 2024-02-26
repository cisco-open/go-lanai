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

package healthep

import (
    "context"
    "encoding/json"
    "github.com/cisco-open/go-lanai/pkg/actuator"
    "github.com/cisco-open/go-lanai/pkg/actuator/health"
)

const (
	ID              = "health"
	EnableByDefault = true
)

type Input struct{}

type Output struct {
	health.Health
	sc int
}

type CompositeHealthV2 struct {
	health.SimpleHealth
	Components map[string]health.Health `json:"details,omitempty"`
}

// StatusCode http.StatusCoder
func (o Output) StatusCode() int {
	return o.sc
}

// Body web.BodyContainer
func (o Output) Body() interface{} {
	return o.Health
}

// MarshalJSON json.Marshaler
func (o Output) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Health)
}

type EndpointOptions func(opt *EndpointOption)
type EndpointOption struct {
	Contributor       health.Indicator
	StatusCodeMapper  health.StatusCodeMapper
	MgtProperties     actuator.ManagementProperties
	Properties        health.HealthProperties
	DetailsControl    health.DetailsDisclosureControl
	ComponentsControl health.ComponentsDisclosureControl
}

// HealthEndpoint implements actuator.Endpoint, actuator.WebEndpoint
type HealthEndpoint struct {
	actuator.WebEndpointBase
	contributor       health.Indicator
	scMapper          health.StatusCodeMapper
	detailsControl    health.DetailsDisclosureControl
	componentsControl health.ComponentsDisclosureControl
}

func newEndpoint(opts ...EndpointOptions) (*HealthEndpoint, error) {
	opt := EndpointOption{}
	for _, f := range opts {
		f(&opt)
	}

	if opt.StatusCodeMapper == nil {
		scMapper := health.DefaultStaticStatusCodeMapper
		for k, v := range opt.Properties.Status.ScMapping {
			scMapper[k] = v
		}
		opt.StatusCodeMapper = scMapper
	}

	disclosureCtrl, e := newDefaultDisclosureControl(&opt.Properties, opt.DetailsControl, opt.ComponentsControl)
	if e != nil {
		return nil, e
	}

	ep := HealthEndpoint{
		contributor:       opt.Contributor,
		scMapper:          opt.StatusCodeMapper,
		detailsControl:    disclosureCtrl,
		componentsControl: disclosureCtrl,
	}

	properties := opt.MgtProperties
	ep.WebEndpointBase = actuator.MakeWebEndpointBase(func(opt *actuator.EndpointOption) {
		opt.Id = ID
		opt.Ops = []actuator.Operation{
			actuator.NewReadOperation(ep.Read),
		}
		opt.Properties = &properties.Endpoints
		opt.EnabledByDefault = EnableByDefault
	})

	return &ep, nil
}

// Read never returns error
func (ep *HealthEndpoint) Read(ctx context.Context, _ *Input) (*Output, error) {
	opts := health.Options{
		ShowDetails:    ep.detailsControl.ShouldShowDetails(ctx),
		ShowComponents: ep.componentsControl.ShouldShowComponents(ctx),
	}
	h := ep.contributor.Health(ctx, opts)
	switch f := ep.WebEndpointBase.NegotiateFormat(ctx); f {
	case actuator.ContentTypeSpringBootV2:
		h = ep.toSpringBootV2(h)
	}

	// Note: we know that *SystemHealthInitializer respect options (as all CompositeIndicator)
	// we don't need to sanitize result
	return &Output{
		Health: h,
		sc:     ep.scMapper.StatusCode(ctx, h.Status()),
	}, nil
}

func (ep *HealthEndpoint) toSpringBootV2(h health.Health) health.Health {
	var composite *health.CompositeHealth
	switch v := h.(type) {
	case health.CompositeHealth:
		composite = &v
	case *health.CompositeHealth:
		composite = v
	default:
		return h
	}

	ret := CompositeHealthV2{
		SimpleHealth: composite.SimpleHealth,
		Components:   make(map[string]health.Health),
	}
	// recursively convert components
	for k, v := range composite.Components {
		ret.Components[k] = ep.toSpringBootV2(v)
	}
	return ret
}
