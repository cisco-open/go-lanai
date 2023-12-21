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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"encoding/json"
)

type ClaimsRequest struct {
	UserInfo requestedClaims `json:"userinfo"`
	IdToken  requestedClaims `json:"id_token"`
}

// requestedClaims implements claims.RequestedClaims
type requestedClaims map[string]requestedClaim

func (r requestedClaims) Get(claim string) (c claims.RequestedClaim, ok bool) {
	c, ok = r[claim]
	return
}

type rcDetails struct {
	Essential   bool     `json:"essential"`
	Values      []string `json:"values,omitempty"`
	SingleValue *string  `json:"value,omitempty"`
}

// requestedClaim implements claims.RequestedClaim and json.Unmarshaler
type requestedClaim struct {
	rcDetails
}

func (r requestedClaim) Essential() bool {
	return r.rcDetails.Essential
}

func (r requestedClaim) Values() []string {
	return r.rcDetails.Values
}

func (r requestedClaim) IsDefault() bool {
	return len(r.rcDetails.Values) == 0
}

func (r *requestedClaim) UnmarshalJSON(data []byte) error {
	r.rcDetails.Values = []string{}
	if e := json.Unmarshal(data, &r.rcDetails); e != nil {
		return e
	}

	if r.rcDetails.SingleValue != nil {
		r.rcDetails.Values = []string{*r.rcDetails.SingleValue}
		r.rcDetails.SingleValue = nil
	}
	return nil
}
