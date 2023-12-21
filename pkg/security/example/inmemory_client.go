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

package example

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

// in memory oauth2.OAuth2ClientStore
type InMemoryClientStore struct {
	lookupByClientId map[string]*auth.DefaultOAuth2Client
}

func NewInMemoryClientStore(props ClientsProperties) oauth2.OAuth2ClientStore {
	lookup := make(map[string]*auth.DefaultOAuth2Client)
	for _, v := range props.Clients {
		lookup[v.ClientId] = newOAuth2Client(v)
	}
	return &InMemoryClientStore{
		lookupByClientId: lookup,
	}
}

func (s *InMemoryClientStore) LoadClientByClientId(c context.Context, clientId string) (oauth2.OAuth2Client, error) {
	if client, ok := s.lookupByClientId[clientId]; ok {
		return client, nil
	}
	return nil, oauth2.NewClientNotFoundError("invalid client")
}

func newOAuth2Client(props PropertiesBasedClient) *auth.DefaultOAuth2Client {
	return &auth.DefaultOAuth2Client{
		ClientDetails: auth.ClientDetails{
			ClientId:             props.ClientId,
			Secret:               props.Secret,
			GrantTypes:           utils.NewStringSet(props.GrantTypes...),
			RedirectUris:         utils.NewStringSet(props.RedirectUris...),
			Scopes:               utils.NewStringSet(props.Scopes...),
			AutoApproveScopes:    utils.NewStringSet(props.AutoApproveScopes...),
			AccessTokenValidity:  utils.ParseDuration(props.AccessTokenValidity),
			RefreshTokenValidity: utils.ParseDuration(props.RefreshTokenValidity),
			UseSessionTimeout:    props.UseSessionTimeout,
			AssignedTenantIds:    utils.NewStringSet(props.TenantRestrictions...),
		},
	}
}
