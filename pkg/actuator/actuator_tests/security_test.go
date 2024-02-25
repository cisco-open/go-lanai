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

package actuator_tests

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/actuator/actuator_tests/testdata"
	"github.com/cisco-open/go-lanai/pkg/actuator/env"
	"github.com/cisco-open/go-lanai/pkg/actuator/info"
	"github.com/cisco-open/go-lanai/pkg/actuator/loggers"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/actuatortest"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Tests
 *************************/

func ConfigureAccessByScopesWithPermissions(reg *actuator.Registrar, props actuator.ManagementProperties) {
	reg.MustRegister(actuator.NewAccessControlByScopes(props.Security, true))
}

func ConfigureAccessByScopesHardcoded(reg *actuator.Registrar, props actuator.ManagementProperties) {
	reg.MustRegister(actuator.NewAccessControlByScopes(props.Security, false, SpecialScopeAdmin))
}

func TestAccessByScopeUsingPermissions(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(env.Module, loggers.Module, info.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithProperties("management.security.permissions: " + SpecialScopeAdmin),
		apptest.WithFxOptions(
			fx.Invoke(ConfigureAccessByScopesWithPermissions),
		),
		test.GomegaSubTest(SubTestEnvWithAccess(mockedSecurityScopedAdmin()), "TestEnvWithAccess"),
		test.GomegaSubTest(SubTestEnvWithoutAccess(mockedSecurityAdmin()), "TestEnvWithoutAccess"),
		test.GomegaSubTest(SubTestLoggersWithAccess(mockedSecurityScopedAdmin()), "TestLoggersWithAccess"),
		test.GomegaSubTest(SubTestLoggersWithoutAccess(mockedSecurityAdmin()), "TestLoggersWithoutAccess"),
		test.GomegaSubTest(SubTestInfoEndpointV3(), "TestInfoEndpoint"),
	)
}

func TestAccessByHardcodedScope(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(env.Module, loggers.Module, info.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Invoke(ConfigureAccessByScopesHardcoded),
		),
		test.GomegaSubTest(SubTestEnvWithAccess(mockedSecurityScopedAdmin()), "TestEnvWithAccess"),
		test.GomegaSubTest(SubTestEnvWithoutAccess(mockedSecurityAdmin()), "TestEnvWithoutAccess"),
		test.GomegaSubTest(SubTestLoggersWithAccess(mockedSecurityScopedAdmin()), "TestLoggersWithAccess"),
		test.GomegaSubTest(SubTestLoggersWithoutAccess(mockedSecurityAdmin()), "TestLoggersWithoutAccess"),
		test.GomegaSubTest(SubTestInfoEndpointV3(), "TestInfoEndpoint"),
	)
}