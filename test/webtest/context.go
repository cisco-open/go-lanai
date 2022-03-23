package webtest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"net/http"
	"net/http/httptest"
)

type TestServerOptions func(conf *TestServerConfig)

type TestServerConfig struct {
	Port        int
	ContextPath string
	LogLevel    log.LoggingLevel
}

type ExecResult struct {
	Response         *http.Response
	ResponseRecorder *httptest.ResponseRecorder
}
