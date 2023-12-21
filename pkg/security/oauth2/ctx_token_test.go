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
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestAccessTokenJSONSerialization(t *testing.T) {
	refresh := NewDefaultRefreshToken("refresh token value").PutDetails("d1", "v1")
	token := NewDefaultAccessToken("token value")
	token.
		SetExpireTime(time.Now().Add(2*time.Hour)).
		PutDetails("d1", "v1").
		PutDetails("d2", "v1").
		AddScopes("s1", "s2").
		SetRefreshToken(refresh)
	token.Claims().Set("c1", "v1")
	token.Claims().Set("c2", "v2")

	bytes, err := json.Marshal(token)
	str := string(bytes)
	fmt.Printf("JSON: %s\n", str)

	switch {
	case err != nil:
		t.Errorf("Marshalling should not return error. But got %v \n", err)

	case len(str) == 0:
		t.Errorf("json should not be empty")

		//TODO more cases
	}

	// Deserialize
	parsed := NewDefaultAccessToken("")
	err = json.Unmarshal([]byte(str), &parsed)

	switch {
	case err != nil:
		t.Errorf("Unmarshalling should not return error. But got %v \n", err)

	case parsed.Value() != "token value":
		t.Errorf("parsed value should be [%s], but is [%s]\n", "token value", parsed.Value())

	case parsed.Type().HttpHeader() != "Bearer":
		t.Errorf("parsed token http header should be [%s], but is [%s]\n", "Bearer", parsed.Type().HttpHeader())

	case parsed.IssueTime().IsZero():
		t.Errorf("parsed issue time should not be zero\n")

	case parsed.ExpiryTime().IsZero():
		t.Errorf("parsed expiry time should not be zero\n")

	case len(parsed.Scopes()) != 2:
		t.Errorf("parsed scopes should have [%d] items, but has [%d]\n", 2, len(parsed.Scopes()))

	case len(parsed.Details()) != 2:
		t.Errorf("parsed details should have [%d] items, but has [%d]\n", 2, len(parsed.Details()))

	case parsed.Claims() != nil && len(parsed.Claims().Values()) != 0:
		t.Errorf("parsed claims should be empty (ignored), but got %v\n", parsed.Claims())

	case parsed.RefreshToken().Value() != "refresh token value":
		t.Errorf("parsed refresh token should be correct\n")

		//TODO more cases
	}
}

type TestTimeFormat time.Time

type TestAccessToken struct {
	AccessToken  string         `json:"access_token,omitempty"`
	ExpiresIn    float64        `json:"expires_in,omitempty"`
	Expiry       TestTimeFormat `json:"expiry"`
	Iat          TestTimeFormat `json:"iat"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	Scope        string         `json:"scope,omitempty"`
	TokenType    string         `json:"token_type,omitempty"`
}

// MarshalJSON is defined so we can set the iat and expiry time formatting correctly
func (t TestTimeFormat) MarshalJSON() ([]byte, error) {
	newTime := time.Time(t)
	return []byte(fmt.Sprintf("%q", newTime.Format(utils.ISO8601Seconds))), nil

}

func TestUnmarshal(t *testing.T) {
	// timeNow can be used as the replacement to the time.Now() call. That way
	// none of the times will be offset from each other
	timeNow := time.Now().UTC().Truncate(time.Second)
	tests := []struct {
		name            string
		testAccessToken TestAccessToken
		validator       func(t *testing.T, token DefaultAccessToken)
	}{
		{
			name: "Check Expires In",
			testAccessToken: TestAccessToken{
				AccessToken: "simple create",
				TokenType:   "bearer", Scope: "s1",
				ExpiresIn: 3600,
				Iat:       TestTimeFormat(timeNow),
			},
			validator: func(t *testing.T, token DefaultAccessToken) {
				expectedExpiry := timeNow.Add(3600 * time.Second).UTC()
				expiryTime := token.expiryTime.UTC()
				if !expectedExpiry.Equal(expiryTime) {
					t.Errorf("expected: %v , got: %v", expectedExpiry, expiryTime)
				}
			},
		},
		{
			name: "Check Expiry",
			testAccessToken: TestAccessToken{
				AccessToken: "simple create",
				TokenType:   "bearer", Scope: "s1",
				Expiry: TestTimeFormat(timeNow.Add(3600 * time.Second)),
				Iat:    TestTimeFormat(timeNow),
			},
			validator: func(t *testing.T, token DefaultAccessToken) {
				expectedExpiry := timeNow.Add(3600 * time.Second).UTC()
				expiryTime := token.expiryTime.UTC()
				if !expectedExpiry.Equal(expiryTime) {
					t.Errorf("expected: %v , got: %v", expectedExpiry, expiryTime)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken := DefaultAccessToken{}
			b, err := json.Marshal(tt.testAccessToken)
			if err != nil {
				t.Fatalf("unable to setup test: %v", err)
			}
			err = json.Unmarshal(b, &accessToken)
			if err != nil {
				t.Errorf("error unmarshalling: %v", err)
			}
			tt.validator(t, accessToken)
		})
	}
}
