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

package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"net/http"
	"sort"
	"sync"
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

type AuthenticatorBuilder interface {
	Build(context.Context) (Authenticator, error)
}

// AuthenticationSuccessHandler handles authentication success event
// The counterpart of this interface is AuthenticationErrorHandler
type AuthenticationSuccessHandler interface {
	HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to Authentication)
}

/*****************************
	Common Impl.
 *****************************/

// CompositeAuthenticator implement Authenticator interface
type CompositeAuthenticator struct {
	init sync.Once
	authenticators []Authenticator
	flattened      []Authenticator
}

func NewAuthenticator(authenticators ...Authenticator) Authenticator {
	ret := &CompositeAuthenticator{}
	ret.authenticators = ret.processAuthenticators(authenticators)
	return ret
}

func (a *CompositeAuthenticator) Authenticate(ctx context.Context, candidate Candidate) (auth Authentication, err error) {
	a.init.Do(func() {a.flattened = a.Authenticators()})
	for _, authenticator := range a.flattened {
		auth, err = authenticator.Authenticate(ctx, candidate)
		if auth != nil || err != nil {
			return
		}
	}
	return nil, NewAuthenticatorNotAvailableError(fmt.Sprintf("unable to find authenticator for cadidate %T", candidate))
}

// Authenticators returns list of authenticators, any nested composite handlers are flattened
func (a *CompositeAuthenticator) Authenticators() []Authenticator {
	flattened := make([]Authenticator, 0, len(a.authenticators))
	for _, handler := range a.authenticators {
		switch v := handler.(type) {
		case *CompositeAuthenticator:
			flattened = append(flattened, v.Authenticators()...)
		default:
			flattened = append(flattened, handler)
		}
	}
	sort.SliceStable(flattened, func(i, j int) bool {
		return order.OrderedFirstCompare(flattened[i], flattened[j])
	})
	return flattened
}

func (a *CompositeAuthenticator) Add(authenticator Authenticator) *CompositeAuthenticator {
	a.authenticators = a.processAuthenticators(append(a.authenticators, authenticator))
	sort.SliceStable(a.authenticators, func(i, j int) bool {
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
	sort.SliceStable(authenticators, func(i, j int) bool {
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

// CompositeAuthenticationSuccessHandler implement AuthenticationSuccessHandler interface
type CompositeAuthenticationSuccessHandler struct {
	init      sync.Once
	handlers  []AuthenticationSuccessHandler
	flattened []AuthenticationSuccessHandler
}

func NewAuthenticationSuccessHandler(handlers ...AuthenticationSuccessHandler) *CompositeAuthenticationSuccessHandler {
	ret := &CompositeAuthenticationSuccessHandler{}
	ret.handlers = ret.processSuccessHandlers(handlers)
	return ret
}

func (h *CompositeAuthenticationSuccessHandler) HandleAuthenticationSuccess(
	c context.Context, r *http.Request, rw http.ResponseWriter, from, to Authentication) {

	h.init.Do(func() { h.flattened = h.Handlers() })
	for _, handler := range h.flattened {
		handler.HandleAuthenticationSuccess(c, r, rw, from, to)
	}
}

// Handlers returns list of authentication handlers, any nested composite handlers are flattened
func (h *CompositeAuthenticationSuccessHandler) Handlers() []AuthenticationSuccessHandler {
	flattened := make([]AuthenticationSuccessHandler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		switch v := handler.(type) {
		case *CompositeAuthenticationSuccessHandler:
			flattened = append(flattened, v.Handlers()...)
		default:
			flattened = append(flattened, handler)
		}
	}
	sort.SliceStable(flattened, func(i, j int) bool {
		return order.OrderedFirstCompare(flattened[i], flattened[j])
	})
	return flattened
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
	sort.SliceStable(handlers, func(i, j int) bool {
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

// CompositeAuthenticatorBuilder implements AuthenticatorBuilder
type CompositeAuthenticatorBuilder struct {
	builders []AuthenticatorBuilder
}

func NewAuthenticatorBuilder() *CompositeAuthenticatorBuilder {
	return &CompositeAuthenticatorBuilder{builders: []AuthenticatorBuilder{}}
}

func (b *CompositeAuthenticatorBuilder) Build(c context.Context) (Authenticator, error) {
	authenticators := make([]Authenticator, len(b.builders))
	for i, builder := range b.builders {
		a, err := builder.Build(c)
		if err != nil {
			return nil, err
		}
		authenticators[i] = a
	}
	return NewAuthenticator(authenticators...), nil
}
