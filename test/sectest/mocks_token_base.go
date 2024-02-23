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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"sync"
	"time"
)

/*************************
	Interface
 *************************/

type MockedTokenRevoker interface {
	Revoke(value string)
	RevokeAll()
}

/*************************
	Base
 *************************/

// mockedTokenBase implements MockedTokenRevoker, this serves as a base of multiple mock implementation.
// e.g. mockedTokenStoreReader, mockedAuthClient
type mockedTokenBase struct {
	sync.RWMutex
	accounts  *mockedAccounts
	tenants   *mockedTenants
	revoked   utils.StringSet
	notBefore time.Time
}

func (b *mockedTokenBase) Revoke(value string) {
	b.Lock()
	defer b.Unlock()
	b.revoked.Add(value)
}

func (b *mockedTokenBase) RevokeAll() {
	b.Lock()
	defer b.Unlock()
	b.notBefore = time.Now().UTC()
}

func (b *mockedTokenBase) isTokenRevoked(token *MockedToken, value string) bool {
	b.RLock()
	defer b.RUnlock()
	return token.IssTime.Before(b.notBefore) || b.revoked.Has(value)
}

func (b *mockedTokenBase) newMockedToken(acct *MockedAccount, tenant *mockedTenant, exp time.Time, origUser string) *MockedToken {
	return &MockedToken{
		MockedTokenInfo: MockedTokenInfo{
			UName:       acct.MockedAccountDetails.Username,
			UID:         acct.UserId,
			TID:         tenant.ID,
			TExternalId: tenant.ExternalId,
			OrigU:       origUser,
		},
		ExpTime: exp,
		IssTime: time.Now().UTC(),
	}
}

func (b *mockedTokenBase) parseMockedToken(value string) (*MockedToken, error) {
	mt := &MockedToken{}
	if e := mt.UnmarshalText([]byte(value)); e != nil {
		return nil, e
	}
	if b.isTokenRevoked(mt, value) {
		return nil, fmt.Errorf("[Mocked Error]: token revoked")
	}
	return mt, nil
}

func (b *mockedTokenBase) newMockedAuth(mt *MockedToken, acct *MockedAccount) oauth2.Authentication {
	user := oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
		opt.Principal = mt.UName
		opt.State = security.StateAuthenticated
		opt.Permissions = map[string]interface{}{}
		for perm := range acct.MockedAccountDetails.Permissions {
			opt.Permissions[perm] = true
		}
	})
	details := NewMockedSecurityDetails(func(d *SecurityDetailsMock) {
		*d = SecurityDetailsMock{
			Username:         acct.Username(),
			UserId:           acct.UserId,
			TenantExternalId: mt.TExternalId,
			TenantId:         mt.TID,
			Exp:              mt.ExpTime,
			Iss:              mt.IssTime,
			Permissions:      acct.MockedAccountDetails.Permissions,
			Tenants:          acct.AssignedTenants,
			OrigUsername:     mt.OrigU,
		}
	})
	return oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = oauth2.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
			opt.ClientId = mt.ClientID
			if len(mt.ClientID) == 0 {
				opt.ClientId = "mock"
			}
			opt.Approved = true
			opt.Scopes = utils.NewStringSet(mt.MockedTokenInfo.Scopes...)
		})
		opt.Token = mt
		opt.UserAuth = user
		opt.Details = details
	})
}
