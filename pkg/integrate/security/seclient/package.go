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

package seclient

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	securityint "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Client")

var Module = &bootstrap.Module{
	Name: "auth-client",
	Precedence: bootstrap.SecurityIntegrationPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(securityint.DefaultConfigFS),
		fx.Provide(securityint.BindSecurityIntegrationProperties),
		fx.Provide(provideAuthClient),
	},
}

func Use() {
	httpclient.Use()
	bootstrap.Register(Module)
}

type clientDI struct {
	fx.In
	HttpClient  httpclient.Client
	Properties securityint.SecurityIntegrationProperties
}

func provideAuthClient(di clientDI) AuthenticationClient {
	return NewRemoteAuthClient(func(opt *AuthClientOption) {
		opt.Client = di.HttpClient
		opt.ServiceName = di.Properties.ServiceName
		opt.ClientId = di.Properties.Client.ClientId
		opt.ClientSecret = di.Properties.Client.ClientSecret
		opt.BaseUrl = di.Properties.Endpoints.BaseUrl
		opt.PwdLoginPath = di.Properties.Endpoints.PasswordLogin
		opt.SwitchContextPath = di.Properties.Endpoints.SwitchContext
	})
}

