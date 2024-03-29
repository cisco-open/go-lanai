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

package actuatortest

import (
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/actuator/env"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	healthep "github.com/cisco-open/go-lanai/pkg/actuator/health/endpoint"
	"github.com/cisco-open/go-lanai/pkg/actuator/info"
	actuatorinit "github.com/cisco-open/go-lanai/pkg/actuator/init"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/security/errorhandling"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"go.uber.org/fx"
)

// WithEndpoints is a convenient group of test options that enables actuator endpoints with following configuration
// 	- "info", "health" and "env" are initialized
// 	- The default "tokenauth" authentication is disabled. (sectest package can be used to test securities)
// 	- Uses the default properties and permission based access control. Custom access control can be registered
//
// Note 1: 	Choice of web testing environment are configured separately.
//			See webtest.WithMockedServer() and webtest.WithRealServer()
//
// Note 2:	Actuator endpoints usually requires correct properties to be fully functional,
//			make sure the test have all "management" properties configured correctly.
//
// Note 3:	Additional endpoints can be added by directly adding their Modules in test.
//
// Example:
// test.RunTest(context.Background(), t,
//		apptest.Bootstrap(),
//		webtest.WithMockedServer(),
//		sectest.WithMockedMiddleware(),
//		apptest.WithModules(
//			// additional endpoints
//			loggers.Module,
//		),
//		apptest.WithBootstrapConfigFS(testdata.MyTestBootstrapFS),
//		apptest.WithConfigFS(testdata.MyTestConfigFS),
//		apptest.WithProperties("more.properties: value"...),
//		test.GomegaSubTest(SubTestAdminEndpoints(), "MyTests"),
//	)
func WithEndpoints(opts ...ActuatorOptions) test.Options {
	opt := ActuatorOption{
		DisableAllEndpoints:          false,
		DisableDefaultAuthentication: true,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	testOpts := []test.Options{
		apptest.WithModules(actuatorinit.Module, actuator.Module, errorhandling.Module, access.Module),
	}

	if !opt.DisableAllEndpoints {
		testOpts = append(testOpts, apptest.WithModules(health.Module, healthep.Module, info.Module, env.Module))
	}

	if opt.DisableDefaultAuthentication {
		testOpts = append(testOpts, apptest.WithFxOptions(
			fx.Invoke(disableDefaultSecurity),
		))
	}
	return test.WithOptions(testOpts...)
}

// disableDefaultSecurity disable auto-configured "tokenauth" authentication
func disableDefaultSecurity(reg *actuator.Registrar) {
	reg.MustRegister(actuator.SecurityCustomizerFunc(func(ws security.WebSecurity) {/* this would override default */}))
}
