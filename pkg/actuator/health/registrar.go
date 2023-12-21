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
	"fmt"
	"go.uber.org/fx"
)

// SystemHealthRegistrar implements Registrar
type SystemHealthRegistrar struct {
	Indicator            *CompositeIndicator
	DetailsDisclosure    DetailsDisclosureControl
	ComponentsDisclosure ComponentsDisclosureControl
}

type regDI struct {
	fx.In
	Properties    HealthProperties
}

func NewSystemHealthRegistrar(di regDI) *SystemHealthRegistrar {
	return &SystemHealthRegistrar{
		Indicator: &CompositeIndicator{
			name: "system",
			delegates: []Indicator{
				PingIndicator{},
			},
			aggregator: NewSimpleStatusAggregator(func(opt *AggregateOption) {
				if len(di.Properties.Status.Orders) != 0 {
					opt.StatusOrders = di.Properties.Status.Orders
				}
			}),
		},
	}
}

// Register configure SystemHealthRegistrar
// supported input parameters are:
// 	- Indicator
// 	- StatusAggregator
// 	- DetailsDisclosureControl
// 	- ComponentsDisclosureControl
//  - DisclosureControl
func (i *SystemHealthRegistrar) Register(items ...interface{}) error {
	for _, v := range items {
		if e := i.register(v); e != nil {
			return e
		}
	}
	return nil
}

func (i *SystemHealthRegistrar) MustRegister(items ...interface{}) {
	if e := i.Register(items...); e != nil {
		panic(e)
	}
}

func (i *SystemHealthRegistrar) register(item interface{}) error {
	switch v := item.(type) {
	case []interface{}:
		return i.Register(v...)
	case Indicator:
		i.Indicator.Add(v)
	case StatusAggregator:
		i.Indicator.aggregator = v
	case DisclosureControl:
		i.DetailsDisclosure = v
		i.ComponentsDisclosure = v
	case DetailsDisclosureControl:
		i.DetailsDisclosure = v
	case ComponentsDisclosureControl:
		i.ComponentsDisclosure = v
	default:
		return fmt.Errorf("unsupported item %T", item)
	}
	return nil
}
