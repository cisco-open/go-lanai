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

package request_cache

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
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
	Form     url.Values
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

// CachedRequestPreProcessor is designed to be used by code outside of the security package.
// Implements the web.RequestCacheAccessor interface
type CachedRequestPreProcessor struct {
	sessionName string
	store       session.Store
	name        web.RequestPreProcessorName
}

func newCachedRequestPreProcessor(sessionName string, store session.Store) *CachedRequestPreProcessor {
	return &CachedRequestPreProcessor{
		sessionName: sessionName,
		store: store,
		name:  "CachedRequestPreProcessor",
	}
}

func (p *CachedRequestPreProcessor) Name() web.RequestPreProcessorName {
	return p.name
}

func (p *CachedRequestPreProcessor) Process(r *http.Request) error {
	if cookie, err := r.Cookie(p.sessionName); err == nil {
		id := cookie.Value
		if s, err := p.store.WithContext(r.Context()).Get(id, p.sessionName); err == nil {
			cached, ok := s.Get(SessionKeyCachedRequest).(*CachedRequest)
			if ok && cached != nil && requestMatches(r, cached) {
				s.Delete(SessionKeyCachedRequest)
				err := p.store.WithContext(r.Context()).Save(s)
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

func NewSavedRequestAuthenticationSuccessHandler(fallback security.AuthenticationSuccessHandler, condition func(from, to security.Authentication) bool) security.AuthenticationSuccessHandler {
	if condition == nil {
		condition = security.IsBeingAuthenticated
	}
	return &SavedRequestAuthenticationSuccessHandler{
		condition: condition,
		fallback:  fallback,
	}
}

type SavedRequestAuthenticationSuccessHandler struct {
	condition func(from, to security.Authentication) bool
	fallback  security.AuthenticationSuccessHandler
}

func (h *SavedRequestAuthenticationSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	if !h.condition(from, to) {
		return
	}

	cached := GetCachedRequest(c)

	if cached != nil {
		http.Redirect(rw, r, cached.URL.RequestURI(), 302)
		_, _ = rw.Write([]byte{})
		return
	}

	h.fallback.HandleAuthenticationSuccess(c, r, rw, from, to)
}

type SaveRequestEntryPoint struct {
	delegate           security.AuthenticationEntryPoint
	saveRequestMatcher web.RequestMatcher
}

func NewSaveRequestEntryPoint(delegate security.AuthenticationEntryPoint) *SaveRequestEntryPoint {
	notFavicon := matcher.NotRequest(matcher.RequestWithPattern("/**/favicon.*"))
	notXMLHttpRequest := matcher.NotRequest(matcher.RequestWithHeader("X-Requested-With", "XMLHttpRequest", false))
	notTrailer := matcher.NotRequest(matcher.RequestHasHeader("Trailer"))
	notMultiPart := matcher.NotRequest(matcher.RequestWithHeader("Content-Type", "multipart/form-data", true))
	notCsrf := matcher.NotRequest(matcher.RequestHasHeader(security.CsrfHeaderName).Or(matcher.RequestHasPostForm(security.CsrfParamName)))

	saveRequestMatcher := notFavicon.And(notXMLHttpRequest).And(notTrailer).And(notMultiPart).And(notCsrf)

	return &SaveRequestEntryPoint{
		delegate,
		saveRequestMatcher,
	}
}

func (s *SaveRequestEntryPoint) Commence(c context.Context, r *http.Request, w http.ResponseWriter, e error) {
	match, err := s.saveRequestMatcher.MatchesWithContext(c, r)
	if match && err == nil {
		SaveRequest(c)
	}
	s.delegate.Commence(c, r, w, e)
}
