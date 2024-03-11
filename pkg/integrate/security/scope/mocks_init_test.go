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
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/test/sectest"
	"go.uber.org/fx"
)

/*************************
	Mocks
 *************************/

type mocksDIOut struct {
	fx.Out
	AuthClient   seclient.AuthenticationClient
	TokenReader  oauth2.TokenStoreReader
	TokenRevoker sectest.MockedTokenRevoker
	Counter      InvocationCounter
}

func ProvideScopeMocksWithCounter(ctx *bootstrap.ApplicationContext) mocksDIOut {
	out := sectest.ProvideScopeMocks(ctx)
	c := NewCounter()
	return mocksDIOut{
		AuthClient: &MockedAuthenticationClient{
			AuthenticationClient: out.AuthClient,
			Counter:              c,
		},
		TokenReader: &MockedTokenStoreReader{
			TokenStoreReader: out.TokenReader,
			Counter:          c,
		},
		TokenRevoker: out.TokenRevoker,
		Counter:      c,
	}
}

type noopScopeManager struct{}

func (m *noopScopeManager) StartScope(ctx context.Context, _ *scope.Scope) (context.Context, error) {
	return ctx, nil
}

func (m *noopScopeManager) Start(ctx context.Context, _ ...scope.Options) (context.Context, error) {
	return ctx, nil
}

func (m *noopScopeManager) End(ctx context.Context) context.Context {
	return ctx
}

func ProvideNoopScopeManager() scope.ScopeManager {
	return &noopScopeManager{}
}

func NewCustomizer(c InvocationCounter) scope.ManagerCustomizer {
	return scope.ManagerCustomizerFunc(func() []scope.ManagerOptions {
		hook := TestScopeManagerHook{
			Counter: c,
		}
		return []scope.ManagerOptions{scope.BeforeStartHook(hook.Before), scope.AfterEndHook(hook.After)}
	})
}
