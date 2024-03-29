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

package misc

import (
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"time"
)

// CheckTokenClaims implemnts oauth2.Claims
type CheckTokenClaims struct {
	oauth2.FieldClaimsMapper

	/*******************************
	 * Standard Check Token claims
	 *******************************/
	oauth2.BasicClaims
	Active   *bool   `claim:"active"`
	Username string `claim:"username"`

	/*******************************
	* Standard OIDC claims
	*******************************/
	FirstName string    `claim:"given_name"`
	LastName  string    `claim:"family_name"`
	Email     string    `claim:"email"`
	Locale    string    `claim:"locale"` // Typically ISO 639-1 Alpha-2 [ISO639‑1] language code in lowercase and an ISO 3166-1
	AuthTime  time.Time `claim:"auth_time"`

	/*******************************
	 * NFV Additional Claims
	 *******************************/
	UserId          string          `claim:"user_id"`
	AccountType     string          `claim:"account_type"`
	Currency        string          `claim:"currency"`
	TenantId        string          `claim:"tenant_id"`
	TenantExternalId string          `claim:"tenant_name"` //This maps to Tenant's ExternalId for backward compatibility
	TenantSuspended *bool           `claim:"tenant_suspended"`
	ProviderId      string          `claim:"provider_id"`
	ProviderName    string          `claim:"provider_name"`
	ProviderDisplayName string 		`claim:"provider_display_name"`
	ProviderDescription string 		`claim:"provider_description"`
	ProviderNotificationType string `claim:"provider_notification_type"`
	ProviderEmail string 			`claim:"provider_email"`
	AssignedTenants utils.StringSet `claim:"assigned_tenants"`
	Roles           utils.StringSet `claim:"roles"`
	Permissions     utils.StringSet `claim:"permissions"`
	OrigUsername    string          `claim:"original_username"`
}

func (c *CheckTokenClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *CheckTokenClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *CheckTokenClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *CheckTokenClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *CheckTokenClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *CheckTokenClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}
