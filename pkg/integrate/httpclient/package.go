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

package httpclient

import (
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("HttpClient")

var Module = &bootstrap.Module{
	Name:       "http-client",
	Precedence: bootstrap.HttpClientPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(bindHttpClientProperties),
		fx.Provide(provideHttpClient),
	},
}

func Use() {
	bootstrap.Register(Module)
}

// FxClientCustomizers takes providers of ClientCustomizer and wrap them with FxGroup
func FxClientCustomizers(providers ...interface{}) []fx.Annotated {
	annotated := make([]fx.Annotated, len(providers))
	for i, t := range providers {
		annotated[i] = fx.Annotated{
			Group:  FxGroup,
			Target: t,
		}
	}
	return annotated
}

type clientDI struct {
	fx.In
	Properties  HttpClientProperties
	DiscClient  discovery.Client   `optional:"true"`
	Customizers []ClientCustomizer `group:"http-client"`
}

func provideHttpClient(di clientDI) Client {
	options := []ClientOptions{func(opt *ClientOption) {
		opt.SDClient = di.DiscClient
		opt.MaxRetries = di.Properties.MaxRetries
		opt.Timeout = time.Duration(di.Properties.Timeout)
		opt.Logging.Level = di.Properties.Logger.Level
		opt.Logging.DetailsLevel = di.Properties.Logger.DetailsLevel
		opt.Logging.SanitizeHeaders = utils.NewStringSet(di.Properties.Logger.SanitizeHeaders...)
		opt.Logging.ExcludeHeaders = utils.NewStringSet(di.Properties.Logger.ExcludeHeaders...)
	}}
	for _, customizer := range di.Customizers {
		options = append(options, customizer.Customize)
	}

	return NewClient(options...)
}
