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

package web

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/pkg/utils/matcher"
    "github.com/gin-gonic/gin"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"
)

var logger = log.New("Web")

const (
	LogKeyHttp = "http"
	LogKeyHttpStatus = "status"
	LogKeyHttpMethod = "method"
	LogKeyHttpClientIP = "clientIP"
	LogKeyHttpPath = "path"
	LogKeyHttpErrorMsg = "error"
	LogKeyHttpBodySize = "bodySize"
)


// LoggingCustomizer implements Customizer and PostInitCustomizer
type LoggingCustomizer struct {
	enabled    bool
	defaultLvl log.LoggingLevel
	levels     map[RequestMatcher]log.LoggingLevel
}

func NewLoggingCustomizer(props ServerProperties) *LoggingCustomizer {
	return &LoggingCustomizer{
		enabled:    props.Logging.Enabled,
		defaultLvl: props.Logging.DefaultLevel,
		levels:     initLevelMap(&props),
	}
}

// NewSimpleGinLogFormatter is a convenient function that returns a simple gin.LogFormatter without request filtering
// Normally, LoggingCustomizer configures more complicated gin logging schema automatically.
// This function is provided purely for integrating with 3rd-party libraries that configures gin.Engine separately.
// e.g. KrakenD in API Gateway Service
func NewSimpleGinLogFormatter(logger log.ContextualLogger, defaultLevel log.LoggingLevel, levels map[RequestMatcher]log.LoggingLevel) gin.LogFormatter {
	formatter := logFormatter{
		logger:     logger,
		defaultLvl: defaultLevel,
		levels:     levels,
	}
	return formatter.intercept
}

func initLevelMap(props *ServerProperties) map[RequestMatcher]log.LoggingLevel {
	levels := map[RequestMatcher]log.LoggingLevel{}
	for _, v := range props.Logging.Levels {
		pattern := props.ContextPath + v.Pattern
		var m RequestMatcher
		if v.Method == "" || v.Method == "*" {
			m = withLoggingRequestPattern(pattern)
		} else {
			split := strings.Split(v.Method, " ")
			methods := make([]string, 0, len(split))
			for _, s := range split {
				s := strings.TrimSpace(s)
				if s != "" {
					methods = append(methods, s)
				}
			}
			m = withLoggingRequestPattern(pattern, methods...)
		}
		levels[m] = v.Level
	}
	return levels
}

func (c LoggingCustomizer) Customize(ctx context.Context, r *Registrar) error {
	// override gin debug
	gin.DefaultWriter = log.NewWriterAdapter(logger.WithContext(ctx), log.LevelDebug)
	gin.DefaultErrorWriter = log.NewWriterAdapter(logger.WithContext(ctx), log.LevelWarn)

	if !c.enabled {
		return nil
	}

	// setup logger middleware
	formatter := logFormatter{
		defaultLvl: c.defaultLvl,
		logger:     logger,
		levels:     c.levels,
	}
	mw := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: formatter.intercept,
		Output:    io.Discard, // our logFormatter calls logger directly
	})
	if e := r.AddGlobalMiddlewares(mw); e != nil {
		panic(e)
	}
	return nil
}

func (c LoggingCustomizer) PostInit(_ context.Context, _ *Registrar) error {
	// release initializing context
	gin.DefaultWriter = log.NewWriterAdapter(logger, log.LevelDebug)
	gin.DefaultErrorWriter = log.NewWriterAdapter(logger, log.LevelDebug)
	return nil
}

type logFormatter struct {
	logger     log.ContextualLogger
	defaultLvl log.LoggingLevel
	levels     map[RequestMatcher]log.LoggingLevel
}

// intercept uses logger directly and return empty string.
// doing so would allow us to set key-value pairs
func (f logFormatter) intercept(params gin.LogFormatterParams) (empty string) {

	logLevel := f.logLevel(params.Request)
	if logLevel == log.LevelOff {
		return
	}

	var statusColor, methodColor, resetColor string
	methodLen := 7
	if log.IsTerminal(f.logger) {
		statusColor = fixColor(params.StatusCodeColor())
		methodColor = fixColor(params.MethodColor())
		resetColor = params.ResetColor()
		methodLen = methodLen + len(methodColor) + len(resetColor)
	}

	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Minute)
	}

	params.ErrorMessage = strings.Trim(params.ErrorMessage, "\n")

	// prepare message
	method := fmt.Sprintf("%-" + strconv.Itoa(methodLen) + "s", methodColor + " "+ params.Method + " " + resetColor)
	msg := fmt.Sprintf("[HTTP] %s %3d %s | %10v | %8s | %s %#v %s",
		statusColor, params.StatusCode, resetColor,
		params.Latency.Truncate(time.Microsecond),
		formatSize(params.BodySize),
		method,
		params.Path,
		params.ErrorMessage)

	// prepare kv
	ctx := utils.MakeMutableContext(params.Request.Context())

	for k, v := range params.Keys {
		ctx.Set(k, v)
	}
	httpEntry := map[string]interface{} {
		LogKeyHttpStatus:   params.StatusCode,
		LogKeyHttpMethod:   params.Method,
		LogKeyHttpClientIP: params.ClientIP,
		LogKeyHttpPath:     params.Path,
		LogKeyHttpBodySize: params.BodySize,
		LogKeyHttpErrorMsg: params.ErrorMessage,
	}

	// do log
	f.logger.WithContext(ctx).WithLevel(logLevel).WithKV(LogKeyHttp, httpEntry).Printf(msg)
	return
}

func (f logFormatter) logLevel(r *http.Request) log.LoggingLevel {
	for k, v := range f.levels {
		if match, e := k.Matches(r); e == nil && match {
			return v
		}
	}
	return f.defaultLvl
}

const (
	kb = 1024
	mb = kb * kb
	gb = mb * kb
)

func formatSize(n int) string {
	switch {
	case n < kb:
		return fmt.Sprintf("%dB", n)
	case n < mb:
		return fmt.Sprintf("%.2fKB", float64(n) / kb)
	case n < gb:
		return fmt.Sprintf("%.2fMB", float64(n) / mb)
	default:
		return fmt.Sprintf("%.2fGB", float64(n) / gb)
	}
}

func fixColor(color string) string {
	if strings.Contains(color, "43") {
		color = strings.Replace(color, "90;", "97;", 1)
	}
	return color
}

// loggingRequestMatcher implement RequestMatcher
// loggingRequestMatcher is exclusively used by logFormatter.
// The purpose of this matcher is
// 1. break cyclic package dependency
// 2. provide simple and faster matching
type loggingRequestMatcher struct {
	pathMatcher matcher.StringMatcher
	methods     []string
	description string
}

func withLoggingRequestPattern(pattern string, methods...string) *loggingRequestMatcher {
	return &loggingRequestMatcher{
		pathMatcher: matcher.WithPathPattern(pattern),
		methods: methods,
		description: fmt.Sprintf("request matches %v %s", methods, pattern),
	}
}

func (m *loggingRequestMatcher) RequestMatches(_ context.Context, r *http.Request) (bool, error) {
	path := r.URL.Path
	match, e := m.pathMatcher.Matches(path)
	if e != nil || !match {
		return false, e
	}

	if len(m.methods) == 0 {
		return true, nil
	}

	for _, method := range m.methods {
		if r.Method == method {
			return true, nil
		}
	}
	return false, nil
}

func (m *loggingRequestMatcher) Matches(i interface{}) (bool, error) {
	value, ok := i.(*http.Request)
	if !ok {
		return false, fmt.Errorf("unsupported type %T", i)
	}
	return m.RequestMatches(context.TODO(), value)
}

func (m *loggingRequestMatcher) MatchesWithContext(c context.Context, i interface{}) (bool, error) {
	value, ok := i.(*http.Request)
	if !ok {
		return false, fmt.Errorf("unsupported type %T", i)
	}
	return m.RequestMatches(c, value)
}

func (m *loggingRequestMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *loggingRequestMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m *loggingRequestMatcher) String() string {
	return m.description
}
