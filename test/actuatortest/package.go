package actuatortest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/env"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	healthep "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health/endpoint"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/info"
	actuatorinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
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
// Note 3:	Additional endpoints can be added by directly adding their Components in test.
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
