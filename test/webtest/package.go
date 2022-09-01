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
	di := setupDI{}
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
	di := setupDI{}
	return test.WithOptions(
		apptest.WithModules(mockedWebModule),
		apptest.WithProperties(props...),
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

type setupDI struct {
	fx.In
	Registrar *web.Registrar
	Engine    *web.Engine
}

func toProperties(conf *TestServerConfig) []string {
	return []string{
		fmt.Sprintf("server.port: %d", conf.Port),
		fmt.Sprintf("server.context-path: %s", conf.ContextPath),
		fmt.Sprintf("server.logging.default-level: %s", conf.LogLevel.String()),
		"server.logging.enabled: true",
	}
}

func testSetupAddrExtractor(conf *TestServerConfig, di *setupDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		info := serverInfo{
			hostname:    "127.0.0.1",
			port:        di.Registrar.ServerPort(),
			contextPath: conf.ContextPath,
		}
		return newWebTestContext(ctx, &info, nil), nil
	}
}

func testSetupEngineExtractor(conf *TestServerConfig, di *setupDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		info := serverInfo{
			contextPath: conf.ContextPath,
		}
		return  newWebTestContext(ctx, &info, di.Engine), nil
	}
}
