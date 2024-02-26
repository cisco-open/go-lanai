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

package passwdidp

import (
    "embed"
    appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/web"
    "go.uber.org/fx"
)

//var logger = log.New("SEC.Passwd")

const (
	OrderWhiteLabelTemplateFS = 0
	OrderTemplateFSOverwrite  = OrderWhiteLabelTemplateFS - 1000
)

//go:embed web/whitelabel/*
var whiteLabelContent embed.FS

//go:embed defaults-passwd-auth.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name: "password IDP",
	Precedence: security.MaxSecurityPrecedence - 100,
	Options: []fx.Option {
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindPwdAuthProperties),
		fx.Invoke(register),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func register(r *web.Registrar) {
	r.MustRegister(web.OrderedFS(whiteLabelContent, OrderWhiteLabelTemplateFS))
}
