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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

type AuthOptions func(opt *AuthOption)

type AuthOption struct {
	Password         string   // Password is used by password login
	AccessToken      string   // AccessToken is used by switch user/tenant
	Username         string   // Username is used by password login and switch user
	UserId           string   // UserId is used by switch user
	TenantId         string   // TenantId is used by password login and switch user/tenant
	TenantExternalId string   // TenantExternalId is used by password login and switch user/tenant
	Scopes           []string // OAuth Scopes option
	ClientID         string   // ClientID that is used for the client credentials auth flow
	ClientSecret     string   // ClientSecret that is used for the client credentials auth flow
}

type AuthenticationClient interface {
	PasswordLogin(ctx context.Context, opts ...AuthOptions) (*Result, error)
	ClientCredentials(ctx context.Context, opts ...AuthOptions) (*Result, error)
	SwitchUser(ctx context.Context, opts ...AuthOptions) (*Result, error)
	SwitchTenant(ctx context.Context, opts ...AuthOptions) (*Result, error)
}

type Result struct {
	Request oauth2.OAuth2Request
	Token   oauth2.AccessToken
}

/****************************
	AuthOptions
 ****************************/

func WithCredentials(username, password string) AuthOptions {
	return func(opt *AuthOption) {
		opt.Username = username
		opt.Password = password
	}
}

func WithCurrentSecurity(ctx context.Context) AuthOptions {
	return WithAuthentication(security.Get(ctx))
}

func WithAuthentication(auth security.Authentication) AuthOptions {
	oauth, ok := auth.(oauth2.Authentication)
	if !ok {
		return noop()
	}
	return WithAccessToken(oauth.AccessToken().Value())
}

func WithAccessToken(accessToken string) AuthOptions {
	return func(opt *AuthOption) {
		opt.AccessToken = accessToken
	}
}

// WithTenant create an options that specify tenant by either tenantId or tenantExternalId
// username and userId are exclusive, cannot be both empty
func WithTenant(tenantId string, tenantExternalId string) AuthOptions {
	if tenantId != "" {
		return WithTenantId(tenantId)
	} else {
		return WithTenantExternalId(tenantExternalId)
	}
}

func WithTenantId(tenantId string) AuthOptions {
	return func(opt *AuthOption) {
		opt.TenantId = tenantId
		opt.TenantExternalId = ""
	}
}

func WithTenantExternalId(tenantExternalId string) AuthOptions {
	return func(opt *AuthOption) {
		opt.TenantId = ""
		opt.TenantExternalId = tenantExternalId
	}
}

// WithUser create an options that specify user by either username or userId
// username and userId are exclusive, cannot be both empty
func WithUser(userId string, username string) AuthOptions {
	if username != "" {
		return WithUsername(username)
	} else {
		return WithUserId(userId)
	}
}

func WithUsername(username string) AuthOptions {
	return func(opt *AuthOption) {
		opt.Username = username
		opt.UserId = ""
	}
}

func WithUserId(userId string) AuthOptions {
	return func(opt *AuthOption) {
		opt.Username = ""
		opt.UserId = userId
	}
}

func WithScope(scope []string) AuthOptions {
	return func(opt *AuthOption) {
		opt.Scopes = scope
	}
}

func WithClientAuth(clientID, secret string) AuthOptions {
	return func(opt *AuthOption) {
		opt.ClientID = clientID
		opt.ClientSecret = secret
	}
}

func noop() func(opt *AuthOption) {
	return func(_ *AuthOption) {
		// noop
	}
}
