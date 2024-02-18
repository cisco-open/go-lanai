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

package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	sessionKeySecurity = "Security"
)

type sessionCtxKey struct {}
var	contextKeySession  = sessionCtxKey{}

// Get returns Session stored in given context. May return nil
func Get(c context.Context) *Session {
	session, _ := c.Value(contextKeySession).(*Session)
	return session
}

// MustSet is the panicking version of Set
func MustSet(c context.Context, s *Session) {
	if e := Set(c, s); e != nil {
		panic(e)
	}
}

// Set given Session into given context. The function returns error if the given context is not backed by utils.MutableContext.
func Set(c context.Context, s *Session) error {
	mc := utils.FindMutableContext(c)
	if mc == nil {
		return security.NewInternalError(fmt.Sprintf(`unable to set session into context: given context [%T] is not mutable`, c))
	}
	mc.Set(contextKeySession, s)
	return nil
}

type Manager struct {
	name  string
	store Store
}

func NewManager(sessionName string, store Store) *Manager {
	return &Manager{
		name:  sessionName,
		store: store,
	}
}

// SessionHandlerFunc provide middleware for basic session management
func (m *Manager) SessionHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer is FILO
		defer m.saveSession(c)
		defer c.Next()

		var id string

		if cookie, err := c.Request.Cookie(m.name); err == nil {
			id = cookie.Value
		}

		session, err := m.store.WithContext(c).Get(id, m.name)
		// If session store is not operating properly, we cannot continue for misc that needs session
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if session != nil && session.isNew {
			logger.WithContext(c).Debugf("New Session %s", session.id)
			http.SetCookie(c.Writer, NewCookie(session.Name(), session.id, session.options, c.Request))
		}

		if e := Set(c, session); e != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
}

// AuthenticationPersistenceHandlerFunc provide middleware to load security from session and save it at end
func (m *Manager) AuthenticationPersistenceHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer is FILO
		defer m.persistAuthentication(c)
		defer c.Next()

		// load security from session
		current := Get(c)
		if current == nil {
			// no session found in current ctx, do nothing
			return
		}

		if auth, ok := current.Get(sessionKeySecurity).(security.Authentication); ok {
			security.MustSet(c, auth)
		} else {
			security.MustSet(c, nil)
		}
	}
}

func (m *Manager) saveSession(c *gin.Context) {
	session := Get(c)
	if session == nil {
		return
	}

	err := m.store.WithContext(c).Save(session)

	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (m *Manager) persistAuthentication(c *gin.Context) {
	session := Get(c)
	if session == nil {
		return
	}

	auth := security.Get(c)
	session.Set(sessionKeySecurity, auth)
}
