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
	"reflect"
	"strings"
	"time"
)

const (
	errTmplFieldExpectString  = `invalid field type. expected string`
	errTmplFieldExpectInt     = `invalid field type. expected integer`
	errTmplFieldExpectISO8601 = `invalid field format. expected ISO8601 formatted string`
)

type valueConverterFunc func(v interface{}) (reflect.Value, error)

/************************
	DefaultAccessToken
 ************************/

var accessTokenIgnoredDetails = utils.NewStringSet(
	JsonFieldAccessTokenValue, JsonFieldTokenType, JsonFieldScope,
	JsonFieldExpiryTime, JsonFieldIssueTime, JsonFieldExpiresIn, JsonFieldRefreshTokenValue)

var scopeSeparator = " "

// MarshalJSON implements json.Marshaler
func (t *DefaultAccessToken) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	for k, v := range t.details {
		data[k] = v
	}
	data[JsonFieldAccessTokenValue] = t.value
	data[JsonFieldTokenType] = t.tokenType
	data[JsonFieldScope] = strings.Join(t.scopes.Values(), scopeSeparator)
	data[JsonFieldIssueTime] = t.issueTime.Format(utils.ISO8601Seconds)
	if !t.expiryTime.IsZero() {
		data[JsonFieldExpiryTime] = t.expiryTime.Format(utils.ISO8601Seconds)
		data[JsonFieldExpiresIn] = int(t.expiryTime.Sub(time.Now()).Seconds())
	}

	if t.refreshToken != nil {
		data[JsonFieldRefreshTokenValue] = t.refreshToken
	}

	return json.Marshal(data)
}

// UnmarshalJSON implements json.Unmarshaler
func (t *DefaultAccessToken) UnmarshalJSON(data []byte) error {
	parsed := map[string]interface{}{}

	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	if err := extractField(parsed, JsonFieldAccessTokenValue, true, &t.value, anyToString); err != nil {
		return err
	}

	if err := extractField(parsed, JsonFieldTokenType, true, &t.tokenType, stringToTokenType); err != nil {
		return err
	}

	if err := extractField(parsed, JsonFieldScope, true, &t.scopes, stringSliceToStringSet); err != nil {
		return err
	}

	// issue time is optional
	if err := extractField(parsed, JsonFieldIssueTime, false, &t.issueTime, expiryToTime); err != nil {
		return err
	}

	// default to parse expiry time from JsonFieldExpiryTime field, fall back to JsonFieldExpiresIn
	// sets required to true so we can fallback to JsonFieldExpiresIn.
	if err := extractField(parsed, JsonFieldExpiryTime, true, &t.expiryTime, expiryToTime); err != nil {
		if err := extractField(parsed, JsonFieldExpiresIn, true, &t.expiryTime, expireInToTimeConverter(t.issueTime)); err != nil {
			return err
		}
	}

	if err := extractField(parsed, JsonFieldRefreshTokenValue, false, &t.refreshToken, stringToRefreshToken); err != nil {
		return err
	}

	// put the rest of fields to details
	for k, v := range parsed {
		if !accessTokenIgnoredDetails.Has(k) {
			t.details[k] = v
		}
	}
	return nil
}

/************************
	DefaultRefreshToken
 ************************/

// MarshalJSON implements json.Marshaler, only DefaultRefreshToken.value is serialized
func (t *DefaultRefreshToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.value)
}

// UnmarshalJSON implements json.Unmarshaler
func (t *DefaultRefreshToken) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.value)
}

/************************
	Helpers
 ************************/

func extractField(data map[string]interface{}, field string, required bool, destPtr interface{}, converter valueConverterFunc) error {
	v, ok := data[field]
	switch {
	case !ok && required:
		return fmt.Errorf("cannot find required field [%s]", field)
	case !ok:
		return nil
	}

	value, err := converter(v)
	if err != nil {
		return fmt.Errorf("cannot parse field [%s]: %s", field, err.Error())
	}

	dest := reflect.ValueOf(destPtr)
	if !dest.CanSet() {
		dest = dest.Elem()
	}

	dest.Set(value)
	return nil
}

func anyToString(v interface{}) (reflect.Value, error) {
	_, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf(errTmplFieldExpectString)
	}
	return reflect.ValueOf(v), nil
}

func stringToTokenType(v interface{}) (reflect.Value, error) {
	s, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf(errTmplFieldExpectString)
	}
	return reflect.ValueOf(TokenType(s)), nil
}

func stringSliceToStringSet(v interface{}) (reflect.Value, error) {
	stringSlice, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf(errTmplFieldExpectString)
	}

	slice := strings.Split(stringSlice, scopeSeparator)
	scopes := utils.NewStringSet()
	for _, s := range slice {
		scopes.Add(s)
	}
	return reflect.ValueOf(scopes), nil
}

func expiryToTime(v interface{}) (reflect.Value, error) {
	str, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf(errTmplFieldExpectISO8601)
	}

	if t := utils.ParseTimeISO8601(str); !t.IsZero() {
		return reflect.ValueOf(t), nil
	} else if t := utils.ParseTime(utils.ISO8601Milliseconds, str); !t.IsZero() {
		return reflect.ValueOf(t), nil
	}

	return reflect.Value{}, fmt.Errorf(errTmplFieldExpectISO8601)
}

func expireInToTimeConverter(issueTime time.Time) valueConverterFunc {
	return func(v interface{}) (reflect.Value, error) {
		secs, ok := v.(float64)
		if !ok {
			return reflect.Value{}, fmt.Errorf(errTmplFieldExpectInt)
		}

		if issueTime.IsZero() {
			issueTime = time.Now()
		}
		t := issueTime.Add(time.Duration(secs) * time.Second)
		return reflect.ValueOf(t), nil
	}
}

func stringToRefreshToken(v interface{}) (reflect.Value, error) {
	s, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf(errTmplFieldExpectString)
	}
	return reflect.ValueOf(NewDefaultRefreshToken(s)), nil
}
