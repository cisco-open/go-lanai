package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"net/http"
	"sort"
)

/*****************************
	Abstraction
 *****************************/
type Candidate interface {
	Principal() interface{}
	Credentials() interface{}
	Details() interface{}
}

type Authenticator interface {
	// Authenticate function takes the Candidate and authenticate it.
	// if the Candidate type is not supported, return nil,nil
	// if the Candidate is rejected, non-nil error, and the returned Authentication is ignored
	Authenticate(context.Context, Candidate) (Authentication, error)
}

// AuthenticationSuccessHandler handles authentication success event
// The counterpart of this interface is AuthenticationErrorHandler
type AuthenticationSuccessHandler interface {
	HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to Authentication)
}

/*****************************
	Common Impl.
 *****************************/
// *CompositeAuthenticator implement Authenticator interface
type CompositeAuthenticator struct {
	authenticators []Authenticator
}

func NewAuthenticator(authenticators ...Authenticator) Authenticator {
	ret := &CompositeAuthenticator {}
	ret.authenticators = ret.processAuthenticators(authenticators)
	return ret
}

func (a *CompositeAuthenticator) Authenticate(ctx context.Context, candidate Candidate) (auth Authentication, err error) {
	for _,authenticator := range a.authenticators {
		auth, err = authenticator.Authenticate(ctx, candidate)
		if auth != nil || err != nil {
			return
		}
	}
	return nil, NewAuthenticatorNotAvailableError(fmt.Sprintf("unable to find authenticator for cadidate %T", candidate))
}

func (a *CompositeAuthenticator) Add(authenticator Authenticator) *CompositeAuthenticator {
	a.authenticators = a.processAuthenticators(append(a.authenticators, authenticator))
	sort.SliceStable(a.authenticators, func(i,j int) bool {
		return order.OrderedFirstCompare(a.authenticators[i], a.authenticators[j])
	})
	return a
}

func (a *CompositeAuthenticator) Merge(composite *CompositeAuthenticator) *CompositeAuthenticator {
	a.authenticators = a.processAuthenticators(append(a.authenticators, composite.authenticators...))
	return a
}

func (a *CompositeAuthenticator) processAuthenticators(authenticators []Authenticator) []Authenticator {
	// remove self
	authenticators = a.removeSelf(authenticators)
	sort.SliceStable(authenticators, func(i,j int) bool {
		return order.OrderedFirstCompare(authenticators[i], authenticators[j])
	})
	return authenticators
}

func (a *CompositeAuthenticator) removeSelf(authenticators []Authenticator) []Authenticator {
	count := 0
	for _, item := range authenticators {
		if ptr, ok := item.(*CompositeAuthenticator); !ok || ptr != a {
			// copy and increment index
			authenticators[count] = item
			count++
		}
	}
	// Prevent memory leak by erasing truncated values
	for j := count; j < len(authenticators); j++ {
		authenticators[j] = nil
	}
	return authenticators[:count]
}

// *CompositeAuthenticationSuccessHandler implement AuthenticationSuccessHandler interface
type CompositeAuthenticationSuccessHandler struct {
	handlers []AuthenticationSuccessHandler
}

func NewAuthenticationSuccessHandler(handlers ...AuthenticationSuccessHandler) *CompositeAuthenticationSuccessHandler {
	ret := &CompositeAuthenticationSuccessHandler {}
	ret.handlers = ret.processSuccessHandlers(handlers)
	return ret
}

func (h *CompositeAuthenticationSuccessHandler) HandleAuthenticationSuccess(
	c context.Context, r *http.Request, rw http.ResponseWriter, from, to Authentication) {

	for _,handler := range h.handlers {
		handler.HandleAuthenticationSuccess(c, r, rw, from, to)
	}
}

func (h *CompositeAuthenticationSuccessHandler) Add(handler AuthenticationSuccessHandler) *CompositeAuthenticationSuccessHandler {
	h.handlers = h.processSuccessHandlers(append(h.handlers, handler))
	return h
}

func (h *CompositeAuthenticationSuccessHandler) Merge(composite *CompositeAuthenticationSuccessHandler) *CompositeAuthenticationSuccessHandler {
	h.handlers = h.processSuccessHandlers(append(h.handlers, composite.handlers...))
	return h
}

func (h *CompositeAuthenticationSuccessHandler) processSuccessHandlers(handlers []AuthenticationSuccessHandler) []AuthenticationSuccessHandler {
	handlers = h.removeSelf(handlers)
	sort.SliceStable(handlers, func(i,j int) bool {
		return order.OrderedFirstCompare(handlers[i], handlers[j])
	})
	return handlers
}

func (h *CompositeAuthenticationSuccessHandler) removeSelf(items []AuthenticationSuccessHandler) []AuthenticationSuccessHandler {
	count := 0
	for _, item := range items {
		if ptr, ok := item.(*CompositeAuthenticationSuccessHandler); !ok || ptr != h {
			// copy and increment index
			items[count] = item
			count++
		}
	}
	// Prevent memory leak by erasing truncated values
	for j := count; j < len(items); j++ {
		items[j] = nil
	}
	return items[:count]
}