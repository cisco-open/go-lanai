package webtest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

type TestServerOptions func(conf *TestServerConfig)

type TestServerConfig struct {
	Port        int
	ContextPath string
	LogLevel    log.LoggingLevel
}

type DI struct {
	fx.In
	WebRegistrar *web.Registrar
}