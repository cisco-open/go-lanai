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

package consulsd

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/discovery/sd"
	"github.com/cisco-open/go-lanai/pkg/utils/loop"
	"github.com/hashicorp/consul/api"
	"sort"
	"time"
)

const (
	defaultIndex       uint64  = 0
)

type InstancerOptions func(opt *InstancerOption)
type InstancerOption struct {
	sd.InstancerOption
	Selector         discovery.InstanceMatcher
	ConsulConnection *consul.Connection
}

// Instancer implements discovery.Instancer
// It yields service for a serviceName in Consul.
// See discovery.Instancer
type Instancer struct {
	sd.CachedInstancer
	consul   *consul.Connection
	lastMeta *api.QueryMeta
	selector discovery.InstanceMatcher
}

// NewInstancer returns a discovery.Instancer with Consul service discovery APIs.
// See discovery.Instancer
func NewInstancer(ctx context.Context, opts ...InstancerOptions) *Instancer {
	opt := InstancerOption{
		InstancerOption: sd.InstancerOption{
			Logger: logger,
			RefresherOptions: []loop.TaskOptions{
				loop.ExponentialRepeatIntervalOnError(50*time.Millisecond, sd.DefaultRefreshBackoffFactor),
			},
		},
	}
	for _, f := range opts {
		f(&opt)
	}
	i := &Instancer{
		CachedInstancer: sd.MakeCachedInstancer(func(baseOpt *sd.CachedInstancerOption) {
			baseOpt.InstancerOption = opt.InstancerOption
		}),
		consul:   opt.ConsulConnection,
		selector: opt.Selector,
	}
	i.BackgroundRefreshFunc = i.resolveInstancesTask()
	i.Start(ctx)
	return i
}

func (i *Instancer) resolveInstancesTask() func(ctx context.Context) (*discovery.Service, error) {
	// Note:
	// 		Consul doesn't support more than one tag in its serviceName query method.
	// 		https://github.com/hashicorp/consul/issues/294
	// 		Hashi suggest prepared queries, but they don't support blocking.
	// 		https://www.consul.io/docs/agent/http/query.html#execute
	// 		If we want blocking for efficiency, we can use single tag
	return func(ctx context.Context) (*discovery.Service, error) {
		// Note: i.lastMeta is only updated in this function, and this function is executed via loop.Loop.
		// 		 because loop.Loop guarantees that all tasks are executed one-by-one,
		// 		 there is no need to use Lock or locking
		lastIndex := defaultIndex
		if i.lastMeta != nil {
			lastIndex = i.lastMeta.LastIndex
		}
		opts := &api.QueryOptions{
			WaitIndex: lastIndex,
		}
		//entries, meta, e := i.client.Service(i.serviceName, "", false, opts.WithContext(ctx))
		entries, meta, e := i.consul.Client().Health().Service(i.ServiceName(), "", false, opts.WithContext(ctx))

		i.lastMeta = meta
		insts := makeInstances(entries, i.selector)
		service := &discovery.Service{
			Name:  i.ServiceName(),
			Insts: insts,
			Time:  time.Now(),
			Err:   e,
		}
		return service, e
	}
}

/***********************
	Helpers
***********************/

func makeInstances(entries []*api.ServiceEntry, selector discovery.InstanceMatcher) []*discovery.Instance {
	instances := make([]*discovery.Instance, 0)
	for _, entry := range entries {
		addr := entry.Service.Address
		if addr == "" {
			addr = entry.Node.Address
		}
		inst := &discovery.Instance{
			ID:       entry.Service.ID,
			Service:  entry.Service.Service,
			Address:  addr,
			Port:     entry.Service.Port,
			Tags:     entry.Service.Tags,
			Meta:     entry.Service.Meta,
			Health:   parseHealth(entry),
			RawEntry: entry,
		}

		if selector == nil {
			instances = append(instances, inst)
		} else if matched, e := selector.Matches(inst); e == nil && matched {
			instances = append(instances, inst)
		}
	}
	sort.SliceStable(instances, func(i, j int) bool {
		return instances[i].ID < instances[j].ID
	})
	return instances
}

func parseHealth(entry *api.ServiceEntry) discovery.HealthStatus {
	switch status := entry.Checks.AggregatedStatus(); status {
	case api.HealthPassing:
		return discovery.HealthPassing
	case api.HealthWarning:
		return discovery.HealthWarning
	case api.HealthCritical:
		return discovery.HealthCritical
	case api.HealthMaint:
		return discovery.HealthMaintenance
	default:
		return discovery.HealthAny
	}
}

