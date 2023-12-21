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

package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	sessioncommon "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

/*******************
	Options
 *******************/

func SessionID(sessionId string) webtest.RequestOptions {
	return func(req *http.Request) {
		cookie := http.Cookie{
			Name:       sessioncommon.DefaultName,
			Value:      sessionId,
		}
		req.Header.Set("Cookie", cookie.String())
	}
}

/*******************
	Mocks
 *******************/
var sessionKeyPrincipal = struct{}{}

type MockedSessionStore struct{
	Sessions map[string]*session.Session
}

func NewMockedSessionStore() session.Store {
	return &MockedSessionStore{
		Sessions: map[string]*session.Session{},
	}
}

func MockedSessionStoreDecorator(_ session.Store) session.Store {
	return NewMockedSessionStore()
}

func (ss *MockedSessionStore) Get(id string, name string) (s *session.Session, err error) {
	if id == "" {
		return ss.New(name)
	}

	if s, _ = ss.Sessions[ss.toKey(id, name)]; s == nil {
		s, _ = ss.New(name)
	}
	return
}

func (ss *MockedSessionStore) New(name string) (*session.Session, error) {
	s := session.CreateSession(ss, name)
	if s != nil {
		ss.Sessions[ss.key(s)] = s
	}
	return s, nil
}

func (ss *MockedSessionStore) Save(s *session.Session) error {
	if s != nil {
		ss.Sessions[ss.key(s)] = s
	}
	return nil
}

func (ss *MockedSessionStore) Invalidate(sessions ...*session.Session) error {
	for _, s := range sessions {
		delete(ss.Sessions, ss.key(s))
	}
	return nil
}

func (ss *MockedSessionStore) Options() *session.Options {
	return &session.Options{
		Path:            "/",
		Domain:          "localhost",
	}
}

func (ss *MockedSessionStore) ChangeId(s *session.Session) error {
	if s != nil {
		// Note: we can't actually change ID because session's id is a private field
		newId := uuid.New().String()
		delete(ss.Sessions, ss.key(s))
		ss.Sessions[newId] = s
	}
	return nil
}

func (ss *MockedSessionStore) AddToPrincipalIndex(principal string, s *session.Session) error {
	if s != nil {
		s.Set(sessionKeyPrincipal, principal)
	}
	return nil
}

func (ss *MockedSessionStore) RemoveFromPrincipalIndex(_ string, s *session.Session) error {
	if s != nil {
		s.Delete(sessionKeyPrincipal)
	}
	return nil
}

func (ss *MockedSessionStore) FindByPrincipalName(principal string, sessionName string) ([]*session.Session, error) {
	//iterate through the set members using default count
	var found []*session.Session
	for _, s := range ss.Sessions {
		if p, ok := s.Get(sessionKeyPrincipal).(string); ok && p == principal && s.Name() == sessionName {
			found = append(found, s)
		}
	}
	return found, nil
}

func (ss *MockedSessionStore) InvalidateByPrincipalName(principal, sessionName string) error {
	sessions, e := ss.FindByPrincipalName(principal, sessionName)
	if e != nil {
		return e
	}
	return ss.Invalidate(sessions...)
}

func (ss *MockedSessionStore) WithContext(_ context.Context) session.Store {
	return ss
}

func (ss *MockedSessionStore) toKey(id, name string) string {
	return fmt.Sprintf("session-%ss-%ss", id, name)
}

func (ss *MockedSessionStore) key(sess *session.Session) string {
	return fmt.Sprintf("session-%ss-%ss", sess.GetID(), sess.Name())
}





