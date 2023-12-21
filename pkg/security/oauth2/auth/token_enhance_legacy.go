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

package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
)

/*****************************
	legacyClaims Enhancer
 *****************************/

// legacyClaims implements Claims and includes BasicClaims
type legacyClaims struct {
	oauth2.FieldClaimsMapper
	*oauth2.BasicClaims
	FirstName string   `claim:"firstName"`
	LastName  string   `claim:"lastName"`
	Email     string   `claim:"email"`
	TenantId  string   `claim:"tenantId"`
	Username  string   `claim:"user_name"`
	Roles     []string `claim:"roles"`
}

func (c *legacyClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *legacyClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *legacyClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *legacyClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *legacyClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *legacyClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}

// LegacyTokenEnhancer implements order.Ordered and TokenEnhancer
// LegacyTokenEnhancer add legacy claims and response fields that was supported by Java version of IDM
// but deprecated in Go version
type LegacyTokenEnhancer struct {}

func (te *LegacyTokenEnhancer) Order() int {
	return TokenEnhancerOrderDetailsClaims
}

func (te *LegacyTokenEnhancer) Enhance(_ context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	if t.Claims() == nil {
		return nil, oauth2.NewInternalError("LegacyTokenEnhancer need to be placed immediately after BasicClaimsEnhancer")
	}

	basic, ok := t.Claims().(*oauth2.BasicClaims)
	if !ok {
		return nil, oauth2.NewInternalError("LegacyTokenEnhancer need to be placed immediately after BasicClaimsEnhancer")
	}

	legacy := &legacyClaims{
		BasicClaims: basic,
		Username:    basic.Subject,
	}
	t.PutDetails(oauth2.ClaimUsername, legacy.Username)

	if ud, ok := oauth.Details().(security.UserDetails); ok {
		legacy.FirstName = ud.FirstName()
		legacy.LastName = ud.LastName()
		legacy.Email = ud.Email()
	}

	if td, ok := oauth.Details().(security.TenantDetails); ok {
		legacy.TenantId = td.TenantId()
		t.PutDetails(oauth2.ClaimLegacyTenantId, td.TenantId())
	}

	if ad, ok := oauth.Details().(security.AuthenticationDetails); ok {
		legacy.Roles = ad.Roles().Values()
		t.PutDetails(oauth2.ClaimRoles, legacy.Roles)
	}

	t.SetClaims(legacy)
	return t, nil
}

// ResourceIdTokenEnhancer impelments order.Ordered and TokenEnhancer
// spring-security-oauth2 based java implementation expecting "aud" claims to be the resource ID
type ResourceIdTokenEnhancer struct {
}

func (te *ResourceIdTokenEnhancer) Order() int {
	return TokenEnhancerOrderResourceIdClaims
}

func (te *ResourceIdTokenEnhancer) Enhance(c context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	if t.Claims() == nil || !t.Claims().Has(oauth2.ClaimAudience) {
		return nil, oauth2.NewInternalError("ResourceIdTokenEnhancer need to be placed after BasicClaimsEnhancer")
	}

	aud := claims.LegacyAudience(c, &claims.FactoryOption{
		Source: oauth,
	})
	t.Claims().Set(oauth2.ClaimAudience, aud)
	return t, nil
}
