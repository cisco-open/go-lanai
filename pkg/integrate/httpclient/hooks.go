package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

const (
	HighestReservedHookOrder = -10000
	LowestReservedHookOrder = 10000
	HookOrderTokenPassthrough = HighestReservedHookOrder + 10
)

type orderedBeforeHook struct {
	order int
	fn httptransport.RequestFunc
}

func Before(order int, fn httptransport.RequestFunc) BeforeHook {
	return &orderedBeforeHook{
		order: order,
		fn: fn,
	}
}

func (h orderedBeforeHook) Order() int {
	return h.order
}

func (h orderedBeforeHook) RequestFunc() httptransport.RequestFunc {
	return h.fn
}

type orderedAfterHook struct {
	order int
	fn httptransport.ClientResponseFunc
}

func After(order int, fn httptransport.ClientResponseFunc) AfterHook {
	return &orderedAfterHook{
		order: order,
		fn: fn,
	}
}

func (h orderedAfterHook) Order() int {
	return h.order
}

func (h orderedAfterHook) ResponseFunc() httptransport.ClientResponseFunc {
	return h.fn
}

/*************************
	BeforeHook
 *************************/

func TokenPassthrough() BeforeHook {
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
