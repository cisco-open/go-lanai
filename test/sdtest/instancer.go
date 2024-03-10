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

package sdtest

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/discovery"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "time"
)

type InstanceMockOptions func(inst *discovery.Instance)
type InstanceMockMatcher func(inst *discovery.Instance) bool

type InstancerMock struct {
	ctx           context.Context
	SName         string
	InstanceMocks []*discovery.Instance
	ErrTimeMock   time.Time
	ErrMock       error
	Started       bool
}

func NewMockInstancer(ctx context.Context, svcName string) *InstancerMock {
	return &InstancerMock{
		ctx:           ctx,
		SName:         svcName,
		InstanceMocks: make([]*discovery.Instance, 0, 4),
	}
}

/* discovery.Instancer impelementation */

func (i *InstancerMock) ServiceName() string {
	return i.SName
}

func (i *InstancerMock) Service() *discovery.Service {
	return &discovery.Service{
		Name:       i.SName,
		Insts:      i.InstanceMocks,
		Time:       time.Now(),
		Err:        i.ErrMock,
		FirstErrAt: i.ErrTimeMock,
	}
}

func (i *InstancerMock) Instances(matcher discovery.InstanceMatcher) ([]*discovery.Instance, error) {
	if i.ErrMock != nil {
		return nil, i.ErrMock
	}

	if matcher == nil {
		matcher = discovery.InstanceIsHealthy()
	}

	ret := make([]*discovery.Instance, 0, len(i.InstanceMocks))
	for _, inst := range i.InstanceMocks {
		if ok, e := matcher.MatchesWithContext(i.ctx, inst); e == nil && ok {
			ret = append(ret, inst)
		}
	}
	return ret, nil
}

func (i *InstancerMock) Start(_ context.Context) {
	i.Started = true
}

func (i *InstancerMock) Stop() {
	i.Started = false
}

func (i *InstancerMock) RegisterCallback(_ interface{}, _ discovery.Callback) {
	// noop
}

func (i *InstancerMock) DeregisterCallback(_ interface{}) {
	// noop
}

/* Addtional mock methods */

func (i *InstancerMock) MockInstances(count int, opts ...InstanceMockOptions) []*discovery.Instance {
	defer i.resetError()
	i.InstanceMocks = make([]*discovery.Instance, count)
	for j := 0; j < count; j++ {
		var inst = discovery.Instance{
			ID:       fmt.Sprintf("%d-%s", j, utils.RandomString(10)),
			Service:  i.SName,
			Address:  "127.0.0.1",
			Port:     utils.RandomIntN(32767) + 32768,
			Tags:     []string{},
			Meta:     map[string]string{},
			Health:   discovery.HealthPassing,
		}
		for _, fn := range opts {
			fn(&inst)
		}
		i.InstanceMocks[j] = &inst
	}
	return i.InstanceMocks
}

func (i *InstancerMock) UpdateInstances(matcher InstanceMockMatcher, opts ...InstanceMockOptions) (count int) {
	defer i.resetError()
	for _, inst := range i.InstanceMocks {
		if ok := matcher(inst); !ok {
			continue
		}
		for _, fn := range opts {
			fn(inst)
		}
		count ++
	}
	return
}

func (i *InstancerMock) MockError(what error, when time.Time) {
	i.ErrMock = what
	i.ErrTimeMock = when
}

func (i *InstancerMock) resetError() {
	i.ErrMock = nil
	i.ErrTimeMock = time.Time{}
}
