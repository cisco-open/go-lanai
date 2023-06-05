package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/env"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/info"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/loggers"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
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