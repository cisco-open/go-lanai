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

import "context"

/*******************************
	CompositeIndicator
********************************/

type IndicatorOptions func(opt *IndicatorOption)
type IndicatorOption struct {
	Name         string
	Contributors []Indicator
	Aggregator   StatusAggregator
}

// CompositeIndicator implement Indicator and SystemHealthRegistrar
type CompositeIndicator struct {
	name       string
	delegates  []Indicator
	aggregator StatusAggregator
}

func NewCompositeIndicator(opts ...IndicatorOptions) *CompositeIndicator {
	opt := IndicatorOption{
		Contributors: []Indicator{},
		Aggregator:   NewSimpleStatusAggregator(),
	}
	for _, f := range opts {
		f(&opt)
	}
	return &CompositeIndicator{
		name:       opt.Name,
		delegates:  opt.Contributors,
		aggregator: opt.Aggregator,
	}
}

func (c *CompositeIndicator) Add(contributors ...Indicator) {
	c.delegates = append(c.delegates, contributors...)
}

func (c *CompositeIndicator) Name() string {
	return c.name
}

func (c *CompositeIndicator) Health(ctx context.Context, options Options) Health {
	components := map[string]Health{}
	statuses := []Status{}
	for _, d := range c.delegates {
		h := d.Health(ctx, options)
		// although delegates should respect options, we don't want to leave any changes
		h = trySanitize(h, options, false)
		if options.ShowComponents {
			components[d.Name()] = h
		}
		statuses = append(statuses, h.Status())
	}
	status := c.aggregator.Aggregate(ctx, statuses...)
	return NewCompositeHealth(status, "", components)
}

/*******************************
	helpers
********************************/
func trySanitize(health Health, opts Options, deep bool) Health {
	// sanitize components
	if !opts.ShowComponents {
		health = sanitizeComponents(health)
	}

	// sanitize details
	if !opts.ShowDetails {
		health = sanitizeDetails(health, deep)
	}
	return health
}

func sanitizeComponents(health Health) Health {
	// sanitize components
	switch health.(type) {
	case *CompositeHealth:
		health.(*CompositeHealth).Components = map[string]Health{}
	case CompositeHealth:
		return NewCompositeHealth(health.Status(), health.Description(), map[string]Health{})
	}
	return health
}

// recursively clean up details if deep == true
func sanitizeDetails(health Health, deep bool) Health {
	// sanitize details
	switch health.(type) {
	case *DetailedHealth:
		health.(*DetailedHealth).Details = nil
	case DetailedHealth:
		health = NewDetailedHealth(health.Status(), health.Description(), nil)
	}

	if !deep {
		return health
	}
	switch health.(type) {
	case *CompositeHealth:
		for k, v := range health.(*CompositeHealth).Components {
			health.(*CompositeHealth).Components[k] = sanitizeDetails(v, deep)
		}
	case CompositeHealth:
		comps := map[string]Health{}
		for k, v := range health.(CompositeHealth).Components {
			comps[k] = sanitizeDetails(v, deep)
		}
		return NewCompositeHealth(health.Status(), health.Description(), comps)
	}
	return health
}
