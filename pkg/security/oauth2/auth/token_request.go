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
)

type TokenRequest struct {
	Parameters map[string]string
	ClientId   string
	Scopes     utils.StringSet
	GrantType  string
	Extensions map[string]interface{}
	context    utils.MutableContext
}

func (r *TokenRequest) Context() utils.MutableContext {
	return r.context
}

func (r *TokenRequest) WithContext(ctx context.Context) *TokenRequest {
	r.context = utils.MakeMutableContext(ctx)
	return r
}

func (r *TokenRequest) OAuth2Request(client oauth2.OAuth2Client) oauth2.OAuth2Request {
	return oauth2.NewOAuth2Request(func(details *oauth2.RequestDetails) {
		details.Parameters = r.Parameters
		details.ClientId = client.ClientId()
		details.Scopes = r.Scopes
		details.Approved = true
		details.GrantType = r.GrantType
		details.Extensions = r.Extensions
	})
}

func NewTokenRequest() *TokenRequest {
	return &TokenRequest{
		Parameters:    map[string]string{},
		Scopes:        utils.NewStringSet(),
		Extensions:    map[string]interface{}{},
		context:       utils.NewMutableContext(context.Background()),
	}
}

func ParseTokenRequest(req *http.Request) (*TokenRequest, error) {
	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	values := flattenValuesToMap(req.Form);
	return &TokenRequest{
		Parameters:    toStringMap(values),
		ClientId:      extractStringParam(oauth2.ParameterClientId, values),
		Scopes:        extractStringSetParam(oauth2.ParameterScope, " ", values),
		GrantType:     extractStringParam(oauth2.ParameterGrantType, values),
		Extensions:    values,
		context:       utils.MakeMutableContext(req.Context()),
	}, nil
}

func (r *TokenRequest) String() string {
	return fmt.Sprintf("[client=%s, grant=%s, scope=%s, ext=%s]",
		r.ClientId, r.GrantType, r.Scopes, r.Extensions)
}