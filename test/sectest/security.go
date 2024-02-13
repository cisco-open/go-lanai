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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

/**************************
	Function
 **************************/

type SecurityContextOptions func(opt *SecurityContextOption)

type SecurityContextOption struct {
	// Authentication override any other fields
	Authentication security.Authentication
}

// ContextWithSecurity set given SecurityContextOptions in the given context, returning the new context
func ContextWithSecurity(ctx context.Context, opts ...SecurityContextOptions) context.Context {
	opt := SecurityContextOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	// with a no-op valuer, we force a new utils.MutableContext is created,
	// in order to preserve any security context in the original "ctx"
	testCtx := utils.MakeMutableContext(ctx, func(key interface{}) interface{} { return nil})
	security.MustSet(testCtx, opt.Authentication)
	return testCtx
}

// WithMockedSecurity used to mock an oauth2.Authentication in the given context, returning a new context
// Deprecated: use ContextWithSecurity(ctx, MockedAuthentication(opts...)) instead
func WithMockedSecurity(ctx context.Context, opts ...SecurityMockOptions) context.Context {
	return ContextWithSecurity(ctx, MockedAuthentication(opts...))
}

/**************************
	Options
 **************************/

func Authentication(auth security.Authentication) SecurityContextOptions {
	return func(opt *SecurityContextOption) {
		opt.Authentication = auth
	}
}

func MockedAuthentication(opts ...SecurityMockOptions) SecurityContextOptions {
	return func(opt *SecurityContextOption) {
		details := NewMockedSecurityDetails(opts...)
		user := oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
			opt.Principal = details.Username()
			opt.State = security.StateAuthenticated
			opt.Permissions = map[string]interface{}{}
			for perm := range details.Permissions() {
				opt.Permissions[perm] = true
			}
			opt.Details = details.KVs
		})
		token := &MockedToken{
			MockedTokenInfo: MockedTokenInfo{
				UName:       details.Username(),
				UID:         details.UserId(),
				TID:         details.TenantId(),
				TExternalId: details.TenantExternalId(),
				OrigU:       details.OrigUsername,
			},
			Token: details.AccessToken,
			ExpTime: details.Exp,
			IssTime: details.Iss,
		}

		auth := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
			opt.Request = oauth2.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
				opt.ClientId = details.ClientID
				opt.Scopes = details.Scopes
				opt.Approved = true
				opt.GrantType = details.OAuth2GrantType
				opt.ResponseTypes = utils.NewStringSetFrom(details.OAuth2ResponseTypes)
				for k, v := range details.OAuth2Parameters {
					opt.Parameters[k] = v
					opt.Extensions[k] = v
				}
				for k, v := range details.OAuth2Extensions {
					opt.Extensions[k] = v
				}
			})
			opt.Token = token
			opt.UserAuth = user
			opt.Details = details
		})
		opt.Authentication = auth
	}
}


