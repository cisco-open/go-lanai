package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

type CachedRequest interface {
	GetMethod()   string
	GetURL()      *url.URL
	GetHeader()   http.Header
	GetForm() 	 url.Values
	GetPostForm() url.Values
	GetHost()     string
}

type RequestCacheAccessor interface {
	PopMatchedRequest(r *http.Request) CachedRequest
}

type Engine struct {
	*gin.Engine
	requestCacheMatcher RequestCacheAccessor
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cached := e.requestCacheMatcher.PopMatchedRequest(r)
	if cached != nil {
		r.Method = cached.GetMethod()
		//because popMatchRequest only matches on GET, so incoming request body is always http.nobody
		//therefore we set the form and post form directly.
		//multi part form (used for file uploads) are not supported - if original request was multi part form, it's not cached.
		//trailer headers are also not supported - if original request has trailer, it's not cached.
		r.Form = cached.GetForm()
		r.PostForm = cached.GetPostForm()
		//get all the headers from the cached request except the cookie header
		cookie := r.Header["Cookie"]
		r.Header = cached.GetHeader()
		r.Header["Cookie"] = cookie
	}

	e.Engine.ServeHTTP(w, r)
}

func NewEngine(matcher RequestCacheAccessor) *Engine {
	e := &Engine{
		gin.Default(),
		matcher,
	}
	return e
}