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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"strconv"
	"strings"
)

func BeHealthy() InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Health = discovery.HealthPassing
	}
}

func BeCritical() InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Health = discovery.HealthCritical
	}
}

func WithExtraTag(tags ...string) InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Tags = append(inst.Tags, tags...)
	}
}

func WithMeta(k, v string) InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Meta[k] = v
	}
}

func AnyInstance() InstanceMockMatcher {
	return func(inst *discovery.Instance) bool {
		return true
	}
}

func NthInstance(n int) InstanceMockMatcher {
	return func(inst *discovery.Instance) bool {
		i := extractIndexIfPossible(inst)
		return i == n
	}
}

func InstanceAfterN(n int) InstanceMockMatcher {
	return func(inst *discovery.Instance) bool {
		i := extractIndexIfPossible(inst)
		return i > n
	}
}

func extractIndexIfPossible(inst *discovery.Instance) int {
	split := strings.SplitN(inst.ID, "-", 2)
	i, e := strconv.Atoi(split[0])
	if e != nil {
		return -1
	}
	return i
}