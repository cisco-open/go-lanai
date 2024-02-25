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
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
)

/***********************************
	default implmentation
 ***********************************/

// OAuth2ClientAccountStore wraps an delegate and implement both security.AccountStore and client oauth2.OAuth2ClientStore
type OAuth2ClientAccountStore struct {
	oauth2.OAuth2ClientStore
}

func WrapOAuth2ClientStore(clientStore oauth2.OAuth2ClientStore) *OAuth2ClientAccountStore {
	return &OAuth2ClientAccountStore{
		OAuth2ClientStore: clientStore,
	}
}

// security.AccountStore
func (s *OAuth2ClientAccountStore) LoadAccountById(ctx context.Context, id interface{}) (security.Account, error) {
	switch id.(type) {
	case string:
		return s.LoadAccountByUsername(ctx, id.(string))
	default:
		return nil, security.NewUsernameNotFoundError("invalid clientId type")
	}
}

// security.AccountStore
func (s *OAuth2ClientAccountStore) LoadAccountByUsername(ctx context.Context, username string) (security.Account, error) {

	if client, err := s.OAuth2ClientStore.LoadClientByClientId(ctx, username); err != nil {
		return nil, security.NewUsernameNotFoundError("invalid clientId")
	} else if acct, ok := client.(security.Account); !ok {
		return nil, security.NewInternalAuthenticationError("loaded client is not an account")
	} else {
		return acct, nil
	}
}

// security.AccountStore
func (s *OAuth2ClientAccountStore) LoadLockingRules(ctx context.Context, acct security.Account) (security.AccountLockingRule, error) {
	return nil, security.NewInternalAuthenticationError("client doesn't have locking rule")
}

// security.AccountStore
func (s *OAuth2ClientAccountStore) LoadPwdAgingRules(ctx context.Context, acct security.Account) (security.AccountPwdAgingRule, error) {
	return nil, security.NewInternalAuthenticationError("client doesn't have aging rule")
}

// security.AccountStore
func (s *OAuth2ClientAccountStore) Save(ctx context.Context, acct security.Account) error {
	return security.NewInternalAuthenticationError("client is inmutable during authentication")
}
