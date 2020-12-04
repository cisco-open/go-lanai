package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"net/http"
	"sort"
	"strings"
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

/*****************************
	Common Impl.
 *****************************/
// *CompositeAuthenticationErrorHandler implement AuthenticationErrorHandler interface
type CompositeAuthenticationErrorHandler struct {
	handlers []AuthenticationErrorHandler
}

func NewAuthenticationErrorHandler(handlers ...AuthenticationErrorHandler) *CompositeAuthenticationErrorHandler {
	ret := &CompositeAuthenticationErrorHandler{}
	ret.handlers = ret.processErrorHandlers(handlers)
	return ret
}

func (h *CompositeAuthenticationErrorHandler) HandleAuthenticationError(
	c context.Context, r *http.Request, rw http.ResponseWriter, err error) {

	for _,handler := range h.handlers {
		handler.HandleAuthenticationError(c, r, rw, err)
	}
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
	sort.SliceStable(handlers, func(i,j int) bool {
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

/**************************
	Default Impls
***************************/
// *DefaultAccessDeniedHandler implements AccessDeniedHandler
type DefaultAccessDeniedHandler struct {
}

func (h *DefaultAccessDeniedHandler) HandleAccessDenied(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	if isJson(r) {
		writeErrorAsJson(ctx, http.StatusForbidden, err, rw)
	} else {
		writeErrorAsHtml(ctx, http.StatusForbidden, err, rw)
	}
}

// *DefaultAuthenticationErrorHandler implements AuthenticationErrorHandler
type DefaultAuthenticationErrorHandler struct {
}

func (h *DefaultAuthenticationErrorHandler) HandleAuthenticationError(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	if isJson(r) {
		writeErrorAsJson(ctx, http.StatusForbidden, err, rw)
	} else {
		writeErrorAsHtml(ctx, http.StatusForbidden, err, rw)
	}
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

func writeErrorAsHtml(ctx context.Context, code int, err error, rw http.ResponseWriter) {
	httpError := web.NewHttpError(code, err)
	template.TemplateErrorEncoder(ctx, httpError, rw)
}

func writeErrorAsJson(ctx context.Context, code int, err error, rw http.ResponseWriter) {
	httpError := web.NewHttpError(code, err)
	rest.JsonErrorEncoder(ctx, httpError, rw)
}