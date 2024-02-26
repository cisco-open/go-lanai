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
	"github.com/cisco-open/go-lanai/pkg/utils"
	"time"
)

type ContextDetailsStore interface {
	ReadContextDetails(ctx context.Context, key interface{}) (ContextDetails, error)
	SaveContextDetails(ctx context.Context, key interface{}, details ContextDetails) error
	RemoveContextDetails(ctx context.Context, key interface{}) error
	ContextDetailsExists(ctx context.Context, key interface{}) bool
}

type ContextDetails interface {
	AuthenticationDetails
	KeyValueDetails
}

// ProviderDetails is available if tenant is selected (tenant dictates provider)
type ProviderDetails interface {
	ProviderId() string
	ProviderName() string
	ProviderDisplayName() string
	ProviderDescription() string
	ProviderEmail() string
	ProviderNotificationType() string
}

// TenantDetails is available in the following scenarios:
//
//	user auth, tenant can be determined (either selected tenant, or there is a default tenant)
//	client auth, tenant can be determined (either selected tenant, or there is a default tenant)
type TenantDetails interface {
	TenantId() string
	TenantExternalId() string
	TenantSuspended() bool
}

// UserDetails is available for user authentication
type UserDetails interface {
	UserId() string
	Username() string
	AccountType() AccountType
	// AssignedTenantIds
	// Deprecated: usage of this method is not encouraged. Designs that require user to select tenancy is preferred
	// i.e. design tenancy based on TenantDetails instead.
	AssignedTenantIds() utils.StringSet
	LocaleCode() string
	CurrencyCode() string
	FirstName() string
	LastName() string
	Email() string
}

type AuthenticationDetails interface {
	ExpiryTime() time.Time
	IssueTime() time.Time
	Roles() utils.StringSet
	Permissions() utils.StringSet
	AuthenticationTime() time.Time
}

type ProxiedUserDetails interface {
	OriginalUsername() string
	Proxied() bool
}

type KeyValueDetails interface {
	Value(string) (interface{}, bool)
	Values() map[string]interface{}
}
