package webtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"fmt"
	"testing"
)

const (
	DefaultContextPath = "/test"
)

// WithRealTestServer start a real web server at random port with context-path as DefaultContextPath.
// By default, the server doesn't allow CORS and have no security configured.
// When using this mode, *web.Registrar became available to inject
func WithRealTestServer(opts ...TestServerOptions) test.Options {
	conf := TestServerConfig{
		ContextPath: DefaultContextPath,
		LogLevel: log.LevelInfo,
	}
	for _, fn := range opts {
		fn(&conf)
	}
	props := toProperties(&conf)
	di := DI{}
	return test.WithOptions(
		apptest.WithModules(webinit.Module),
		apptest.WithProperties(props...),
		apptest.WithDI(&di),
		test.SubTestSetup(testSetupAddrExtractor(&conf, &di)),
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

func toProperties(conf *TestServerConfig) []string {
	return []string{
		fmt.Sprintf("server.port: %d", conf.Port),
		fmt.Sprintf("server.context-path: %s", conf.ContextPath),
		fmt.Sprintf("server.logging.default-level: %s", conf.LogLevel.String()),
		"server.logging.enabled: true",
	}
}

func testSetupAddrExtractor(conf *TestServerConfig, di *DI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		v := addr{
			hostname:    "127.0.0.1",
			port:        di.WebRegistrar.ServerPort(),
			contextPath: conf.ContextPath,
		}
		return contestWithAddr(ctx, v), nil
	}
}