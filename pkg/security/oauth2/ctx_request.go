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
)

/******************************
	OAuth2Request
******************************/
var excludedParameters = utils.NewStringSet(ParameterPassword, ParameterClientSecret)

//goland:noinspection GoNameStartsWithPackageName
type OAuth2Request interface {
	Parameters() map[string]string
	ClientId() string
	Scopes() utils.StringSet
	Approved() bool
	GrantType() string
	RedirectUri() string
	ResponseTypes() utils.StringSet
	Extensions() map[string]interface{}
	NewOAuth2Request(...RequestOptionsFunc) OAuth2Request
}

/******************************
	Implementation
******************************/

type RequestDetails struct {
	Parameters    map[string]string      `json:"parameters"`
	ClientId      string                 `json:"clientId"`
	Scopes        utils.StringSet        `json:"scope"`
	Approved      bool                   `json:"approved"`
	GrantType     string                 `json:"grantType"`
	RedirectUri   string                 `json:"redirectUri"`
	ResponseTypes utils.StringSet        `json:"responseTypes"`
	Extensions    map[string]interface{} `json:"extensions"`
}

type RequestOptionsFunc func(opt *RequestDetails)

type oauth2Request struct {
	RequestDetails
}

func NewOAuth2Request(optFuncs ...RequestOptionsFunc) OAuth2Request {
	request := oauth2Request{ RequestDetails: RequestDetails{
		Parameters: map[string]string{},
		Scopes: utils.NewStringSet(),
		ResponseTypes: utils.NewStringSet(),
		Extensions: map[string]interface{}{},
	}}

	for _, optFunc := range optFuncs {
		optFunc(&request.RequestDetails)
	}

	for param := range excludedParameters {
		delete(request.RequestDetails.Parameters, param)
	}
	return &request
}

func (r *oauth2Request) Parameters() map[string]string {
	return r.RequestDetails.Parameters
}

func (r *oauth2Request) ClientId() string {
	return r.RequestDetails.ClientId
}

func (r *oauth2Request) Scopes() utils.StringSet {
	return r.RequestDetails.Scopes
}

func (r *oauth2Request) Approved() bool {
	return r.RequestDetails.Approved
}

func (r *oauth2Request) GrantType() string {
	return r.RequestDetails.GrantType
}

func (r *oauth2Request) RedirectUri() string {
	return r.RequestDetails.RedirectUri
}

func (r *oauth2Request) ResponseTypes() utils.StringSet {
	return r.RequestDetails.ResponseTypes
}

func (r *oauth2Request) Extensions() map[string]interface{} {
	return r.RequestDetails.Extensions
}

func (r *oauth2Request) NewOAuth2Request(additional ...RequestOptionsFunc) OAuth2Request {
	all := append([]RequestOptionsFunc{r.copyFunc()}, additional...)
	return NewOAuth2Request(all...)
}

func (r *oauth2Request) copyFunc() RequestOptionsFunc {
	return func(opt *RequestDetails) {
		opt.ClientId = r.RequestDetails.ClientId
		opt.Scopes = r.RequestDetails.Scopes.Copy()
		opt.Approved = r.RequestDetails.Approved
		opt.GrantType = r.RequestDetails.GrantType
		opt.RedirectUri = r.RequestDetails.RedirectUri
		opt.ResponseTypes = r.RequestDetails.ResponseTypes
		for k, v := range r.RequestDetails.Parameters {
			opt.Parameters[k] = v
		}
		for k, v := range r.RequestDetails.Extensions {
			opt.Extensions[k] = v
		}
	}
}

// MarshalJSON json.Marshaler
func (r *oauth2Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.RequestDetails)
}

// UnmarshalJSON json.Unmarshaler
func (r *oauth2Request) UnmarshalJSON(data []byte) error {
	if e := json.Unmarshal(data, &r.RequestDetails); e != nil {
		return e
	}
	return nil
}


