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

package swagger

import (
	"embed"
	"fmt"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

//go:generate npm install --prefix nodejs
//go:generate npm run --prefix nodejs build --output_dir=../generated
//go:embed generated/*
var Content embed.FS

//go:embed defaults-swagger.yml
var defaultConfigFS embed.FS

var logger = log.New("Swagger")

var Module = &bootstrap.Module{
	Name:       "swagger",
	Precedence: bootstrap.SwaggerPrecedence,
	PriorityOptions: []fx.Option{
		fx.Invoke(configureSecurity),
	},
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			bindSwaggerProperties,
			fx.Annotate(newSwaggerInfoDiscoveryCustomizer, fx.ResultTags(fmt.Sprintf(`group:"%s"`, discovery.FxGroup))),
		),
		fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	Registrar            *web.Registrar
	Properties           SwaggerProperties
	Resolver             bootstrap.BuildInfoResolver `optional:"true"`
}

func initialize(di initDI) {
	di.Registrar.MustRegister(Content)
	di.Registrar.MustRegister(NewSwaggerController(di.Properties, di.Resolver))
}

func bindSwaggerProperties(ctx *bootstrap.ApplicationContext) SwaggerProperties {
	props := NewSwaggerSsoProperties()
	if err := ctx.Config().Bind(props, SwaggerPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SwaggerSsoProperties"))
	}
	return *props
}

type secDI struct {
	fx.In
	SecRegistrar security.Registrar `optional:"true"`
	Properties   SwaggerProperties
}

// configureSecurity register security.Configurer that control how security works on endpoints
func configureSecurity(di secDI) {
	if di.SecRegistrar != nil && di.Properties.Security.SecureDocs {
		di.SecRegistrar.Register(&swaggerSecurityConfigurer{})
	}
}

type DiscoveryCustomizerDIOut struct {
	fx.Out

	Customizer discovery.ServiceRegistrationCustomizer `group:"discovery_customizer"`
}
