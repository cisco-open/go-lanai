package web

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
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

type LoggingCustomizer struct {
	
}

func NewLoggingCustomizer() *LoggingCustomizer {
	return &LoggingCustomizer{}
}

func (c LoggingCustomizer) Customize(r *Registrar) error {
	// override gin debug
	gin.DefaultWriter = log.NewWriterAdapter(logger, log.LevelDebug)
	gin.DefaultErrorWriter = log.NewWriterAdapter(logger, log.LevelDebug)

	// setup logger middleware
	mw := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: logFormatter{logger: logger}.intercept,
		Output:    ioutil.Discard, // our logFormatter calls logger directly
	})
	r.AddGlobalMiddlewares(mw)
	return nil
}

type logFormatter struct {
	logger log.ContextualLogger
}

// intercept uses logger directly and return empty string.
// doing so would allow us to set key-value pairs
func (f logFormatter) intercept(params gin.LogFormatterParams) (empty string) {
	var statusColor, methodColor, resetColor string
	methodLen := 7
	if log.IsTerminal(f.logger) {
		statusColor = params.StatusCodeColor()
		methodColor = params.MethodColor()
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
	http := map[string]interface{} {
		LogKeyHttpStatus:   params.StatusCode,
		LogKeyHttpMethod:   params.Method,
		LogKeyHttpClientIP: params.ClientIP,
		LogKeyHttpPath:     params.Path,
		LogKeyHttpBodySize: params.BodySize,
		LogKeyHttpErrorMsg: params.ErrorMessage,
	}

	// do log
	f.logger.WithContext(ctx).Debug(msg, LogKeyHttp, http)
	return
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



