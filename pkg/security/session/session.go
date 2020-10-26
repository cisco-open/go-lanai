package session

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Manager struct {
	store sessions.Store
}

func NewManager() *Manager {
	store := cookie.NewStore()
	return &Manager{store: store}
}

func (m *Manager) SessionHandlerFunc() gin.HandlerFunc {
	return sessions.Sessions("Default", m.store)
}

func (m *Manager) SessionPostHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		if session.Get("TEST") == nil {
			session.Set("TEST", "VALUE")
			session.Save()
		} else {
			fmt.Printf("Have Session Value %s\n", session.Get("TEST"))
		}
		ctx.Next()
	}
}


