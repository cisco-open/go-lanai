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

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"

// ExtendedClaims imlements oauth2.Claims. It's used only for access token decoding
type ExtendedClaims struct {
	oauth2.FieldClaimsMapper
	oauth2.BasicClaims
	oauth2.Claims
}

func NewExtendedClaims(claims ...oauth2.Claims) *ExtendedClaims {
	ptr := &ExtendedClaims{
		Claims: oauth2.MapClaims{},
	}
	for _, c := range claims {
		values := c.Values()
		for k, v := range values {
			ptr.Set(k, v)
		}
	}

	return ptr
}

func (c *ExtendedClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *ExtendedClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *ExtendedClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *ExtendedClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *ExtendedClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *ExtendedClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}

