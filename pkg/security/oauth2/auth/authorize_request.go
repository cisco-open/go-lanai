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

package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AuthorizeRequest struct {
	Parameters    map[string]string
	ClientId      string
	ResponseTypes utils.StringSet
	Scopes        utils.StringSet
	RedirectUri   string
	State         string
	Extensions    map[string]interface{}
	Approved      bool
	context		  utils.MutableContext

	// resource IDs is removed from OAuth2 Specs.
	// For backward compatibility, we use client's registered values or hard code it to "nfv-api"
}

func (r *AuthorizeRequest) Context() utils.MutableContext {
	return r.context
}

func (r *AuthorizeRequest) WithContext(ctx context.Context) *AuthorizeRequest {
	r.context = utils.MakeMutableContext(ctx)
	return r
}

func (r *AuthorizeRequest) OAuth2Request() oauth2.OAuth2Request {
	return oauth2.NewOAuth2Request(func(details *oauth2.RequestDetails) {
		if grantType, ok := r.Parameters[oauth2.ParameterGrantType]; ok {
			details.GrantType = grantType
		}

		details.Parameters = r.Parameters
		details.ClientId = r.ClientId
		details.Scopes = r.Scopes
		details.Approved = true
		details.RedirectUri = r.RedirectUri
		details.ResponseTypes = r.ResponseTypes
		details.Extensions = r.Extensions
	})
}

func (r *AuthorizeRequest) String() string {
	return fmt.Sprintf("[client=%s, response_type=%s, redirect=%s, scope=%s, ext=%s]",
		r.ClientId, r.ResponseTypes, r.RedirectUri, r.Scopes, r.Extensions)
}

func NewAuthorizeRequest(opts ...func(req *AuthorizeRequest)) *AuthorizeRequest {
	ar := AuthorizeRequest{
		Parameters:    map[string]string{},
		ResponseTypes: utils.NewStringSet(),
		Scopes:        utils.NewStringSet(),
		Extensions:    map[string]interface{}{},
		context:       utils.NewMutableContext(context.Background()),
	}
	for _, fn := range opts {
		fn(&ar)
	}
	return &ar
}

func ParseAuthorizeRequest(req *http.Request) (*AuthorizeRequest, error) {
	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	values := flattenValuesToMap(req.Form)
	return ParseAuthorizeRequestWithKVs(req.Context(), values)
}

func ParseAuthorizeRequestWithKVs(ctx context.Context, values map[string]interface{}) (*AuthorizeRequest, error) {
	return &AuthorizeRequest{
		Parameters:    toStringMap(values),
		ClientId:      extractStringParam(oauth2.ParameterClientId, values),
		ResponseTypes: extractStringSetParam(oauth2.ParameterResponseType, " ", values),
		Scopes:        extractStringSetParam(oauth2.ParameterScope, " ", values),
		RedirectUri:   extractStringParam(oauth2.ParameterRedirectUri, values),
		State:         extractStringParam(oauth2.ParameterState, values),
		Extensions:    values,
		context:       utils.MakeMutableContext(ctx),
	}, nil
}

/************************
	Helpers
 ************************/
func flattenValuesToMap(src url.Values) (dest map[string]interface{}) {
	dest = map[string]interface{}{}
	for k, v := range src {
		if len(v) == 0 {
			continue
		}
		dest[k] = strings.Join(v, " ")
	}
	return
}

func toStringMap(src map[string]interface{}) (dest map[string]string) {
	dest = map[string]string{}
	for k, v := range src {
		switch v.(type) {
		case string:
			dest[k] = v.(string)
		case fmt.Stringer:
			dest[k] = v.(fmt.Stringer).String()
		}
	}
	return
}

func extractStringParam(key string, params map[string]interface{}) string {
	if v, ok := params[key]; ok {
		delete(params, key)
		return v.(string)
	}
	return ""
}

func extractStringSetParam(key, sep string, params map[string]interface{}) utils.StringSet {
	if v, ok := params[key]; ok {
		delete(params, key)
		return utils.NewStringSet(strings.Split(v.(string), sep)...)
	}
	return utils.NewStringSet()
}