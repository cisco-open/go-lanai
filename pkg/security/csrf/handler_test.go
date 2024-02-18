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
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/authmock"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/sessionmock"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"testing"
)

func TestChangeCsrfHanlderShouldChangeCSRFTokenWhenAuthenticated(t *testing.T) {
	csrfStore := newSessionBackedStore()
	handler := &ChangeCsrfHandler{
		csrfStore,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := sessionmock.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, common.DefaultName)

	//The request itself is not important
	c := webtest.NewGinContext(context.Background(), "GET", "/something", nil)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	token := &Token{
		uuid.New().String(),
		security.CsrfParamName,
		security.CsrfHeaderName,
	}
	s.Set(SessionKeyCsrfToken, token)

	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)

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
		if savedCsrfToken.(*Token).Value == token.Value {
			t.Errorf("Expect csrf token value should change")
		}
	})

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)
}

func TestChangeCsrfHanlderShouldNotChangeCSRFTokenIfNotAuthenticated(t *testing.T) {
	csrfStore := newSessionBackedStore()
	handler := &ChangeCsrfHandler{
		csrfStore,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := sessionmock.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, common.DefaultName)

	//The request itself is not important
	c := webtest.NewGinContext(context.Background(), "GET", "/something", nil)
	if e := session.Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	token := &Token{
		uuid.New().String(),
		security.CsrfParamName,
		security.CsrfHeaderName,
	}
	s.Set(SessionKeyCsrfToken, token)

	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAuthenticated)

	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAnonymous)

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)

	actualToken := s.Get(SessionKeyCsrfToken).(*Token)
	if actualToken.Value != token.Value {
		t.Errorf("csrf token should be unchanged")
	}
}