package store

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/security/session"
	"github.com/gorilla/sessions"
	"github.com/quasoft/memstore"
)

type MemoryStore struct {
	*memstore.MemStore
}

func (s *MemoryStore) Options(opt *sessions.Options) session.Store {
	s.MemStore.Options = opt
	return s
}

func NewMemoryStore(keyPairs ...[]byte) session.Store {
	return &MemoryStore{memstore.NewMemStore(keyPairs...)}
}