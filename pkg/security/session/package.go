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

package session

import (
    "encoding/gob"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/passwd"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/template"
    "go.uber.org/fx"
    "path"
    "time"
)

var logger = log.New("SEC.Session")

var Module = &bootstrap.Module{
	Name:       "session",
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(security.BindSessionProperties),
		fx.Provide(provideSessionStore),
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)

	GobRegister()
	security.GobRegister()
	passwd.GobRegister()
	template.RegisterGlobalModelValuer(template.ModelKeySession, template.ContextModelValuer(Get))
}

func GobRegister() {
	gob.Register([]interface{}{})
}

type storeDI struct {
	fx.In
	AppContext    *bootstrap.ApplicationContext
	SecRegistrar  security.Registrar `optional:"true"`
	SessionProps  security.SessionProperties
	ServerProps   web.ServerProperties         `optional:"true"`
	ClientFactory redis.ClientFactory          `optional:"true"`
	SettingReader security.GlobalSettingReader `optional:"true"`
}

func provideSessionStore(di storeDI) Store {
	if di.SecRegistrar == nil || di.ClientFactory == nil {
		return nil
	}
	redisClient, e := di.ClientFactory.New(di.AppContext, func(opt *redis.ClientOption) {
		opt.DbIndex = di.SessionProps.DbIndex
	})
	if e != nil {
		panic(e)
	}

	return NewRedisStore(redisClient, func(opt *StoreOption) {
		opt.SettingReader = di.SettingReader

		opt.Options.Path = path.Clean(di.SessionProps.Cookie.Path)
		opt.Options.Domain = di.SessionProps.Cookie.Domain
		opt.Options.MaxAge = di.SessionProps.Cookie.MaxAge
		opt.Options.Secure = di.SessionProps.Cookie.Secure
		opt.Options.HttpOnly = di.SessionProps.Cookie.HttpOnly
		opt.Options.SameSite = di.SessionProps.Cookie.SameSite()
		opt.Options.IdleTimeout = time.Duration(di.SessionProps.IdleTimeout)
		opt.Options.AbsoluteTimeout = time.Duration(di.SessionProps.AbsoluteTimeout)
	})
}

type initDI struct {
	fx.In
	AppContext            *bootstrap.ApplicationContext
	SecRegistrar          security.Registrar `optional:"true"`
	SessionProps          security.SessionProperties
	SessionStore          Store          `optional:"true"`
	SessionSettingService SettingService `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil && di.SessionStore != nil {
		configurer := newSessionConfigurer(di.SessionProps, di.SessionStore)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}
