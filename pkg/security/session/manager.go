package session

import (
	"context"
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"fmt"
	"github.com/gin-gonic/gin"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"net/http"
	"path"
)

const (
	ContextKeySession = "kSession"
	DefaultName = "SESSION"

	sessionKeySecurity = "kSecurity"
)

func Get(c context.Context) *Session {
	switch s := c.Value(ContextKeySession); s.(type) {
	case *Session:
		return s.(*Session)
	default:
		return nil
	}
}

type Manager struct {
	store Store
}

func NewManager(store Store, sessionProps security.SessionProperties, serverProps web.ServerProperties) *Manager {
	options := &sessions.Options{
		Path: path.Clean("/" + serverProps.ContextPath),
		Domain: sessionProps.Cookie.Domain,
		MaxAge: 0,
		Secure: false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	return &Manager{store: store.Options(options)}
}

// SessionHandlerFunc provide middleware for basic session management
func (m *Manager) SessionHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO better error handling
		// defer is FILO
		defer gcontext.Clear(c.Request)
		defer m.saveSession(c)
		defer c.Next()

		existing, err := m.store.Get(c.Request, DefaultName)
		if err != nil {
			_ = c.Error(err)
			return
		}

		// TODO validate session

		// TODO logger
		if existing != nil && existing.IsNew {
			fmt.Printf("New Session %v\n", existing.ID)
		}

		m.registerSession(c, existing)
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
		current := m.getCurrent(c)
		if current == nil {
			// no session found in current context, do nothing
			return
		}

		if auth, ok := current.Get(sessionKeySecurity).(security.Authentication); ok {
			c.Set(security.ContextKeySecurity, auth)
		}
	}
}


func (m *Manager) getCurrent(c *gin.Context) *Session {
	switch i,_ := c.Get(ContextKeySession); i.(type) {
	case *Session:
		return i.(*Session)
	default:
		return nil
	}
}

func (m *Manager) registerSession(c *gin.Context, s *sessions.Session) {
	session := &Session{
		session: s,
		request: c.Request,
		writer:  c.Writer,
	}
	c.Set(ContextKeySession, session)
}

func (m *Manager) saveSession(c *gin.Context) {
	session := m.getCurrent(c)
	if session == nil {
		return
	}

	if err := session.Save(); err != nil {
		_ = c.Error(err)
	}
}

func (m *Manager) persistAuthentication(c *gin.Context) {
	session := m.getCurrent(c)
	if session == nil {
		return
	}

	auth := security.Get(c)
	session.Set(sessionKeySecurity, auth)

	if err := session.Save(); err != nil {
		_ = c.Error(err)
	}
}



