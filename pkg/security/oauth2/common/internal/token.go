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

package internal

import (
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"time"
)

// DecodedAccessToken implements oauth2.AccessToken and oauth2.ClaimsContainer
type DecodedAccessToken struct {
	DecodedClaims *ExtendedClaims
	TokenValue    string
	ExpireAt      time.Time
	IssuedAt      time.Time
	ScopesSet     utils.StringSet
}

func NewDecodedAccessToken() *DecodedAccessToken {
	return &DecodedAccessToken{}
}

func (t *DecodedAccessToken) Value() string {
	return t.TokenValue
}

func (t *DecodedAccessToken) ExpiryTime() time.Time {
	return t.ExpireAt
}

func (t *DecodedAccessToken) Expired() bool {
	return !t.ExpireAt.IsZero() && t.ExpireAt.Before(time.Now())
}

func (t *DecodedAccessToken) Details() map[string]interface{} {
	return map[string]interface{}{}
}

func (t *DecodedAccessToken) Type() oauth2.TokenType {
	return oauth2.TokenTypeBearer
}

func (t *DecodedAccessToken) IssueTime() time.Time {
	return t.IssuedAt
}

func (t *DecodedAccessToken) Scopes() utils.StringSet {
	return t.ScopesSet
}

func (t *DecodedAccessToken) RefreshToken() oauth2.RefreshToken {
	return nil
}

// oauth2.ClaimsContainer
func (t *DecodedAccessToken) Claims() oauth2.Claims {
	return t.DecodedClaims
}

// oauth2.ClaimsContainer
func (t *DecodedAccessToken) SetClaims(claims oauth2.Claims) {
	if c, ok := claims.(*ExtendedClaims); ok {
		t.DecodedClaims = c
		return
	}
	t.DecodedClaims = NewExtendedClaims(claims)
}


// DecodedRefreshToken implements oauth2.RefreshToken and oauth2.ClaimsContainer
type DecodedRefreshToken struct {
	DecodedClaims *ExtendedClaims
	TokenValue    string
	ExpireAt      time.Time
	IssuedAt      time.Time
	ScopesSet     utils.StringSet
}

func (t *DecodedRefreshToken) Value() string {
	return t.TokenValue
}

func (t *DecodedRefreshToken) ExpiryTime() time.Time {
	return t.ExpireAt
}

func (t *DecodedRefreshToken) Expired() bool {
	return !t.ExpireAt.IsZero() && t.ExpireAt.Before(time.Now())
}

func (t *DecodedRefreshToken) Details() map[string]interface{} {
	return map[string]interface{}{}
}

func (t *DecodedRefreshToken) WillExpire() bool {
	return !t.ExpireAt.IsZero()
}

// oauth2.ClaimsContainer
func (t *DecodedRefreshToken) Claims() oauth2.Claims {
	return t.DecodedClaims
}

// oauth2.ClaimsContainer
func (t *DecodedRefreshToken) SetClaims(claims oauth2.Claims) {
	if c, ok := claims.(*ExtendedClaims); ok {
		t.DecodedClaims = c
		return
	}
	t.DecodedClaims = NewExtendedClaims(claims)
}


