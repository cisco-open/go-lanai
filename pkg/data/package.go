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

package data

import (
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"time"
)

var logger = log.New("Data")

var Module = &bootstrap.Module{
	Name:       "DB",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(
			BindDataProperties,
			BindDatabaseProperties,
			provideGorm,
			gormErrTranslatorProvider(),
		),
		fx.Invoke(registerHealth),
	},
}

/**************************
	Provider
***************************/

type gormInitDI struct {
	fx.In
	Dialector   gorm.Dialector
	Properties  DataProperties
	Configurers []GormConfigurer   `group:"gorm_config"`
	Translators []ErrorTranslator  `group:"gorm_config"`
	Tracer      opentracing.Tracer `optional:"true"`
}

func provideGorm(di gormInitDI) *gorm.DB {
	return NewGorm(func(cfg *GormConfig) {
		cfg.Dialector = di.Dialector
		cfg.LogLevel = di.Properties.Logging.Level
		cfg.Configurers = append(cfg.Configurers, NewGormErrorHandlingConfigurer(di.Translators...))
		if di.Tracer != nil {
			cfg.Configurers = append(cfg.Configurers, NewGormTracingConfigurer(di.Tracer))
		}
		cfg.Configurers = append(cfg.Configurers, di.Configurers...)
		if di.Properties.Logging.SlowThreshold > 0 {
			cfg.LogSlowQueryThreshold = time.Duration(di.Properties.Logging.SlowThreshold)
		}
	})
}

func gormErrTranslatorProvider() fx.Annotated {
	return fx.Annotated{
		Group: GormConfigurerGroup,
		Target: func() ErrorTranslator {
			return NewGormErrorTranslator()
		},
	}
}

/**************************
	Initialize
***************************/

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	GormDB          *gorm.DB         `optional:"true"`
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil || di.GormDB == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&DbHealthIndicator{
		db: di.GormDB,
	})
}
