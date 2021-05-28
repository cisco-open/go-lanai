package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type requestLog struct {
	Method  string            `json:"method,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type responseLog struct {
	requestLog
	SC         int           `json:"statusCode,omitempty"`
	RespLength int           `json:"length,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

func logRequest(ctx context.Context, r *http.Request, logger log.ContextualLogger, logging *LoggingConfig) {
	if logging.DetailsLevel < LogDetailsLevelMinimum {
		return
	}

	kv, msg := constructRequestLog(r, logging)
	_ = logger.WithContext(ctx).WithKV(logKey, &kv).WithLevel(logging.Level).Log(log.LogKeyMessage, msg)
}

func logResponse(ctx context.Context, resp *http.Response, logger log.ContextualLogger, logging *LoggingConfig) {
	if logging.DetailsLevel < LogDetailsLevelMinimum {
		return
	}

	kv, msg := constructResponseLog(ctx, resp, logging)
	_ = logger.WithContext(ctx).WithKV(logKey, &kv).WithLevel(logging.Level).Log(log.LogKeyMessage, msg)
}

func constructRequestLog(r *http.Request, logging *LoggingConfig) (*requestLog, string) {
	msg := []string{fmt.Sprintf("[HTTP Request] %s %#v", r.Method, r.URL.RequestURI())}
	kv := requestLog{
		Method: r.Method,
		URL: r.URL.RequestURI(),
	}

	if logging.DetailsLevel >= LogDetailsLevelHeaders {
		var text string
		kv.Headers, text = sanitizedHeaders(r.Header, logging.SanitizeHeaders, logging.ExcludeHeaders)
		msg = append(msg, text)
	}

	if logging.DetailsLevel >= LogDetailsLevelFull {
		kv.Body = "Request logging is currently unsupported"
		msg = append(msg, kv.Body)
	}
	return &kv, strings.Join(msg, " | ")
}

func constructResponseLog(ctx context.Context, resp *http.Response, logging *LoggingConfig) (*responseLog, string) {
	var duration time.Duration
	start, ok := ctx.Value(ctxKeyStartTime).(time.Time)
	if ok {
		duration = time.Since(start).Truncate(time.Microsecond)
	}

	kv := responseLog{
		requestLog: requestLog{
			Method: resp.Request.Method,
			URL: resp.Request.URL.RequestURI(),
		},
		SC:         resp.StatusCode,
		RespLength: int(resp.ContentLength),
		Duration:   duration,
	}
	msg := []string{fmt.Sprintf("[HTTP Response] %3d | %10v | %6s",
		resp.StatusCode, duration, formatSize(kv.RespLength))}

	if logging.DetailsLevel >= LogDetailsLevelHeaders {
		var text string
		kv.Headers, text = sanitizedHeaders(resp.Header, logging.SanitizeHeaders, logging.ExcludeHeaders)
		msg = append(msg, text)
	}

	if logging.DetailsLevel >= LogDetailsLevelFull {
		kv.Body = "Response logging is currently unsupported"
		msg = append(msg, kv.Body)
	}
	return &kv, strings.Join(msg, " | ")
}

func sanitizedHeaders(headers http.Header, sanitize utils.StringSet, exclude utils.StringSet) (map[string]string, string) {
	kv := map[string]string{}
	msg := make([]string, 0)
	for k, v := range headers {
		if exclude.Has(k) {
			continue
		}
		value := "******"
		if !sanitize.Has(k) {
			value = strings.Join(v, " ")
		}
		kv[k] = value
		msg = append(msg, k + `[` + value + `]`)
	}
	return kv, strings.Join(msg, ", ")
}

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