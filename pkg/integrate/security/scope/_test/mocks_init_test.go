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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
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

func provideScopeMocksWithCounter(ctx *bootstrap.ApplicationContext) mocksDIOut {
	out := sectest.ProvideScopeMocks(ctx)
	counter := counter{
		AuthenticationClient: out.AuthClient,
		TokenStoreReader:     out.TokenReader,
		counts:               map[interface{}]*uint64{},
	}
	return mocksDIOut{
		AuthClient:   &counter,
		TokenReader:  &counter,
		TokenRevoker: out.TokenRevoker,
		Counter:      &counter,
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

func provideNoopScopeManager() scope.ScopeManager {
	return &noopScopeManager{}
}
