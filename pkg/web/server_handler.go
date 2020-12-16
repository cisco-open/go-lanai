package web

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

//TODO: define interface here, security provides a implementation and make it available

type RequestCacheMatcher interface {
	PopMatchedRequest(r *http.Request) *CachedRequest
}

type CachedRequest struct {
	Method   string
	URL      *url.URL
	BodyData string
	Host     string
}

func (c *CachedRequest) getBody() *CachedRequestBody {
	b := &CachedRequestBody{
		src: strings.NewReader(c.BodyData),
	}
	return b
}

type CachedRequestBody struct {
	src io.Reader
	mu sync.Mutex
}

func (b *CachedRequestBody)	Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.src.Read(p)
}

func (b *CachedRequestBody) Close() error {
	return nil
}

type Engine struct {
	*gin.Engine
	requestCacheMatcher RequestCacheMatcher
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cached := e.requestCacheMatcher.PopMatchedRequest(r)
	if cached != nil {
		r.Method = cached.Method
		//because popMatchRequest only matches on GET, so original body is always http.nobody which is safe to replace
		r.Body = cached.getBody()
	}

	e.Engine.ServeHTTP(w, r)
}

func NewEngine(matcher RequestCacheMatcher) *Engine {
	e := &Engine{
		gin.Default(),
		matcher,
	}
	return e
}