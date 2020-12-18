package requestcache

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"reflect"
)

const SessionKeyCachedRequest = "CachedRequest"

type Request struct {
	Method   string
	URL      *url.URL
	Header   http.Header
	Form 	 url.Values
	PostForm url.Values
	Host     string
}

func(r *Request) GetMethod()   string {
	return r.Method
}

func(r *Request) GetURL()      *url.URL {
	return r.URL
}

func(r *Request) GetHeader()   http.Header {
	return r.Header
}

func(r *Request) GetForm() 	 url.Values {
	return r.Form
}

func(r *Request) GetPostForm() url.Values {
	return r.PostForm
}

func(r *Request) GetHost()     string {
	return r.Host
}


func SaveRequest(ctx *gin.Context) {
	s := session.Get(ctx)
	// we don't know if other components have already parsed the form.
	// if other components have already parsed the form, then the body is already read, so if we read it again we'll just get ""
	// therefore we call parseForm to make sure it's read into the form field, and we serialize the form field ourselves.
	_ = ctx.Request.ParseForm()

	cached := &Request{
		Method:   ctx.Request.Method,
		URL:      ctx.Request.URL,
		Host:     ctx.Request.Host,
		PostForm: ctx.Request.PostForm,
		Form: 	  ctx.Request.Form,
		Header:   ctx.Request.Header,
	}
	s.Set(SessionKeyCachedRequest, cached)
}

func GetCachedRequest(ctx *gin.Context) *Request {
	s := session.Get(ctx)
	cached, _ := s.Get(SessionKeyCachedRequest).(*Request)
	return cached
}

func RemoveCachedRequest(ctx *gin.Context) {
	s := session.Get(ctx)
	s.Delete(SessionKeyCachedRequest)
}

// Designed to be used by code outside of the security package.
// Implements the web.RequestCacheAccessor interface
type Accessor struct {
	store session.Store
}

func (m *Accessor) PopMatchedRequest(r *http.Request) (web.CachedRequest, error) {
	if cookie, err := r.Cookie(session.DefaultName); err == nil {
		id := cookie.Value
		if s, err := m.store.Get(id, session.DefaultName); err == nil {
			cached, ok := s.Get(SessionKeyCachedRequest).(*Request)
			if ok && cached != nil && requestMatches(r, cached) {
				s.Delete(SessionKeyCachedRequest)
				err := m.store.Save(s)
				if err != nil {
					return nil, err
				}
				return cached, nil
			}
		}
	}
	return nil, nil
}

func requestMatches(r *http.Request, cached *Request) bool {
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
