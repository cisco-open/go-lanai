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
	Authenticate(Candidate) (Authentication, error)
}

// AuthenticationSuccessHandler handles authentication success event
// The counterpart of this interface is AuthenticationErrorHandler
type AuthenticationSuccessHandler interface {
	HandleAuthenticationSuccess(context.Context, *http.Request, http.ResponseWriter, Authentication)
}

/*****************************
	Common Impl.
 *****************************/
// *CompositeAuthenticator implement Authenticator interface
type CompositeAuthenticator struct {
	authenticators []Authenticator
}

func NewAuthenticator(authenticators ...Authenticator) Authenticator {
	return &CompositeAuthenticator {
		authenticators: sortAuthenticators(authenticators),
	}
}

func (a *CompositeAuthenticator) Authenticate(candidate Candidate) (auth Authentication, err error) {
	for _,authenticator := range a.authenticators {
		auth, err = authenticator.Authenticate(candidate)
		if auth != nil || err != nil {
			return
		}
	}
	return nil, NewAuthenticatorNotAvailableError(fmt.Sprintf("unable to find authenticator for cadidate %T", candidate))
}

func (a *CompositeAuthenticator) Add(authenticator Authenticator) *CompositeAuthenticator {
	a.authenticators = sortAuthenticators(append(a.authenticators, authenticator))
	sort.SliceStable(a.authenticators, func(i,j int) bool {
		return order.OrderedFirstCompare(a.authenticators[i], a.authenticators[j])
	})
	return a
}

func (a *CompositeAuthenticator) Merge(composite *CompositeAuthenticator) *CompositeAuthenticator {
	a.authenticators = sortAuthenticators(append(a.authenticators, composite.authenticators...))
	return a
}

func sortAuthenticators(authenticators []Authenticator) []Authenticator {
	sort.SliceStable(authenticators, func(i,j int) bool {
		return order.OrderedFirstCompare(authenticators[i], authenticators[j])
	})
	return authenticators
}

// *CompositeAuthenticationSuccessHandler implement AuthenticationSuccessHandler interface
type CompositeAuthenticationSuccessHandler struct {
	handlers []AuthenticationSuccessHandler
}

func NewAuthenticationSuccessHandler(handlers ...AuthenticationSuccessHandler) *CompositeAuthenticationSuccessHandler {
	return &CompositeAuthenticationSuccessHandler {
		handlers: sortSuccessHandlers(handlers),
	}
}

func (h *CompositeAuthenticationSuccessHandler) HandleAuthenticationSuccess(
	c context.Context, r *http.Request, rw http.ResponseWriter, auth Authentication) {

	for _,handler := range h.handlers {
		handler.HandleAuthenticationSuccess(c, r, rw, auth)
	}
}

func (h *CompositeAuthenticationSuccessHandler) Add(handler AuthenticationSuccessHandler) *CompositeAuthenticationSuccessHandler {
	h.handlers = sortSuccessHandlers(append(h.handlers, handler))
	return h
}

func (h *CompositeAuthenticationSuccessHandler) Merge(composite *CompositeAuthenticationSuccessHandler) *CompositeAuthenticationSuccessHandler {
	h.handlers = sortSuccessHandlers(append(h.handlers, composite.handlers...))
	return h
}

func sortSuccessHandlers(handlers []AuthenticationSuccessHandler) []AuthenticationSuccessHandler {
	sort.SliceStable(handlers, func(i,j int) bool {
		return order.OrderedFirstCompare(handlers[i], handlers[j])
	})
	return handlers
}