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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
)

const MockingPrefix = "mocking"

type MockingProperties struct {
	Accounts    map[string]*sectest.MockedAccountProperties       `json:"accounts"`
	FedAccounts map[string]*sectest.MockedFederatedUserProperties `json:"fed-accounts"`
	Tenants     map[string]*sectest.MockedTenantProperties        `json:"tenants"`
	Clients     map[string]*sectest.MockedClientProperties        `json:"clients"`
}

func BindMockingProperties(appCtx *bootstrap.ApplicationContext) MockingProperties {
	props := MockingProperties{
		Accounts: map[string]*sectest.MockedAccountProperties{},
		Tenants:  map[string]*sectest.MockedTenantProperties{},
	}
	if e := appCtx.Config().Bind(&props, MockingPrefix); e != nil {
		panic(e)
	}
	return props
}

func MapValues[T any](m map[string]T) []T {
	ret := make([]T, 0, len(m))
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}
