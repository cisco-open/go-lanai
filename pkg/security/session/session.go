package session

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

type Manager struct {
	store sessions.Store
}

func NewManager() *Manager {
	store := memstore.NewStore()
	return &Manager{store: store}
}

func (m *Manager) SessionHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//TODO
		fmt.Println("TODO load or create Session")
	}
}
