package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"reflect"
)

const SessionKeyCachedRequest = "CachedRequest"

func SaveRequest(ctx *gin.Context) {
	//TODO: not all request should be saved
	// see Spring RequestCacheConfigurable.createDefaultSavedRequestMatcher

	s := Get(ctx)
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	cached := &web.CachedRequest{
		Method:   ctx.Request.Method,
		URL:      ctx.Request.URL,
		Host:     ctx.Request.Host,
		BodyData: string(body),
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