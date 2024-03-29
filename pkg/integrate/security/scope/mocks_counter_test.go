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

package scope_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
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

func NewCounter() InvocationCounter {
	return &counter{
		counts: map[interface{}]*uint64{},
	}
}

type counter struct {
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

type MockedAuthenticationClient struct {
	seclient.AuthenticationClient
	Counter InvocationCounter
}

func (c *MockedAuthenticationClient) PasswordLogin(ctx context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	c.Counter.Increase(seclient.AuthenticationClient.PasswordLogin, 1)
	return c.AuthenticationClient.PasswordLogin(ctx, opts...)
}

func (c *MockedAuthenticationClient) SwitchUser(ctx context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	c.Counter.Increase(seclient.AuthenticationClient.SwitchUser, 1)
	return c.AuthenticationClient.SwitchUser(ctx, opts...)
}

func (c *MockedAuthenticationClient) SwitchTenant(ctx context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	c.Counter.Increase(seclient.AuthenticationClient.SwitchTenant, 1)
	return c.AuthenticationClient.SwitchTenant(ctx, opts...)
}

type MockedTokenStoreReader struct {
	oauth2.TokenStoreReader
	Counter InvocationCounter
}

func (c *MockedTokenStoreReader) ReadAuthentication(ctx context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	c.Counter.Increase(oauth2.TokenStoreReader.ReadAuthentication, 1)
	return c.TokenStoreReader.ReadAuthentication(ctx, tokenValue, hint)
}

func (c *MockedTokenStoreReader) ReadAccessToken(ctx context.Context, value string) (oauth2.AccessToken, error) {
	c.Counter.Increase(oauth2.TokenStoreReader.ReadAccessToken, 1)
	return c.TokenStoreReader.ReadAccessToken(ctx, value)
}

func (c *MockedTokenStoreReader) ReadRefreshToken(ctx context.Context, value string) (oauth2.RefreshToken, error) {
	c.Counter.Increase(oauth2.TokenStoreReader.ReadRefreshToken, 1)
	return c.TokenStoreReader.ReadRefreshToken(ctx, value)
}


type TestScopeManagerHook struct {
	Counter InvocationCounter
}

func (h TestScopeManagerHook) Before(ctx context.Context, scope *scope.Scope) context.Context {
	h.Counter.Increase(TestScopeManagerHook.Before, 1)
	return ctx
}

func (h TestScopeManagerHook) After(ctx context.Context, scope *scope.Scope) context.Context {
	h.Counter.Increase(TestScopeManagerHook.After, 1)
	return ctx
}