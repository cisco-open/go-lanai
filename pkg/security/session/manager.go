package session

import (
	"context"
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"fmt"
	"github.com/gorilla/sessions"
	gcontext "github.com/gorilla/context"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"path"
	"time"
)

const (
	ContextKeySession = "kSession"
	DefaultName = "SESSION"
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

func NewManager(ctx *bootstrap.ApplicationContext, store Store) *Manager {
	options := &sessions.Options{
		Path: path.Clean("/" + ctx.Value(web.PropertyServerContextPath).(string)),
		MaxAge: 0,
		Secure: false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	// TODO better keyPairs for sessions
	//store := sessions.NewFilesystemStore("session/", []byte("secret"))

	return &Manager{store: store.Options(options)}
}

// Provide session
func (m *Manager) SessionHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer is FILO
		defer c.Next()
		defer gcontext.Clear(c.Request)

		s, err := m.store.Get(c.Request, DefaultName)
		if err != nil {
			_ = c.Error(err)
		}
		session := &Session{
			session: s,
			request: c.Request,
			writer: c.Writer,
		}
		c.Set(ContextKeySession, session)
	}
}

// Session management for HTTP requests
func (m *Manager) SessionTestHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := Get(ctx)
		if session.Get("TEST") == nil {

			session.Set("TEST", RandomString(10240))
			err := session.Save()
			if err != nil {
				fmt.Printf("ERROR when saving session: %v\n", err)
			}
		} else {
			fmt.Printf("Have Session Value %s\n", "TEST")
		}
		ctx.Next()
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}


