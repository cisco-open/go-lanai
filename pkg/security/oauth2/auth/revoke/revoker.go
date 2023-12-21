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

package revoke

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"fmt"
)

type RevokerOptions func(opt *RevokerOption)

type RevokerOption struct {
	AuthRegistry     auth.AuthorizationRegistry
	SessionName      string
	SessionStore     session.Store
	TokenStoreReader oauth2.TokenStoreReader
}

// DefaultAccessRevoker implements auth.AccessRevoker
type DefaultAccessRevoker struct {
	authRegistry     auth.AuthorizationRegistry
	sessionName      string
	sessionStore     session.Store
	tokenStoreReader oauth2.TokenStoreReader
}

func NewDefaultAccessRevoker(opts ...RevokerOptions) *DefaultAccessRevoker {
	opt := RevokerOption{
		SessionName: common.DefaultName,
	}
	for _, f := range opts {
		f(&opt)
	}
	return &DefaultAccessRevoker{
		authRegistry: opt.AuthRegistry,
		sessionName: opt.SessionName,
		sessionStore:     opt.SessionStore,
		tokenStoreReader: opt.TokenStoreReader,
	}
}

func (r DefaultAccessRevoker) RevokeWithSessionId(ctx context.Context, sessionId string, sessionName string) (err error) {
	// expire session
	if s, e := r.sessionStore.Get(sessionId, sessionName); e == nil && s != nil {
		if e := s.ExpireNow(ctx); e != nil {
			logger.WithContext(ctx).Warnf("Unable to expire session for session ID [%s]: %v", s.GetID(), e)
			err = e
		}
	}

	// revoke all tokens
	if e := r.authRegistry.RevokeSessionAccess(ctx, sessionId, true); e != nil {
		return e
	}
	return
}

func (r DefaultAccessRevoker) RevokeWithUsername(ctx context.Context, username string, revokeRefreshToken bool) (err error) {
	// expire all sessions
	if e := r.sessionStore.WithContext(ctx).InvalidateByPrincipalName(username, r.sessionName); e != nil {
		logger.WithContext(ctx).Warnf("Unable to expire session for username [%s]: %v", username, e)
		err = e
	}
	// revoke all tokens
	if e := r.authRegistry.RevokeUserAccess(ctx, username, revokeRefreshToken); e != nil {
		return e
	}
	return
}

func (r DefaultAccessRevoker) RevokeWithClientId(ctx context.Context, clientId string, revokeRefreshToken bool) error {
	return r.authRegistry.RevokeClientAccess(ctx, clientId, true)
}

func (r DefaultAccessRevoker) RevokeWithTokenValue(ctx context.Context, tokenValue string, hint auth.RevokerTokenHint) error {
	switch hint {
	case auth.RevokerHintAccessToken:
		token, e := r.tokenStoreReader.ReadAccessToken(ctx, tokenValue)
		if e != nil {
			return e
		}
		return r.authRegistry.RevokeAccessToken(ctx, token)
	case auth.RevokerHintRefreshToken:
		token, e := r.tokenStoreReader.ReadRefreshToken(ctx, tokenValue)
		if e != nil {
			return e
		}
		return r.authRegistry.RevokeRefreshToken(ctx, token)
	default:
		return fmt.Errorf("unsupported revoker token hint")
	}
}
