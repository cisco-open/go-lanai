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

package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const (
	tokenDelimiter = "~"
)

/*************************
	Token
 *************************/

type MockedTokenInfo struct {
	UName       string `json:"UName"`
	UID         string `json:"UID"`
	TID         string `json:"TID"`
	TExternalId string `json:"TExternalId"`
	OrigU       string `json:"OrigU"`
	Exp         int64  `json:"Exp"`
	Iss         int64  `json:"Iss"`
}

// MockedToken implements oauth2.AccessToken
type MockedToken struct {
	MockedTokenInfo
	Token        string
	ExpTime      time.Time `json:"-"`
	IssTime      time.Time `json:"-"`
	MockedScopes []string  `json:"-"`
}

func (mt MockedToken) MarshalText() (text []byte, err error) {
	if len(mt.Token) != 0 {
		return []byte(mt.Token), nil
	}
	mt.Exp = mt.ExpTime.UnixNano()
	mt.Iss = mt.IssTime.UnixNano()
	text, err = json.Marshal(mt.MockedTokenInfo)
	if err != nil {
		return
	}
	return []byte(base64.StdEncoding.EncodeToString(text)), nil
}

func (mt *MockedToken) UnmarshalText(text []byte) error {
	data, e := base64.StdEncoding.DecodeString(string(text))
	if e != nil {
		return e
	}
	if e := json.Unmarshal(data, &mt.MockedTokenInfo); e != nil {
		return e
	}
	mt.ExpTime = time.Unix(0, mt.Exp)
	mt.IssTime = time.Unix(0, mt.Iss)
	return nil
}

func (mt MockedToken) String() string {
	vals := []string{mt.UName, mt.UID, mt.TID, mt.TExternalId, mt.OrigU, mt.ExpTime.Format(utils.ISO8601Milliseconds)}
	return strings.Join(vals, tokenDelimiter)
}

func (mt *MockedToken) Value() string {
	text, e := mt.MarshalText()
	if e != nil {
		return ""
	}
	return string(text)
}

func (mt *MockedToken) ExpiryTime() time.Time {
	return mt.ExpTime
}

func (mt *MockedToken) Expired() bool {
	return !mt.ExpTime.IsZero() && !time.Now().Before(mt.ExpTime)
}

func (mt *MockedToken) Details() map[string]interface{} {
	return map[string]interface{}{}
}

func (mt *MockedToken) Type() oauth2.TokenType {
	return oauth2.TokenTypeBearer
}

func (mt *MockedToken) IssueTime() time.Time {
	return mt.IssTime
}

func (mt *MockedToken) Scopes() utils.StringSet {
	return utils.NewStringSet(mt.MockedScopes...)
}

func (mt *MockedToken) RefreshToken() oauth2.RefreshToken {
	return nil
}
