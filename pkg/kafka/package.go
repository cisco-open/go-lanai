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

package kafka

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/certs"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
	"reflect"
)

var logger = log.New("Kafka")

var Module = &bootstrap.Module{
	Precedence: bootstrap.KafkaPrecedence,
	Options: []fx.Option{
		fx.Provide(BindKafkaProperties, ProvideKafkaBinder),
		fx.Provide(tracingProvider()),
		fx.Invoke(initialize),
	},
}

const (
	FxGroup = "kafka"
)

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

type binderDI struct {
	fx.In
	AppContext           *bootstrap.ApplicationContext
	Properties           KafkaProperties
	ProducerInterceptors []ProducerMessageInterceptor  `group:"kafka"`
	ConsumerInterceptors []ConsumerDispatchInterceptor `group:"kafka"`
	HandlerInterceptors  []ConsumerHandlerInterceptor  `group:"kafka"`
	TLSCertsManager      certs.Manager                 `optional:"true"`
}

func ProvideKafkaBinder(di binderDI) Binder {
	return NewBinder(di.AppContext, func(opt *BinderOption) {
		*opt = BinderOption{
			ApplicationConfig:    di.AppContext.Config(),
			Properties:           di.Properties,
			ProducerInterceptors: append(opt.ProducerInterceptors, filterZeroValues(di.ProducerInterceptors)...),
			ConsumerInterceptors: append(opt.ConsumerInterceptors, filterZeroValues(di.ConsumerInterceptors)...),
			HandlerInterceptors:  append(opt.HandlerInterceptors, filterZeroValues(di.HandlerInterceptors)...),
			TLSCertsManager:      di.TLSCertsManager,
		}
	})
}

type initDI struct {
	fx.In
	AppCtx          *bootstrap.ApplicationContext
	Lifecycle       fx.Lifecycle
	Properties      KafkaProperties
	Binder          Binder
	HealthRegistrar health.Registrar `optional:"true"`
}

func initialize(di initDI) {
	// register lifecycle functions
	di.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			//nolint:contextcheck // intentional, given context is cancelled after bootstrap, AppCtx is cancelled when app close
			return di.Binder.(BinderLifecycle).Start(di.AppCtx)
		},
		OnStop: func(ctx context.Context) error {
			return di.Binder.(BinderLifecycle).Shutdown(ctx)
		},
	})

	// register health endpoints if applicable
	if di.HealthRegistrar == nil {
		return
	}

	di.HealthRegistrar.MustRegister(NewHealthIndicator(di.Binder))
}

func filterZeroValues[T any](values []T) []T {
	filtered := make([]T, 0, len(values))
	for i := range values {
		rv := reflect.ValueOf(values[i])
		if rv.IsValid() && !rv.IsZero() {
			filtered = append(filtered, values[i])
		}
	}
	return filtered
}
