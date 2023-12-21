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

package seclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"encoding/base64"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Test_remoteAuthClient_withClientAuth(t *testing.T) {
	type fields struct {
		clientId     string
		clientSecret string
	}
	type args struct {
		opt *AuthOption
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantAuthString string
	}{
		{
			name: "Test that AuthOption ID/Secret have higher priority",
			fields: fields{
				clientId:     "should not be this",
				clientSecret: "should not be this",
			},
			args: args{
				opt: &AuthOption{
					ClientID:     "ClientID",
					ClientSecret: "ClientSecret",
				},
			},
			wantAuthString: "ClientID:ClientSecret",
		},
		{
			name: "Test fallback",
			fields: fields{
				clientId:     "UseThisID",
				clientSecret: "UseThisSecret",
			},
			args: args{
				opt: &AuthOption{
					ClientID:     "",
					ClientSecret: "",
				},
			},
			wantAuthString: "UseThisID:UseThisSecret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &remoteAuthClient{
				clientId:     tt.fields.clientId,
				clientSecret: tt.fields.clientSecret,
			}
			req := httpclient.NewRequest("", "")
			// Use the AuthOptions that the withClientAuth returns to apply to our blank request
			options := c.withClientAuth(tt.args.opt)
			options(req)
			// Check the request headers for what we expect. Need to break apart and decode some parts first
			fullAuth := req.Headers.Get(httpclient.HeaderAuthorization)
			authSplit := strings.Split(fullAuth, " ")
			if len(authSplit) != 2 {
				t.Fatalf("expected two parts of the header auth, got: %v", len(authSplit))
			}
			b64auth := authSplit[1] // take the base64 encoded username + ":" + password
			auth, err := base64.StdEncoding.DecodeString(b64auth)
			if err != nil {
				t.Fatalf("unable to decode b64auth: %v", err)
			}
			if string(auth) != tt.wantAuthString {
				t.Errorf("expected: %v, got: %v", tt.wantAuthString, string(auth))
			}
		})
	}
}

func TestWithNonEmptyURLValues(t *testing.T) {
	type args struct {
		values map[string][]string
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{
			name: "two keys, with two values",
			args: args{values: map[string][]string{
				"key1": {"value1", "value2"},
				"key2": {"value1", "value2"},
			}},
			want: url.Values{
				"key1": {"value1", "value2"},
				"key2": {"value1", "value2"},
			},
		},
		{
			name: "two keys, second key has empty values",
			args: args{values: map[string][]string{
				"key1": {"value1", "value2"},
				"key2": {},
			}},
			want: url.Values{
				"key1": {"value1", "value2"},
			},
		},
		{
			name: "two keys, both have empty values",
			args: args{values: map[string][]string{
				"key1": {},
				"key2": {},
			}},
			want: url.Values{},
		},
		{
			name: "two keys, first has empty values",
			args: args{values: map[string][]string{
				"key1": {},
				"key2": {"value1", "value2"},
			}},
			want: url.Values{
				"key2": {"value1", "value2"},
			},
		},
		{
			name: "two keys, first has empty values, but using urlValues",
			args: args{values: url.Values{
				"key1": {},
				"key2": {"value1", "value2"},
			}},
			want: url.Values{
				"key2": {"value1", "value2"},
			},
		},
		{
			name: "two keys, first has empty values, but using urlValues",
			args: args{values: url.Values{
				"key1": {""},
				"key2": {"value1", "value2"},
			}},
			want: url.Values{
				"key2": {"value1", "value2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithNonEmptyURLValues(tt.args.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithNonEmptyURLValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
