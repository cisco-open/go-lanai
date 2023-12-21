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
	"math"
)

/*******************************
	SimpleStatusAggregator
********************************/

var (
	DefaultStatusOrders = []Status{
		StatusDown, StatusOutOfService, StatusUp, StatusDown,
	}
)

type AggregateOptions func(opt *AggregateOption)
type AggregateOption struct {
	StatusOrders []Status
}

// SimpleStatusAggregator implements StatusAggregator
type SimpleStatusAggregator struct {
	orders map[Status]int
}

func NewSimpleStatusAggregator(opts ...AggregateOptions) *SimpleStatusAggregator {
	opt := AggregateOption{
		StatusOrders: DefaultStatusOrders,
	}
	for _, f := range opts {
		f(&opt)
	}
	orders := map[Status]int{}
	for i, s := range opt.StatusOrders {
		orders[s] = i
	}
	return &SimpleStatusAggregator{
		orders: orders,
	}
}

func (a SimpleStatusAggregator) Aggregate(_ context.Context, statuses ...Status) Status {
	var status Status
	unknown := true
	for _, s := range statuses {
		if unknown || a.compare(s, status) < 0 {
			unknown = false
			status = s
		}
	}

	if unknown {
		return StatusUnknown
	}
	return status
}

func (a SimpleStatusAggregator) compare(s1, s2 Status) int {
	o1, ok := a.orders[s1]
	if !ok {
		o1 = math.MaxInt64
	}

	o2, ok := a.orders[s2]
	if !ok {
		o2 = math.MaxInt64
	}

	switch {
	case o1 < o2:
		return -1
	case o1 > o2:
		return 1
	default:
		return 0
	}
}
