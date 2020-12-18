package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

//TODO: define interface here, security provides a implementation and make it available

type RequestCacheMatcher interface {
	PopMatchedRequest(r *http.Request) *CachedRequest
}

type CachedRequest struct {
	Method   string
	URL      *url.URL
	Header   http.Header
	Form 	 url.Values
	PostForm url.Values
	Host     string
}

func (c *CachedRequest) GetRedirect() {

}

type Engine struct {
	*gin.Engine
	requestCacheMatcher RequestCacheMatcher
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cached := e.requestCacheMatcher.PopMatchedRequest(r)
	if cached != nil {
		r.Method = cached.Method
		//because popMatchRequest only matches on GET, so incoming request body is always http.nobody
		//therefore we set the form and post form directly.
		//multi part form (used for file uploads) are not supported - if original request was multi part form, it's not cached.
		//trailer headers are also not supported - if original request has trailer, it's not cached.
		r.Form = cached.Form
		r.PostForm = cached.PostForm
		//get all the headers from the cached request except the cookie header
		cookie := r.Header["Cookie"]
		r.Header = cached.Header
		r.Header["Cookie"] = cookie
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