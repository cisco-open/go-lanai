package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	"time"
)

const (
	HighestReservedHookOrder  = -10000
	LowestReservedHookOrder   = 10000
	HookOrderTokenPassthrough = HighestReservedHookOrder + 10
	HookOrderRequestLogger    = LowestReservedHookOrder
	HookOrderResponseLogger   = HighestReservedHookOrder
)

const (
	logKey = "remote-http"
	ctxKeyStartTime = "start"
)

const (
	kb = 1024
	mb = kb * kb
	gb = mb * kb
)

// beforeHook implements BeforeHook, order.Ordered
type beforeHook struct {
	order int
	fn httptransport.RequestFunc
}

func Before(order int, fn httptransport.RequestFunc) BeforeHook {
	return &beforeHook{
		order: order,
		fn: fn,
	}
}

func (h beforeHook) Order() int {
	return h.order
}

func (h beforeHook) RequestFunc() httptransport.RequestFunc {
	return h.fn
}

// afterHook implements AfterHook, order.Ordered
type afterHook struct {
	order int
	fn httptransport.ClientResponseFunc
}

func After(order int, fn httptransport.ClientResponseFunc) AfterHook {
	return &afterHook{
		order: order,
		fn: fn,
	}
}

func (h afterHook) Order() int {
	return h.order
}

func (h afterHook) ResponseFunc() httptransport.ClientResponseFunc {
	return h.fn
}

// configurableBeforeHook implements ConfigurableBeforeHook
type configurableBeforeHook struct {
	beforeHook
	factory func(cfg *ClientConfig) beforeHook
}

func (h *configurableBeforeHook) WithConfig(cfg *ClientConfig) BeforeHook {
	return &configurableBeforeHook{
		beforeHook: h.factory(cfg),
		factory: h.factory,
	}
}

// configurableAfterHook implements ConfigurableAfterHook
type configurableAfterHook struct {
	afterHook
	factory func(cfg *ClientConfig) afterHook
}

func (h *configurableAfterHook) WithConfig(cfg *ClientConfig) AfterHook {
	return &configurableAfterHook{
		afterHook: h.factory(cfg),
		factory: h.factory,
	}
}

/*************************
	BeforeHook
 *************************/

func HookTokenPassthrough() BeforeHook {
	fn := func(ctx context.Context, request *http.Request) context.Context {
		authHeader := request.Header.Get(HeaderAuthorization)
		if authHeader != "" {
			return ctx
		}

		auth, ok := security.Get(ctx).(oauth2.Authentication)
		if !ok || !security.IsFullyAuthenticated(auth) || auth.AccessToken() == nil {
			return ctx
		}

		authHeader = fmt.Sprintf("Bearer %s", auth.AccessToken().Value())
		request.Header.Set(HeaderAuthorization, authHeader)
		return ctx
	}
	return Before(HookOrderTokenPassthrough, fn)
}

type httpLog struct {
	Method     string        `json:"method,omitempty"`
	URL        string        `json:"url,omitempty"`
	SC         int           `json:"statusCode,omitempty"`
	RespType   string        `json:"response_type,omitempty"`
	RespLength int           `json:"response_length,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

func hookRequestLogger(logger log.ContextualLogger, verbose bool) beforeHook {
	fn := func(ctx context.Context, request *http.Request) context.Context {
		kv := httpLog{
			Method: request.Method,
			URL: request.URL.RequestURI(),
		}
		logger.WithContext(ctx).
			WithKV(logKey, &kv).
			Debugf("[HTTP Request] %s %#v", request.Method, request.URL.RequestURI())
		return context.WithValue(ctx, ctxKeyStartTime, time.Now().UTC())
	}
	return beforeHook{
		order: HookOrderRequestLogger,
		fn: fn,
	}
}

func HookRequestLogger(logger log.ContextualLogger, verbose bool) BeforeHook {
	return &configurableBeforeHook{
		beforeHook: hookRequestLogger(logger, verbose),
		factory: func(cfg *ClientConfig) beforeHook {
			return hookRequestLogger(cfg.Logger, cfg.Verbose)
		},
	}
}

/*************************
	AfterHook
 *************************/

func hookResponseLogger(logger log.ContextualLogger, verbose bool) afterHook {
	fn := func(ctx context.Context, response *http.Response) context.Context {
		var duration time.Duration
		start, ok := ctx.Value(ctxKeyStartTime).(time.Time)
		if ok {
			duration = time.Since(start).Truncate(time.Microsecond)
		}
		kv := httpLog{
			Method:     response.Request.Method,
			URL:        response.Request.RequestURI,
			SC:         response.StatusCode,
			RespType:   response.Header.Get(HeaderContentType),
			RespLength: int(response.ContentLength),
			Duration:   duration,
		}
		logger.WithContext(ctx).
			WithKV(logKey, &kv).
			Debugf("[HTTP Response] %3d | %10v | %6s | %s ",
				response.StatusCode, duration, formatSize(kv.RespLength), kv.RespType)
		return ctx
	}
	return afterHook{
		order: HookOrderResponseLogger,
		fn: fn,
	}
}

func HookResponseLogger(logger log.ContextualLogger, verbose bool) AfterHook {
	return &configurableAfterHook{
		afterHook: hookResponseLogger(logger, verbose),
		factory: func(cfg *ClientConfig) afterHook {
			return hookResponseLogger(cfg.Logger, cfg.Verbose)
		},
	}
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