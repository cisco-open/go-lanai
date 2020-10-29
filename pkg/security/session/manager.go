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

// Session management for HTTP requests
func (m *Manager) SessionHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer is FILO
		defer c.Next()
		defer gcontext.Clear(c.Request)

		s, err := m.store.Get(c.Request, DefaultName)
		if err != nil {
			_ = c.Error(err)
		}

		// TODO logger
		if s != nil && s.IsNew {
			fmt.Printf("New Session %v\n", s.ID)
		}

		session := &Session{
			session: s,
			request: c.Request,
			writer: c.Writer,
		}
		c.Set(ContextKeySession, session)
	}
}




