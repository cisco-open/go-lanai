package request_cache

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"reflect"
)

const SessionKeyCachedRequest = "CachedRequest"

type CachedRequest struct {
	Method   string
	URL      *url.URL
	Header   http.Header
	Form 	 url.Values
	PostForm url.Values
	Host     string
}

func SaveRequest(ctx context.Context) {
	gc := web.GinContext(ctx)
	if gc == nil {
		return
	}

	s := session.Get(ctx)
	// we don't know if other components have already parsed the form.
	// if other components have already parsed the form, then the body is already read, so if we read it again we'll just get ""
	// therefore we call parseForm to make sure it's read into the form field, and we serialize the form field ourselves.
	_ = gc.Request.ParseForm()

	cached := &CachedRequest{
		Method:   gc.Request.Method,
		URL:      gc.Request.URL,
		Host:     gc.Request.Host,
		PostForm: gc.Request.PostForm,
		Form:     gc.Request.Form,
		Header:   gc.Request.Header,
	}
	s.Set(SessionKeyCachedRequest, cached)
}

func GetCachedRequest(ctx context.Context) *CachedRequest {
	s := session.Get(ctx)
	cached, _ := s.Get(SessionKeyCachedRequest).(*CachedRequest)
	return cached
}

func RemoveCachedRequest(ctx *gin.Context) {
	s := session.Get(ctx)
	s.Delete(SessionKeyCachedRequest)
}

// Designed to be used by code outside of the security package.
// Implements the web.RequestCacheAccessor interface
type CachedRequestPreProcessor struct {
	store session.Store
	name web.RequestPreProcessorName
}

func newCachedRequestPreProcessor(store session.Store) *CachedRequestPreProcessor {
	return &CachedRequestPreProcessor{
		store:store,
		name: "CachedRequestPreProcessor",
	}
}

func (p *CachedRequestPreProcessor) Name() web.RequestPreProcessorName {
	return p.name
}

func (p *CachedRequestPreProcessor) Process(r *http.Request) error {
	if cookie, err := r.Cookie(session.DefaultName); err == nil {
		id := cookie.Value
		if s, err := p.store.Get(id, session.DefaultName); err == nil {
			cached, ok := s.Get(SessionKeyCachedRequest).(*CachedRequest)
			if ok && cached != nil && requestMatches(r, cached) {
				s.Delete(SessionKeyCachedRequest)
				err := p.store.Save(s)
				if err != nil {
					return err
				}

				r.Method = cached.Method
				//because popMatchRequest only matches on GET, so incoming request body is always http.nobody
				//therefore we set the form and post form directly.
				//multi part form (used for file uploads) are not supported - if original request was multi part form, it's not cached.
				//trailer headers are also not supported - if original request has trailer, it's not cached.
				r.Form = cached.Form
				r.PostForm = cached.PostForm
				//get all the headers from the cached request except the cookie header
				if cached.Header != nil {
					cookie := r.Header["Cookie"]
					r.Header = cached.Header
					r.Header["Cookie"] = cookie
				}
				return nil
			}
		}
	}
	return nil
}

func requestMatches(r *http.Request, cached *CachedRequest) bool {
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
	if !security.IsBeingAuthenticated(from, to) {
		return
	}

	cached := GetCachedRequest(c)

	if cached != nil {
		http.Redirect(rw, r, cached.URL.RequestURI(), 302)
		_,_ = rw.Write([]byte{})
		return
	}

	h.fallback.HandleAuthenticationSuccess(c, r, rw, from, to)
}

type SaveRequestEntryPoint struct {
	delegate security.AuthenticationEntryPoint
	saveRequestMatcher web.RequestMatcher
}

func NewSaveRequestEntryPoint(delegate security.AuthenticationEntryPoint) *SaveRequestEntryPoint {
	notFavicon := matcher.NotRequest(matcher.RequestWithPattern("/**/favicon.*"))
	notXMLHttpRequest := matcher.NotRequest(matcher.RequestWithHeader("X-Requested-With", "XMLHttpRequest", false))
	notTrailer := matcher.NotRequest(matcher.RequestHasHeader("Trailer"))
	notMultiPart := matcher.NotRequest(matcher.RequestWithHeader("Content-Type", "multipart/form-data", true))
	notCsrf := matcher.NotRequest(matcher.RequestHasHeader(security.CsrfHeaderName).Or(matcher.RequestHasPostParameter(security.CsrfParamName)))

	saveRequestMatcher := notFavicon.And(notXMLHttpRequest).And(notTrailer).And(notMultiPart).And(notCsrf)

	return &SaveRequestEntryPoint{
		delegate,
		saveRequestMatcher,
	}
}

func (s *SaveRequestEntryPoint) Commence(c context.Context, r *http.Request, w http.ResponseWriter, e error) {
	match, err := s.saveRequestMatcher.MatchesWithContext(c, r)
	if match && err == nil{
		SaveRequest(c)
	}
	s.delegate.Commence(c, r, w, e)
}