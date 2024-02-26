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

package sectest

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
)

const (
	MockedProviderID = "test-provider"
	MockedProviderName = "test-provider"
)

type MockedProviderStore struct {}

func (s MockedProviderStore) LoadProviderById(_ context.Context, id string) (*security.Provider, error) {
	if id != MockedProviderID {
		return nil, fmt.Errorf("cannot find provider with id [%s]", id)
	}
	return &security.Provider{
		Id:               id,
		Name:             MockedProviderName,
		DisplayName:      MockedProviderName,
		Description:      MockedProviderName,
		LocaleCode:       "en_US",
		NotificationType: "EMAIL",
		Email:            "admin@cisco.com",
	}, nil
}
