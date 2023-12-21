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

package openid

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

// IdTokenClaims implements oauth2.Claims
type IdTokenClaims struct {
	oauth2.FieldClaimsMapper

	/*******************************
	 * Standard Jwt claims
	 *******************************/

	Issuer   string                `claim:"iss"`
	Subject  string                `claim:"sub"`
	Audience oauth2.StringSetClaim `claim:"aud"`
	Expire   time.Time             `claim:"exp"`
	IssueAt  time.Time             `claim:"iat"`

	/*******************************
	 * Standard ID Token claims
	 *******************************/

	/* Standard */
	AuthTime        time.Time `claim:"auth_time"`
	Nonce           string    `claim:"nonce"`
	AuthCtxClassRef string    `claim:"acr"`
	AuthMethodRef   []string  `claim:"amr"`
	AuthorizedParty string    `claim:"azp"`
	AccessTokenHash string    `claim:"at_hash"`

	/* Profile Scope */
	FullName          string    `claim:"name"`
	FirstName         string    `claim:"given_name"`
	LastName          string    `claim:"family_name"`
	MiddleName        string    `claim:"middle_name"`
	Nickname          string    `claim:"nickname"`
	PreferredUsername string    `claim:"preferred_username"`
	ProfileUrl        string    `claim:"profile"`
	PictureUrl        string    `claim:"picture"`
	Website           string    `claim:"website"`
	Gender            string    `claim:"gender"`
	Birthday          string    `claim:"birthdate"` // ISO 8601:2004 [ISO8601‑2004] YYYY-MM-DD format
	ZoneInfo          string    `claim:"zoneinfo"`  // Europe/Paris or America/Los_Angeles
	Locale            string    `claim:"locale"`    // Typically ISO 639-1 Alpha-2 [ISO639‑1] language code in lowercase and an ISO 3166-1
	UpdatedAt         time.Time `claim:"updated_at"`

	/* Email Scope */
	Email         string `claim:"email"`
	EmailVerified *bool  `claim:"email_verified"`

	/* Phone Number Scope */
	PhoneNumber      string `claim:"phone_number"` // RFC 3966 [RFC3966] e.g. +1 (604) 555-1234;ext=5678
	PhoneNumVerified *bool  `claim:"phone_number_verified"`

	/* Address Scope */
	Address *claims.AddressClaim `claim:"address"`

	/*******************************
	 * NFV Additional Claims
	 *******************************/

	/* Profile Scope */
	DefaultTenantId string          `claim:"default_tenant_id"`
	AssignedTenants utils.StringSet `claim:"assigned_tenants"`
	Roles           utils.StringSet `claim:"roles"`
	Permissions     utils.StringSet `claim:"permissions"`

	/* General Scope */
	UserId          string `claim:"user_id"`
	AccountType     string `claim:"account_type"`
	TenantId        string `claim:"tenant_id"`
	TenantExternalId      string `claim:"tenant_name"` //for backward compatibility, map to tenant_name
	TenantSuspended *bool  `claim:"tenant_suspended"`
	ProviderId      string `claim:"provider_id"`
	ProviderName    string `claim:"provider_name"`
	OrigUsername    string `claim:"original_username"`
	Currency        string `claim:"currency"`
}

func (c *IdTokenClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *IdTokenClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *IdTokenClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *IdTokenClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *IdTokenClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *IdTokenClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}
