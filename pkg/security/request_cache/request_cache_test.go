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
	"github.com/cisco-open/go-lanai/pkg/security/session/common"
	"github.com/cisco-open/go-lanai/test/mocks/authmock"
	"github.com/cisco-open/go-lanai/test/mocks/redismock"
	"github.com/cisco-open/go-lanai/test/mocks/sessionmock"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSaveAndGetCachedRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := redismock.NewMockUniversalClient(ctrl)

	sessionStore := session.NewRedisStore(mockRedis)
	s, _ := sessionStore.New(common.DefaultName)

	c := webtest.NewGinContext(context.Background(),
		"POST", "/something", strings.NewReader("a=b&c=d"),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	SaveRequest(c)
	cached := GetCachedRequest(c)

	if cached.Method != "POST" {
		t.Errorf("expected POST, but got %s", cached.Method)
	}

	if cached.PostForm.Get("a") != "b" && cached.PostForm.Get("c") != "d" {
		t.Errorf("expected post form to have a=b and c=d")
	}
}


func TestCachedRequestPreProcessor_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionStore := sessionmock.NewMockStore(ctrl)

	processor := newCachedRequestPreProcessor(common.DefaultName, mockSessionStore)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})

	s := session.CreateSession(mockSessionStore, common.DefaultName)

	cached := &CachedRequest{
		Host: "example.com",
		Method: "POST",
		URL: &url.URL{Path: "/something"},
		PostForm: url.Values{"a":[]string{"b"},"c":[]string{"d"}},
	}

	s.Set(SessionKeyCachedRequest, cached)
	mockSessionStore.EXPECT().WithContext(gomock.Any()).Return(mockSessionStore).AnyTimes()
	mockSessionStore.EXPECT().Get(s.GetID(), s.Name()).Return(s, nil)
	mockSessionStore.EXPECT().Save(s).Do(func(session *session.Session) {
		if session.Get(SessionKeyCachedRequest) != nil {
			t.Errorf("cached request should be removed from the session")
		}
	})

	//GET request to the same path
	req := httptest.NewRequest("GET", "/something", nil)
	req.Header.Set("Cookie", common.DefaultName+"="+s.GetID())

	_ = processor.Process(req)

	if req.Method != "POST" {
		t.Errorf("expect the method to be changed to match the cached request")
	}

	if req.PostForm.Get("a") != "b" && req.PostForm.Get("c") != "d" {
		t.Errorf("expected post form to have cached value")
	}
}

func TestSavedRequestAuthenticationSuccessHandler_HandleAuthenticationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := redismock.NewMockUniversalClient(ctrl)

	sessionStore := session.NewRedisStore(mockRedis)
	s, _ := sessionStore.New(common.DefaultName)

	c := webtest.NewGinContext(context.Background(),
		"POST", "/something", strings.NewReader("a=b&c=d"),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}

	SaveRequest(c)

	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)

	handler := NewSavedRequestAuthenticationSuccessHandler(nil, nil)

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)

	recorder := webtest.GinContextRecorder(c)
	if recorder.Result().StatusCode != 302 {
		t.Errorf("expected 302 but got %v ", recorder.Result().StatusCode )
	}

	l, _ := recorder.Result().Location()

	if l.String() != "/something" {
		t.Errorf("expected redirect location, got %v", l)
	}
}

func TestSaveRequestEntryPoint_Commence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := redismock.NewMockUniversalClient(ctrl)

	sessionStore := session.NewRedisStore(mockRedis)
	s, _ := sessionStore.New(common.DefaultName)
	c := webtest.NewGinContext(context.Background(),
		"GET", "/something/favicon.jpg", nil,
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}

	entryPoint := NewSaveRequestEntryPoint(&noOpEntryPoint{})

	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request for favicon should not be cached")
	}

	s, _ = sessionStore.New(common.DefaultName)
	c = webtest.NewGinContext(context.Background(),
		"GET", "/something", nil,
		webtest.Headers("X-Requested-With", "XMLHttpRequest"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with  XMLHttpRequest should not be cached")
	}

	s, _ = sessionStore.New(common.DefaultName)
	c = webtest.NewGinContext(context.Background(),
		"GET", "/something", nil,
		webtest.Headers("Trailer", "anything"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with  XMLHttpRequest should not be cached")
	}

	s, _ = sessionStore.New(common.DefaultName)
	c = webtest.NewGinContext(context.Background(),
		"GET", "/something", nil,
		webtest.Headers("Content-Type", "multipart/form-data something"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with multipart/form-data should not be cached")
	}

	s, _ = sessionStore.New(common.DefaultName)
	c = webtest.NewGinContext(context.Background(),
		"POST", "/something", strings.NewReader("a=b&c=d"),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded", security.CsrfHeaderName, "something"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with csrf header should not be cached")
	}

	s, _ = sessionStore.New(common.DefaultName)
	c = webtest.NewGinContext(context.Background(),
		"POST", "/something", strings.NewReader(security.CsrfParamName + "=something"),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with csrf param should not be cached")
	}

	s, _ = sessionStore.New(common.DefaultName)
	c = webtest.NewGinContext(context.Background(),
		"POST", "/something", strings.NewReader("a=b&c=d"),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))

	if GetCachedRequest(c) == nil {
		t.Errorf("expect request to be cached")
	}
}

type noOpEntryPoint struct {}
func (e *noOpEntryPoint) Commence(context.Context, *http.Request, http.ResponseWriter, error) {}

