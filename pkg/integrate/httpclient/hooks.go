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

func hookRequestLogger(logger log.ContextualLogger, logging *LoggingConfig) beforeHook {
	fn := func(ctx context.Context, request *http.Request) context.Context {
		now := time.Now().UTC()
		logRequest(ctx, request, logger, logging)
		return context.WithValue(ctx, ctxKeyStartTime, now)
	}
	return beforeHook{
		order: HookOrderRequestLogger,
		fn: fn,
	}
}

func HookRequestLogger(logger log.ContextualLogger, logging *LoggingConfig) BeforeHook {
	return &configurableBeforeHook{
		beforeHook: hookRequestLogger(logger, logging),
		factory: func(cfg *ClientConfig) beforeHook {
			return hookRequestLogger(cfg.Logger, &cfg.Logging)
		},
	}
}

/*************************
	AfterHook
 *************************/

func hookResponseLogger(logger log.ContextualLogger, logging *LoggingConfig) afterHook {
	fn := func(ctx context.Context, response *http.Response) context.Context {
		logResponse(ctx, response, logger, logging)
		return ctx
	}
	return afterHook{
		order: HookOrderResponseLogger,
		fn: fn,
	}
}

func HookResponseLogger(logger log.ContextualLogger, logging *LoggingConfig) AfterHook {
	return &configurableAfterHook{
		afterHook: hookResponseLogger(logger, logging),
		factory: func(cfg *ClientConfig) afterHook {
			return hookResponseLogger(cfg.Logger, &cfg.Logging)
		},
	}
}
