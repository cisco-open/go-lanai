package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

const SessionKeyCachedRequest = "CachedRequest"

func SaveRequest(ctx *gin.Context) {
	s := Get(ctx)
	// we don't know if other components have already parsed the form.
	// if other components have already parsed the form, then the body is already read, so if we read it again we'll just get ""
	// therefore we call parseForm to make sure it's read into the form field, and we serialize the form field ourselves.
	_ = ctx.Request.ParseForm()

	cached := &web.CachedRequest{
		Method:   ctx.Request.Method,
		URL:      ctx.Request.URL,
		Host:     ctx.Request.Host,
		PostForm: ctx.Request.PostForm,
		Form: 	  ctx.Request.Form,
		Header:   ctx.Request.Header,
	}
	s.Set(SessionKeyCachedRequest, cached)
}

func GetCachedRequest(ctx *gin.Context) *web.CachedRequest {
	s := Get(ctx)
	cached, _ := s.Get(SessionKeyCachedRequest).(*web.CachedRequest)
	return cached
}

func RemoveCachedRequest(ctx *gin.Context) {
	s := Get(ctx)
	s.Delete(SessionKeyCachedRequest)
}

type RequestCacheMatcher struct {
	store Store
}

func (m *RequestCacheMatcher) PopMatchedRequest(r *http.Request) *web.CachedRequest {
	if cookie, err := r.Cookie(DefaultName); err == nil {
		id := cookie.Value
		if s, err := m.store.Get(id, DefaultName); err == nil {
			cached, ok := s.Get(SessionKeyCachedRequest).(*web.CachedRequest)
			if ok && cached != nil && requestMatches(r, cached) {
				s.Delete(SessionKeyCachedRequest)
				m.store.Save(s)
				return cached
			}
		}
	}
	return nil
}

func requestMatches(r *http.Request, cached *web.CachedRequest) bool {
	// Only support matching incoming GET command, because we will only issue redirect after auth success.
	if r.Method != "GET" {
		return false
	}
	return reflect.DeepEqual(r.URL, cached.URL) && r.Host == cached.Host
}

func NewSavedRequestAuthenticationSuccessHandler(fallback security.AuthenticationSuccessHandler) security.AuthenticationSuccessHandler {
	return &SavedRequestAuthenticationSuccessHandler{
		fallback: fallback,
	}
}

type SavedRequestAuthenticationSuccessHandler struct {
	fallback security.AuthenticationSuccessHandler
}

func (h *SavedRequestAuthenticationSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	if g, ok := c.(*gin.Context); ok {
		cached := GetCachedRequest(g)

		if cached != nil {
			http.Redirect(rw, r, cached.URL.RequestURI(), 302)
			_,_ = rw.Write([]byte{})
			return
		}
	}

	h.fallback.HandleAuthenticationSuccess(c, r, rw, from, to)
}
