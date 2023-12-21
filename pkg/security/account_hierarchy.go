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

package security

import (
	"context"
)

/*********************************
	Abstraction - DTO
 *********************************/

type Provider struct {
	Id               string
	Name             string
	DisplayName      string
	Description      string
	LocaleCode       string
	NotificationType string
	Email            string
	//CurrencyCode string
}

type Tenant struct {
	Id           string
	ExternalId   string
	DisplayName  string
	Description  string
	ProviderId   string
	Suspended    bool
	CurrencyCode string
	LocaleCode   string
}

/*********************************
	Abstraction - Stores
 *********************************/

type ProviderStore interface {
	LoadProviderById(ctx context.Context, id string) (*Provider, error)
}

type TenantStore interface {
	LoadTenantById(ctx context.Context, id string) (*Tenant, error)
	LoadTenantByExternalId(ctx context.Context, name string) (*Tenant, error)
}
