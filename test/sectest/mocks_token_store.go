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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/utils"
)

type mockedTokenStoreReader struct {
	*mockedTokenBase
}

// NewMockedTokenStoreReader create a mocked oauth2.TokenStoreReader based on properties,
// The returned reader also implements MockedTokenRevoker
func NewMockedTokenStoreReader(acctsProps map[string]*MockedAccountProperties, tenantProps map[string]*MockedTenantProperties) oauth2.TokenStoreReader {
	accounts := newMockedAccounts(acctsProps)
	tenants := newMockedTenants(tenantProps)
	return &mockedTokenStoreReader{
		mockedTokenBase: &mockedTokenBase{
			accounts: accounts,
			tenants:  tenants,
			revoked:  utils.NewStringSet(),
		},
	}
}

func newMockedTokenStoreReader(base *mockedTokenBase) oauth2.TokenStoreReader {
	return &mockedTokenStoreReader{
		mockedTokenBase: base,
	}
}

func (r *mockedTokenStoreReader) ReadAuthentication(_ context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	if hint != oauth2.TokenHintAccessToken {
		return nil, fmt.Errorf("[Mocked Error] wrong token hint")
	}
	mt, e := r.parseMockedToken(tokenValue)
	if e != nil {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	acct, ok := r.accounts.lookup[mt.UName]
	if !ok {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	auth := r.newMockedAuth(mt, acct)
	return auth, nil
}

func (r *mockedTokenStoreReader) ReadAccessToken(_ context.Context, value string) (oauth2.AccessToken, error) {
	mt, e := r.parseMockedToken(value)
	if e != nil {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	_, ok := r.accounts.lookup[mt.UName]
	if !ok {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}
	return mt, nil
}

func (r *mockedTokenStoreReader) ReadRefreshToken(_ context.Context, _ string) (oauth2.RefreshToken, error) {
	return nil, fmt.Errorf("ReadRefreshToken is not implemented for mocked token store")
}
