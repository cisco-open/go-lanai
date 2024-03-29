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
    "errors"
    "github.com/cisco-open/go-lanai/pkg/utils/order"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/template"
    "github.com/gin-gonic/gin"
    "net/http"
    "sort"
    "strings"
    "sync"
)

// AccessDeniedHandler handles ErrorSubTypeAccessDenied
type AccessDeniedHandler interface {
	HandleAccessDenied(context.Context, *http.Request, http.ResponseWriter, error)
}

// AuthenticationErrorHandler handles ErrorTypeAuthentication
type AuthenticationErrorHandler interface {
	HandleAuthenticationError(context.Context, *http.Request, http.ResponseWriter, error)
}

// AuthenticationEntryPoint kicks off authentication process
type AuthenticationEntryPoint interface {
	Commence(context.Context, *http.Request, http.ResponseWriter, error)
}

// ErrorHandler handles any other type of errors
type ErrorHandler interface {
	HandleError(context.Context, *http.Request, http.ResponseWriter, error)
}

/*****************************
	Common Impl.
 *****************************/

// CompositeAuthenticationErrorHandler implement AuthenticationErrorHandler interface
type CompositeAuthenticationErrorHandler struct {
	init      sync.Once
	handlers  []AuthenticationErrorHandler
	flattened []AuthenticationErrorHandler
}

func NewAuthenticationErrorHandler(handlers ...AuthenticationErrorHandler) *CompositeAuthenticationErrorHandler {
	ret := &CompositeAuthenticationErrorHandler{}
	ret.handlers = ret.processErrorHandlers(handlers)
	return ret
}

func (h *CompositeAuthenticationErrorHandler) HandleAuthenticationError(
	c context.Context, r *http.Request, rw http.ResponseWriter, err error) {

	h.init.Do(func() { h.flattened = h.Handlers() })
	for _, handler := range h.flattened {
		handler.HandleAuthenticationError(c, r, rw, err)
	}
}

// Handlers returns list of authentication handlers, any nested composite handlers are flattened
func (h *CompositeAuthenticationErrorHandler) Handlers() []AuthenticationErrorHandler {
	flattened := make([]AuthenticationErrorHandler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		switch h := handler.(type) {
		case *CompositeAuthenticationErrorHandler:
			flattened = append(flattened, h.Handlers()...)
		default:
			flattened = append(flattened, handler)
		}
	}
	sort.SliceStable(flattened, func(i, j int) bool {
		return order.OrderedFirstCompare(flattened[i], flattened[j])
	})
	return flattened
}

func (h *CompositeAuthenticationErrorHandler) Size() int {
	return len(h.handlers)
}

func (h *CompositeAuthenticationErrorHandler) Add(handler AuthenticationErrorHandler) *CompositeAuthenticationErrorHandler {
	h.handlers = h.processErrorHandlers(append(h.handlers, handler))
	return h
}

func (h *CompositeAuthenticationErrorHandler) Merge(composite *CompositeAuthenticationErrorHandler) *CompositeAuthenticationErrorHandler {
	h.handlers = h.processErrorHandlers(append(h.handlers, composite.handlers...))
	return h
}

func (h *CompositeAuthenticationErrorHandler) processErrorHandlers(handlers []AuthenticationErrorHandler) []AuthenticationErrorHandler {
	handlers = h.removeSelf(handlers)
	handlers = h.removeDuplicates(handlers)
	sort.SliceStable(handlers, func(i, j int) bool {
		return order.OrderedFirstCompare(handlers[i], handlers[j])
	})
	return handlers
}

func (h *CompositeAuthenticationErrorHandler) removeSelf(items []AuthenticationErrorHandler) []AuthenticationErrorHandler {
	count := 0
	for _, item := range items {
		if ptr, ok := item.(*CompositeAuthenticationErrorHandler); !ok || ptr != h {
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

func (h *CompositeAuthenticationErrorHandler) removeDuplicates(items []AuthenticationErrorHandler) []AuthenticationErrorHandler {
	lookup := map[AuthenticationErrorHandler]struct{}{}
	unique := make([]AuthenticationErrorHandler, 0, len(items))
	for _, v := range items {
		if _, ok := lookup[v]; ok {
			continue
		}
		lookup[v] = struct{}{}
		unique = append(unique, v)
	}
	return unique
}

// CompositeAccessDeniedHandler implement AccessDeniedHandler interface
type CompositeAccessDeniedHandler struct {
	init sync.Once
	handlers  []AccessDeniedHandler
	flattened []AccessDeniedHandler
}

func NewAccessDeniedHandler(handlers ...AccessDeniedHandler) *CompositeAccessDeniedHandler {
	ret := &CompositeAccessDeniedHandler{}
	ret.handlers = ret.processErrorHandlers(handlers)
	return ret
}

func (h *CompositeAccessDeniedHandler) HandleAccessDenied(
	c context.Context, r *http.Request, rw http.ResponseWriter, err error) {

	h.init.Do(func() { h.flattened = h.Handlers()})
	for _, handler := range h.flattened {
		handler.HandleAccessDenied(c, r, rw, err)
	}
}

// Handlers returns list of authentication handlers, any nested composite handlers are flattened
func (h *CompositeAccessDeniedHandler) Handlers() []AccessDeniedHandler {
	flattened := make([]AccessDeniedHandler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		switch h := handler.(type) {
		case *CompositeAccessDeniedHandler:
			flattened = append(flattened, h.Handlers()...)
		default:
			flattened = append(flattened, handler)
		}
	}
	sort.SliceStable(flattened, func(i, j int) bool {
		return order.OrderedFirstCompare(flattened[i], flattened[j])
	})
	return flattened
}

func (h *CompositeAccessDeniedHandler) Size() int {
	return len(h.handlers)
}

func (h *CompositeAccessDeniedHandler) Add(handler AccessDeniedHandler) *CompositeAccessDeniedHandler {
	h.handlers = h.processErrorHandlers(append(h.handlers, handler))
	return h
}

func (h *CompositeAccessDeniedHandler) Merge(composite *CompositeAccessDeniedHandler) *CompositeAccessDeniedHandler {
	h.handlers = h.processErrorHandlers(append(h.handlers, composite.handlers...))
	return h
}

func (h *CompositeAccessDeniedHandler) processErrorHandlers(handlers []AccessDeniedHandler) []AccessDeniedHandler {
	handlers = h.removeSelf(handlers)
	handlers = h.removeDuplicates(handlers)
	sort.SliceStable(handlers, func(i, j int) bool {
		return order.OrderedFirstCompare(handlers[i], handlers[j])
	})
	return handlers
}

func (h *CompositeAccessDeniedHandler) removeSelf(items []AccessDeniedHandler) []AccessDeniedHandler {
	count := 0
	for _, item := range items {
		if ptr, ok := item.(*CompositeAccessDeniedHandler); !ok || ptr != h {
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

func (h *CompositeAccessDeniedHandler) removeDuplicates(items []AccessDeniedHandler) []AccessDeniedHandler {
	lookup := map[AccessDeniedHandler]struct{}{}
	unique := make([]AccessDeniedHandler, 0, len(items))
	for _, v := range items {
		if _, ok := lookup[v]; ok {
			continue
		}
		lookup[v] = struct{}{}
		unique = append(unique, v)
	}
	return unique
}

// *CompositeErrorHandler implement ErrorHandler interface
type CompositeErrorHandler struct {
	handlers []ErrorHandler
}

func NewErrorHandler(handlers ...ErrorHandler) *CompositeErrorHandler {
	ret := &CompositeErrorHandler{}
	ret.handlers = ret.processErrorHandlers(handlers)
	return ret
}

func (h *CompositeErrorHandler) HandleError(
	c context.Context, r *http.Request, rw http.ResponseWriter, err error) {

	for _, handler := range h.handlers {
		handler.HandleError(c, r, rw, err)
	}
}

func (h *CompositeErrorHandler) Size() int {
	return len(h.handlers)
}

func (h *CompositeErrorHandler) Add(handler ErrorHandler) *CompositeErrorHandler {
	h.handlers = h.processErrorHandlers(append(h.handlers, handler))
	return h
}

func (h *CompositeErrorHandler) Merge(composite *CompositeErrorHandler) *CompositeErrorHandler {
	h.handlers = h.processErrorHandlers(append(h.handlers, composite.handlers...))
	return h
}

func (h *CompositeErrorHandler) processErrorHandlers(handlers []ErrorHandler) []ErrorHandler {
	handlers = h.removeSelf(handlers)
	handlers = h.removeDuplicates(handlers)
	sort.SliceStable(handlers, func(i, j int) bool {
		return order.OrderedFirstCompare(handlers[i], handlers[j])
	})
	return handlers
}

func (h *CompositeErrorHandler) removeSelf(items []ErrorHandler) []ErrorHandler {
	count := 0
	for _, item := range items {
		if ptr, ok := item.(*CompositeErrorHandler); !ok || ptr != h {
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

func (h *CompositeErrorHandler) removeDuplicates(items []ErrorHandler) []ErrorHandler {
	lookup := map[ErrorHandler]struct{}{}
	unique := make([]ErrorHandler, 0, len(items))
	for _, v := range items {
		if _, ok := lookup[v]; ok {
			continue
		}
		lookup[v] = struct{}{}
		unique = append(unique, v)
	}
	return unique
}

/**************************
	Default Impls
***************************/

// DefaultAccessDeniedHandler implements AccessDeniedHandler
type DefaultAccessDeniedHandler struct {
}

func (h *DefaultAccessDeniedHandler) HandleAccessDenied(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrorSubTypeInsufficientAuth):
		WriteError(ctx, r, rw, http.StatusUnauthorized, err)
	}
	WriteError(ctx, r, rw, http.StatusForbidden, err)
}

// DefaultAuthenticationErrorHandler implements AuthenticationErrorHandler
type DefaultAuthenticationErrorHandler struct {
}

func (h *DefaultAuthenticationErrorHandler) HandleAuthenticationError(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	WriteError(ctx, r, rw, http.StatusUnauthorized, err)
}

// DefaultErrorHandler implements ErrorHandler
type DefaultErrorHandler struct{}

func (h *DefaultErrorHandler) HandleError(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	WriteError(ctx, r, rw, http.StatusUnauthorized, err)
}

/**************************
	Common Functions
***************************/

func WriteError(ctx context.Context, r *http.Request, rw http.ResponseWriter, code int, err error) {
	if IsResponseWritten(rw) {
		return
	}

	if isJson(r) {
		WriteErrorAsJson(ctx, rw, code, err)
	} else {
		WriteErrorAsHtml(ctx, rw, code, err)
	}
}

func IsResponseWritten(rw http.ResponseWriter) bool {
	ginRw, ok := rw.(gin.ResponseWriter)
	return ok && ginRw.Written()
}

func WriteErrorAsHtml(ctx context.Context, rw http.ResponseWriter, code int, err error) {
	httpError := web.NewHttpError(code, err)
	template.TemplateErrorEncoder(ctx, httpError, rw)
}

func WriteErrorAsJson(ctx context.Context, rw http.ResponseWriter, code int, err error) {
	httpError := web.NewHttpError(code, err)
	web.JsonErrorEncoder()(ctx, httpError, rw)
}

/**************************
	Helpers
***************************/
func isJson(r *http.Request) bool {
	// TODO should be more comprehensive than this
	accept := r.Header.Get("Accept")
	contentType := r.Header.Get("Content-Type")
	return strings.Contains(accept, "application/json") || strings.Contains(contentType, "application/json")
}
