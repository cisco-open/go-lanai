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

package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"reflect"
	"sync"
	"sync/atomic"
)

type InvocationCounter interface {
	Get(fn interface{}) int
	Increase(fn interface{}, increment uint64)
	Reset(fn interface{})
	ResetAll()
}

type counter struct {
	seclient.AuthenticationClient
	oauth2.TokenStoreReader
	mtx sync.RWMutex
	counts map[interface{}]*uint64
}

func (c *counter) Get(fn interface{}) int {
	ptr := c.get(fn)
	if ptr == nil {
		return 0
	}
	return int(atomic.LoadUint64(ptr))
}

func (c *counter) Increase(fn interface{}, increment uint64) {
	ptr := c.get(fn)
	if ptr == nil {
		ptr = c.new(fn)
	}
	atomic.AddUint64(ptr, increment)
}

func (c *counter) Reset(fn interface{}) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.counts, c.key(fn))
}

func (c *counter) ResetAll() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.counts = map[interface{}]*uint64{}
}

func (c *counter) key(fn interface{}) uintptr {
	return reflect.ValueOf(fn).Pointer()
}

func (c *counter) get(fn interface{}) *uint64 {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	p, ok := c.counts[c.key(fn)]
	if !ok {
		return nil
	}
	return p
}

func (c *counter) new(fn interface{}) *uint64 {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	p, ok := c.counts[c.key(fn)]
	if !ok {
		var n uint64
		p = &n
		c.counts[c.key(fn)] = p
	}
	return p
}

func (c *counter) PasswordLogin(ctx context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	c.Increase(seclient.AuthenticationClient.PasswordLogin, 1)
	return c.AuthenticationClient.PasswordLogin(ctx, opts...)
}

func (c *counter) SwitchUser(ctx context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	c.Increase(seclient.AuthenticationClient.SwitchUser, 1)
	return c.AuthenticationClient.SwitchUser(ctx, opts...)
}

func (c *counter) SwitchTenant(ctx context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	c.Increase(seclient.AuthenticationClient.SwitchTenant, 1)
	return c.AuthenticationClient.SwitchTenant(ctx, opts...)
}

func (c *counter) ReadAuthentication(ctx context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	c.Increase(oauth2.TokenStoreReader.ReadAuthentication, 1)
	return c.TokenStoreReader.ReadAuthentication(ctx, tokenValue, hint)
}

func (c *counter) ReadAccessToken(ctx context.Context, value string) (oauth2.AccessToken, error) {
	c.Increase(oauth2.TokenStoreReader.ReadAccessToken, 1)
	return c.TokenStoreReader.ReadAccessToken(ctx, value)
}

func (c *counter) ReadRefreshToken(ctx context.Context, value string) (oauth2.RefreshToken, error) {
	c.Increase(oauth2.TokenStoreReader.ReadRefreshToken, 1)
	return c.TokenStoreReader.ReadRefreshToken(ctx, value)
}

