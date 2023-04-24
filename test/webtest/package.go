package webtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"fmt"
	"go.uber.org/fx"
	"testing"
)

const (
	DefaultContextPath = "/test"
)

// WithRealServer start a real web server at random port with context-path as DefaultContextPath.
// NewRequest(), Exec() and MustExec() can be used to create/send request and verifying result
// By default, the server doesn't allow CORS and have no security configured.
// Actual server port can be retrieved via CurrentPort()
// When using this mode, *web.Registrar became available to inject
func WithRealServer(opts ...TestServerOptions) test.Options {
	conf := TestServerConfig{
		ContextPath: DefaultContextPath,
		LogLevel:    log.LevelInfo,
	}
	for _, fn := range opts {
		fn(&conf)
	}
	props := toProperties(&conf)
	di := realSvrDI{}
	return test.WithOptions(
		apptest.WithModules(webinit.Module),
		apptest.WithProperties(props...),
		apptest.WithDI(&di),
		test.SubTestSetup(testSetupAddrExtractor(&conf, &di)),
	)
}

// WithMockedServer initialize web package without starting an actual web server.
// NewRequest(), Exec() and MustExec() can be used to create/send request and verifying result without creating an actual http connection.
// By default, the server doesn't allow CORS and have no security configured.
// When using this mode, *web.Registrar became available to inject
// Note: In this mode, httptest package is used internally and http.Handler (*web.Engine in our case) is invoked directly
func WithMockedServer(opts ...TestServerOptions) test.Options {
	conf := TestServerConfig{
		ContextPath: DefaultContextPath,
		LogLevel:    log.LevelInfo,
	}
	for _, fn := range opts {
		fn(&conf)
	}
	props := toProperties(&conf)
	di := mockedSvrDI{}
	return test.WithOptions(
		apptest.WithModules(mockedWebModule),
		apptest.WithProperties(props...),
		apptest.WithDI(&di),
		test.SubTestSetup(testSetupEngineExtractor(&conf, &di)),
	)
}

// WithUtilities DOES NOT initialize web package, it only provide properties and setup utilities (e.g. MustExec)
// Important: this mode is mostly for go-lanai internal tests. DO NOT use it in microservices
//
// NewRequest(), Exec() and MustExec() can be used to create/send request and verifying result without creating an actual http connection.
// Note: In this mode, httptest package is used internally and http.Handler (*web.Engine in our case) is invoked directly
func WithUtilities(opts ...TestServerOptions) test.Options {
	conf := TestServerConfig{
		ContextPath: DefaultContextPath,
		LogLevel:    log.LevelInfo,
	}
	for _, fn := range opts {
		fn(&conf)
	}
	props := toProperties(&conf)
	di := mockedSvrDI{}
	return test.WithOptions(
		apptest.WithProperties(props...),
		apptest.WithFxOptions(
			fx.Provide(
				web.BindServerProperties,
			),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(testSetupEngineExtractor(&conf, &di)),
	)
}

// UsePort returns a TestServerOptions that use given port.
// Note: using fixed port might cause issues when run in CI/CD
func UsePort(port int) TestServerOptions {
	return func(conf *TestServerConfig) {
		conf.Port = port
	}
}

// UseContextPath returns a TestServerOptions that overwrite the context-path of the test server
func UseContextPath(contextPath string) TestServerOptions {
	return func(conf *TestServerConfig) {
		conf.ContextPath = contextPath
	}
}

// UseLogLevel returns a TestServerOptions that overwrite the default log level of the test server
func UseLogLevel(lvl log.LoggingLevel) TestServerOptions {
	return func(conf *TestServerConfig) {
		conf.LogLevel = lvl
	}
}

// AddDefaultRequestOptions returns a TestServerOptions that add default RequestOptions on every request
// created via NewRequest.
func AddDefaultRequestOptions(opts...RequestOptions) TestServerOptions {
	return func(conf *TestServerConfig) {
		if conf.RequestOptions == nil {
			conf.RequestOptions = opts
		} else {
			conf.RequestOptions = append(conf.RequestOptions, opts...)
		}
	}
}

type realSvrDI struct {
	fx.In
	Registrar *web.Registrar
	Engine    *web.Engine
}

type mockedSvrDI struct {
	fx.In
	Engine    *web.Engine `optional:"true"`
}

func toProperties(conf *TestServerConfig) []string {
	return []string{
		fmt.Sprintf("server.port: %d", conf.Port),
		fmt.Sprintf("server.context-path: %s", conf.ContextPath),
		fmt.Sprintf("server.logging.default-level: %s", conf.LogLevel.String()),
		"server.logging.enabled: true",
	}
}

func testSetupAddrExtractor(conf *TestServerConfig, di *realSvrDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		info := serverInfo{
			hostname:    "127.0.0.1",
			port:        di.Registrar.ServerPort(),
			contextPath: conf.ContextPath,
		}
		return newWebTestContext(ctx, conf, &info, nil), nil
	}
}

func testSetupEngineExtractor(conf *TestServerConfig, di *mockedSvrDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		info := serverInfo{
			contextPath: conf.ContextPath,
		}
		return  newWebTestContext(ctx, conf, &info, di.Engine), nil
	}
}
