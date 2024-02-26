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
    "encoding/json"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "reflect"
    "time"
)

const (
	ClaimTag = "claim"
)

type Claims interface {
	Get(claim string) interface{}
	Has(claim string) bool
	Set(claim string, value interface{})
	Values() map[string]interface{}
}

// StringSetClaim is an alias of utils.StringSet with different JSON serialization specialized for some Claims
// StringSetClaim serialize as JSON string if there is single element in the set, otherwise as JSON array
type StringSetClaim utils.StringSet

// MarshalJSON json.Marshaler
func (s StringSetClaim) MarshalJSON() ([]byte, error) {
	switch len(s) {
	case 1:
		var v string
		for v = range s {
			// SuppressWarnings go:S108 empty block is intended to get any entry in the set
		}
		return json.Marshal(v)
	default:
		return utils.StringSet(s).MarshalJSON()
	}
}

// UnmarshalJSON json.Unmarshaler
func (s StringSetClaim) UnmarshalJSON(data []byte) error {
	values := make([]string, 0)
	if e := json.Unmarshal(data, &values); e == nil {
		utils.StringSet(s).Add(values...)
		return nil
	}

	// fallback to string
	value := ""
	if e := json.Unmarshal(data, &value); e != nil {
		return e
	}
	if value != "" {
		utils.StringSet(s).Add(value)
	}
	return nil
}

/*********************
	Implements
 *********************/

// MapClaims imlements Claims & claimsMapper
type MapClaims map[string]interface{}

func (c MapClaims) MarshalJSON() ([]byte, error) {
	m, e := c.toMap(true)
	if e != nil {
		return nil, e
	}
	return json.Marshal(m)
}

func (c MapClaims) UnmarshalJSON(bytes []byte) error {
	m := map[string]interface{}{}
	if e := json.Unmarshal(bytes, &m); e != nil {
		return e
	}
	return c.fromMap(m)
}

func (c MapClaims) Get(claim string) interface{} {
	return c[claim]
}

func (c MapClaims) Has(claim string) bool {
	_, ok := c[claim]
	return ok
}

func (c MapClaims) Set(claim string, value interface{}) {
	c[claim] = value
}

func (c MapClaims) Values() map[string]interface{} {
	ret, e := c.toMap(false)
	if e != nil {
		return map[string]interface{}{}
	}
	return ret
}

func (c MapClaims) toMap(convert bool) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for k, v := range c {
		if convert {
			value, e := claimMarshalConvert(reflect.ValueOf(v))
			if e != nil {
				return nil, e
			}
			ret[k] = value.Interface()
		} else {
			ret[k] = v
		}
	}
	return ret, nil
}

func (c MapClaims) fromMap(src map[string]interface{}) error {

	for k, v := range src {
		value, e := claimUnmarshalConvert(reflect.ValueOf(v), anyType)
		if e != nil {
			return e
		}
		c[k] = value.Interface()
	}
	return nil
}

// BasicClaims imlements Claims
type BasicClaims struct {
	FieldClaimsMapper
	Audience  StringSetClaim  `claim:"aud"`
	ExpiresAt time.Time       `claim:"exp"`
	Id        string          `claim:"jti"`
	IssuedAt  time.Time       `claim:"iat"`
	Issuer    string          `claim:"iss"`
	NotBefore time.Time       `claim:"nbf"`
	Subject   string          `claim:"sub"`
	Scopes    utils.StringSet `claim:"scope"`
	ClientId  string          `claim:"client_id"`
}

func (c *BasicClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *BasicClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *BasicClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *BasicClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *BasicClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *BasicClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}
