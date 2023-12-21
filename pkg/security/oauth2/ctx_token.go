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

package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
	"time"
)

/*****************************
	Abstractions
 *****************************/

type TokenType string
const(
	TokenTypeBearer = "bearer"
	TokenTypeMac    = "mac"
	TokenTypeBasic  = "basic"
)

func (t TokenType) HttpHeader() string {
	switch strings.ToLower(string(t)) {
	case TokenTypeMac:
		return "MAC"
	case TokenTypeBasic:
		return "Basic"
	default:
		return "Bearer"
	}
}

type Token interface {
	Value() string
	ExpiryTime() time.Time
	Expired() bool
	Details() map[string]interface{}
}

type ClaimsContainer interface {
	Claims() Claims
	SetClaims(claims Claims)
}

type AccessToken interface {
	Token
	Type() TokenType
	IssueTime() time.Time
	Scopes() utils.StringSet
	RefreshToken() RefreshToken
}

type RefreshToken interface {
	Token
	WillExpire() bool
}

/*******************************
	Common Impl. AccessToken
 *******************************/

// DefaultAccessToken implements AccessToken and ClaimsContainer
type DefaultAccessToken struct {
	claims       Claims
	tokenType    TokenType
	value        string
	expiryTime   time.Time
	issueTime	 time.Time
	scopes       utils.StringSet
	refreshToken *DefaultRefreshToken
	details      map[string]interface{}
}

func NewDefaultAccessToken(value string) *DefaultAccessToken {
	return &DefaultAccessToken{
		value:     value,
		tokenType: TokenTypeBearer,
		scopes:    utils.NewStringSet(),
		issueTime: time.Now(),
		details:   map[string]interface{}{},
		claims:    MapClaims{},
	}
}

func FromAccessToken(token AccessToken) *DefaultAccessToken {
	if t, ok := token.(*DefaultAccessToken); ok {
		return &DefaultAccessToken{
			value:        t.value,
			tokenType:    t.tokenType,
			expiryTime:   t.expiryTime,
			issueTime:    t.issueTime,
			scopes:       t.scopes.Copy(),
			claims:       t.claims,
			details:      copyMap(t.details),
			refreshToken: t.refreshToken,
		}
	}

	cp := &DefaultAccessToken{
		value:      token.Value(),
		tokenType:  token.Type(),
		expiryTime: token.ExpiryTime(),
		issueTime:  token.IssueTime(),
		scopes:     token.Scopes().Copy(),
		details:    copyMap(token.Details()),
	}
	cp.SetRefreshToken(token.RefreshToken())
	return cp
}

// Value implements AccessToken
func (t *DefaultAccessToken) Value() string {
	return t.value
}

// Details implements AccessToken
func (t *DefaultAccessToken) Details() map[string]interface{} {
	return t.details
}

// Type implements AccessToken
func (t *DefaultAccessToken) Type() TokenType {
	return t.tokenType
}

// IssueTime implements AccessToken
func (t *DefaultAccessToken) IssueTime() time.Time {
	return t.issueTime
}

// ExpiryTime implements AccessToken
func (t *DefaultAccessToken) ExpiryTime() time.Time {
	return t.expiryTime
}

// Expired implements AccessToken
func (t *DefaultAccessToken) Expired() bool {
	return !t.expiryTime.IsZero() && t.expiryTime.Before(time.Now())
}

// Scopes implements AccessToken
func (t *DefaultAccessToken) Scopes() utils.StringSet {
	return t.scopes
}

// RefreshToken implements AccessToken
func (t *DefaultAccessToken) RefreshToken() RefreshToken {
	if t.refreshToken == nil {
		return nil
	}
	return t.refreshToken
}

// Claims implements ClaimsContainer
func (t *DefaultAccessToken) Claims() Claims {
	return t.claims
}

// SetClaims implements ClaimsContainer
func (t *DefaultAccessToken) SetClaims(claims Claims) {
	t.claims = claims
}

/* Setters */

func (t *DefaultAccessToken) SetValue(v string) *DefaultAccessToken {
	t.value = v
	return t
}

func (t *DefaultAccessToken) SetIssueTime(v time.Time) *DefaultAccessToken {
	t.issueTime = v.UTC()
	return t
}

func (t *DefaultAccessToken) SetExpireTime(v time.Time) *DefaultAccessToken {
	t.expiryTime = v.UTC()
	return t
}

func (t *DefaultAccessToken) SetRefreshToken(v RefreshToken) *DefaultAccessToken {
	if refresh, ok := v.(*DefaultRefreshToken); ok {
		t.refreshToken = refresh
	} else {
		t.refreshToken = FromRefreshToken(v)
	}
	return t
}

func (t *DefaultAccessToken) SetScopes(scopes utils.StringSet) *DefaultAccessToken {
	t.scopes = scopes.Copy()
	return t
}

func (t *DefaultAccessToken) AddScopes(scopes...string) *DefaultAccessToken {
	t.scopes.Add(scopes...)
	return t
}

func (t *DefaultAccessToken) RemoveScopes(scopes...string) *DefaultAccessToken {
	t.scopes.Remove(scopes...)
	return t
}

func (t *DefaultAccessToken) PutDetails(key string, value interface{}) *DefaultAccessToken {
	if value == nil {
		delete(t.details, key)
	} else {
		t.details[key] = value
	}
	return t
}

/********************************
	Common Impl. RefreshToken
 ********************************/

// DefaultRefreshToken implements RefreshToken and ClaimsContainer
type DefaultRefreshToken struct {
	claims       Claims
	value        string
	expiryTime   time.Time
	details      map[string]interface{}
}

func NewDefaultRefreshToken(value string) *DefaultRefreshToken {
	return &DefaultRefreshToken{
		value:   value,
		details: map[string]interface{}{},
		claims:  MapClaims{},
	}
}

func FromRefreshToken(token RefreshToken) *DefaultRefreshToken {
	if t, ok := token.(*DefaultRefreshToken); ok {
		return &DefaultRefreshToken{
			value: t.value,
			details: copyMap(t.details),
			claims: t.claims,
		}
	}

	return &DefaultRefreshToken{
		value: token.Value(),
		details: copyMap(token.Details()),
	}
}

// Value implements RefreshToken
func (t *DefaultRefreshToken) Value() string {
	return t.value
}

// Details implements RefreshToken
func (t *DefaultRefreshToken) Details() map[string]interface{} {
	return t.details
}

// ExpiryTime implements RefreshToken
func (t *DefaultRefreshToken) ExpiryTime() time.Time {
	return t.expiryTime
}

// Expired implements RefreshToken
func (t *DefaultRefreshToken) Expired() bool {
	return !t.expiryTime.IsZero() && t.expiryTime.Before(time.Now())
}

// WillExpire implements RefreshToken
func (t *DefaultRefreshToken) WillExpire() bool {
	return !t.expiryTime.IsZero()
}

// Claims implements ClaimsContainer
func (t *DefaultRefreshToken) Claims() Claims {
	return t.claims
}

// SetClaims implements ClaimsContainer
func (t *DefaultRefreshToken) SetClaims(claims Claims) {
	t.claims = claims
}

/* Setters */

func (t *DefaultRefreshToken) SetValue(v string) *DefaultRefreshToken {
	t.value = v
	return t
}

func (t *DefaultRefreshToken) SetExpireTime(v time.Time) *DefaultRefreshToken {
	t.expiryTime = v.UTC()
	return t
}

func (t *DefaultRefreshToken) PutDetails(key string, value interface{}) *DefaultRefreshToken {
	if value == nil {
		delete(t.details, key)
	} else {
		t.details[key] = value
	}
	return t
}

/********************************
	Helpers
 ********************************/
func copyMap(src map[string]interface{}) map[string]interface{} {
	dest := map[string]interface{}{}
	for k,v := range src {
		dest[k] = v
	}
	return dest
}


