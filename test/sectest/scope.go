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
    "embed"
    appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    securityint "github.com/cisco-open/go-lanai/pkg/integrate/security"
    "github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
    "github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "go.uber.org/fx"
    "time"
)

//var logger = log.New("SEC.Test")

//go:embed test-scopes.yml
var defaultMockingConfigFS embed.FS

/**************************
	Options
 **************************/

// WithMockedScopes is a test.Options that initialize github.com/cisco-open/go-lanai/pkg/integrate/security/scope
// This option configure mocked security scopes based on yaml provided as embed.FS.
// If no config is provided, the default config is used
func WithMockedScopes(mocksConfigFS ...embed.FS) test.Options {
	fxOpts := make([]fx.Option, len(mocksConfigFS), len(mocksConfigFS) + 3)
	for i, fs := range mocksConfigFS {
		fxOpts[i] = appconfig.FxEmbeddedApplicationAdHoc(fs)
	}
	fxOpts = append(fxOpts,
		appconfig.FxEmbeddedBootstrapAdHoc(defaultMockingConfigFS),
		fx.Provide(securityint.BindSecurityIntegrationProperties),
		fx.Provide(ProvideScopeMocks),
	)
	opts := []test.Options{
		apptest.WithModules(scope.Module),
		apptest.WithFxOptions(fxOpts...),
	}
	return func(opt *test.T) {
		for _, fn := range opts {
			fn(opt)
		}
	}
}

/**************************
	fx options
 **************************/

type MocksDIOut struct {
	fx.Out
	AuthClient   seclient.AuthenticationClient
	TokenReader  oauth2.TokenStoreReader
	TokenRevoker MockedTokenRevoker
}

// ProvideScopeMocks is for internal usage. Exported for cross-package reference
// Try use WithMockedScopes instead
func ProvideScopeMocks(ctx *bootstrap.ApplicationContext) MocksDIOut {
	props := bindScopeMockingProperties(ctx)
	accounts := newMockedAccounts(props.Accounts.MapValues())
	tenants := newMockedTenants(props.Tenants.MapValues())
	base := mockedTokenBase{
		accounts: accounts,
		tenants:  tenants,
		revoked:  utils.NewStringSet(),
	}

	return MocksDIOut{
		AuthClient:   newMockedAuthClient(&base, time.Duration(props.TokenValidity)),
		TokenReader:  newMockedTokenStoreReader(&base),
		TokenRevoker: &base,
	}
}
