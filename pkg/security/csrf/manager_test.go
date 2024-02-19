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

package csrf

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/sessionmock"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestCsrfMiddlewareShouldGenerateToken(t *testing.T) {
	csrfStore := newSessionBackedStore()
	manager := newManager(csrfStore, nil, nil)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := sessionmock.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, common.DefaultName)

	c := webtest.NewGinContext(context.Background(), "GET", "/form", nil)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}

	mockSessionStore.EXPECT().Save(gomock.Any()).Do(func(s *session.Session) {
		savedCsrfToken := s.Get(SessionKeyCsrfToken)
		if savedCsrfToken == nil {
			t.Errorf("Expect the csrf token to be saved in the session")
		}
		if savedCsrfToken.(*Token).ParameterName != "_csrf" {
			t.Errorf("Expected parameter name to be _csrf, but was %v", savedCsrfToken.(Token).ParameterName)
		}
		if savedCsrfToken.(*Token).HeaderName != "X-CSRF-TOKEN" {
			t.Errorf("Expected header name to be X-CSRF-TOKEN, but was %v", savedCsrfToken.(Token).HeaderName)
		}
		if savedCsrfToken.(*Token).Value == "" {
			t.Errorf("Expect csrf token value to not be empty")
		}

	})
	mw := manager.CsrfHandlerFunc()
	mw(c)

	csrfToken := Get(c)

	if csrfToken == nil || csrfToken.Value == "" {
		t.Errorf("expected to have session")
	}
}

func TestCsrfMiddlewareShouldCheckToken(t *testing.T) {
	csrfStore := newSessionBackedStore()
	manager := newManager(csrfStore, nil, nil)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := sessionmock.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, common.DefaultName)

	//RequestDetails with a invalid csrf token
	c := webtest.NewGinContext(context.Background(),
		"POST", "/process", strings.NewReader("_csrf="+uuid.New().String()),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"))
	existingCsrfToken := csrfStore.Generate(c, "_csrf", "X-CSRF-TOKEN")
	s.Set(SessionKeyCsrfToken, existingCsrfToken)

	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	mw := manager.CsrfHandlerFunc()
	mw(c)

	if len(c.Errors) != 1 {
		t.Errorf("there should be one error")
	}

	if !errors.Is(c.Errors.Last().Err, security.NewInvalidCsrfTokenError("")) {
		t.Errorf("expect invalid csrf token error, but was %v", c.Errors.Last())
	}

	//RequestDetails without csrf token
	c = webtest.NewGinContext(context.Background(),
		"POST", "/process", strings.NewReader("a=b"),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	mw(c)

	if len(c.Errors) != 1 {
		t.Errorf("there should be one error")
	}

	if !errors.Is(c.Errors.Last().Err, security.NewMissingCsrfTokenError("")) {
		t.Errorf("expect missing csrf token error, but was %v", c.Errors.Last())
	}

	//RequestDetails with expected csrf token
	c = webtest.NewGinContext(context.Background(),
		"POST", "/process", strings.NewReader("_csrf="+existingCsrfToken.Value),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	mw(c)

	if len(c.Errors) != 0 {
		t.Errorf("there should be no error")
	}

	//RequestDetails with expected csrf token in header instead of form parameter
	c = webtest.NewGinContext(context.Background(),
		"POST", "/process", nil,
		webtest.Headers("X-CSRF-TOKEN", existingCsrfToken.Value),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	mw(c)

	if len(c.Errors) != 0 {
		t.Errorf("there should be no error")
	}

	//RequestDetails without a csrf token associated with it
	c = webtest.NewGinContext(context.Background(),
		"POST", "/process", nil,
		webtest.Headers("X-CSRF-TOKEN", uuid.New().String()),
	)
	s.Delete(SessionKeyCsrfToken) //remove the csrf token from the session
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}

	mockSessionStore.EXPECT().Save(gomock.Any()) //since this request's session doesn't have a csrf token, one will be generated

	mw(c)

	if len(c.Errors) != 1 {
		t.Errorf("there should be one error")
	}

	if !errors.Is(c.Errors.Last().Err, security.NewInvalidCsrfTokenError("")) {
		t.Errorf("expect invalid csrf token error, but was %v", c.Errors.Last())
	}
}

func TestCsrfMiddlewareProtectionAndIgnoreMatcher(t *testing.T) {
	csrfStore := newSessionBackedStore()

	protectionMatcher := matcher.RequestWithMethods("POST")
	ignoreMatcher := matcher.RequestWithPattern("/ignore")

	manager := newManager(csrfStore, protectionMatcher, ignoreMatcher)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := sessionmock.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, common.DefaultName)

	//RequestDetails with a invalid csrf token against the protected path
	c := webtest.NewGinContext(context.Background(),
		"POST", "/process", strings.NewReader("_csrf="+uuid.New().String()),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	existingCsrfToken := csrfStore.Generate(c, "_csrf", "X-CSRF-TOKEN")
	s.Set(SessionKeyCsrfToken, existingCsrfToken)

	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	mw := manager.CsrfHandlerFunc()
	mw(c)

	if len(c.Errors) != 1 {
		t.Errorf("there should be one error")
	}

	//RequestDetails with a invalid csrf token against the ignored path
	c = webtest.NewGinContext(context.Background(),
		"POST", "/ignore", strings.NewReader("_csrf="+uuid.New().String()),
		webtest.Headers("Content-Type", "application/x-www-form-urlencoded"),
	)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	mw(c)

	if len(c.Errors) != 0 {
		t.Errorf("there should be no error")
	}
}
