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

package opainit

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opainput "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/input"
	"embed"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"go.uber.org/fx"
)

var logger = log.New("OPA.Init")

//go:embed defaults-opa.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Precedence: bootstrap.SecurityPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindProperties, ProvideEmbeddedOPA),
		fx.Invoke(InitializeEmbeddedOPA, RegisterHealth),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

func BindProperties(ctx *bootstrap.ApplicationContext) opa.Properties {
	props := opa.NewProperties()
	if e := ctx.Config().Bind(props, opa.PropertiesPrefix); e != nil {
		panic(fmt.Errorf("failed to bind OPA properties: %v", e))
	}
	return *props
}

type EmbeddedOPAOut struct {
	fx.Out
	OPA   *sdk.OPA
	Ready opa.EmbeddedOPAReadyCH
}

type EmbeddedOPADI struct {
	fx.In
	AppCtx      *bootstrap.ApplicationContext
	Properties  opa.Properties
	Customizers []opa.ConfigCustomizer `group:"opa"`
}

func ProvideEmbeddedOPA(di EmbeddedOPADI) (EmbeddedOPAOut, error) {
	cfg, e := opa.LoadConfig(di.AppCtx, di.Properties, di.Customizers...)
	if e != nil {
		return EmbeddedOPAOut{}, fmt.Errorf("unable to load OPA Config: %v", e)
	}
	embedded, ready, e := opa.NewEmbeddedOPA(di.AppCtx,
		opa.WithConfig(cfg),
		opa.WithLogLevel(di.Properties.Logging.LogLevel),
		opa.WithInputCustomizers(opainput.DefaultInputCustomizers...),
	)
	if e != nil {
		return EmbeddedOPAOut{}, e
	}
	return EmbeddedOPAOut{
		OPA:   embedded,
		Ready: ready,
	}, nil
}

func InitializeEmbeddedOPA(lc fx.Lifecycle, opa *sdk.OPA, ready opa.EmbeddedOPAReadyCH) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				select {
				case <-ready:
					logger.WithContext(ctx).Infof("Embedded OPA is Ready")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			opa.Stop(ctx)
			return nil
		},
	})
}
