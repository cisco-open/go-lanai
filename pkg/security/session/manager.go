package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	DefaultName = "SESSION"
	sessionKeySecurity = "kSecurity"
	contextKeySession = web.ContextKeySession
)

func Get(c context.Context) *Session {
	var session *Session
	switch c.(type) {
	case *gin.Context:
		i,_ := c.(*gin.Context).Get(contextKeySession)
		session,_ = i.(*Session)
	default:
		session,_ = c.Value(contextKeySession).(*Session)
	}
	return session
}

type Manager struct {
	store Store
}

func NewManager(store Store) *Manager {
	return &Manager{
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

		if cookie, err := c.Request.Cookie(DefaultName); err == nil {
			id = cookie.Value
		}

		session, err := m.store.Get(id, DefaultName)
		// If session store is not operating properly, we cannot continue for endpoints that needs session
		if err != nil {
			_ = c.Error(err)
			return
		}

		// TODO validate session

		// TODO logger
		if session != nil && session.isNew {
			fmt.Printf("New Session %v\n", session.id)
			http.SetCookie(c.Writer, NewCookie(session.Name(), session.id, session.options))
		}

		m.registerSession(c, session)
	}
}

// AuthenticationPersistenceHandlerFunc provide middleware to load security from session and save it at end
func (m *Manager) AuthenticationPersistenceHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO better error handling
		// defer is FILO
		defer m.persistAuthentication(c)
		defer c.Next()

		// load security from session
		current := Get(c)
		if current == nil {
			// no session found in current context, do nothing
			return
		}

		if auth, ok := current.Get(sessionKeySecurity).(security.Authentication); ok {
			c.Set(security.ContextKeySecurity, auth)
		} else {
			c.Set(security.ContextKeySecurity, nil)
		}
	}
}

func (m *Manager) registerSession(c *gin.Context, s *Session) {
	c.Set(contextKeySession, s)
}

func (m *Manager) saveSession(c *gin.Context) {
	session := Get(c)
	if session == nil {
		return
	}

	err := m.store.Save(session)

	if err != nil {
		_ = c.Error(err)
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