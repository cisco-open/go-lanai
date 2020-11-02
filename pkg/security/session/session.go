package session

import (
	"github.com/gorilla/sessions"
	"net/http"
)

type Store interface {
	sessions.Store
	Options(*sessions.Options) Store
}

type Session struct {
	session *sessions.Session
	request *http.Request
	writer  http.ResponseWriter
	dirty bool
}

// ID returns the name used to register the session.
func (s *Session) ID() string {
	return s.session.ID
}

// Name returns the name used to register the session.
func (s *Session) Name() string {
	return s.session.Name()
}

// Get returns the session value associated to the given key.
func (s *Session) Get(key interface{}) interface{} {
	return s.session.Values[key]
}

// Set sets the session value associated to the given key.
func (s *Session) Set(key interface{}, val interface{}) {
	s.session.Values[key] = val
	s.SetDirty()
}

// Delete removes the session value associated to the given key.
func (s *Session) Delete(key interface{}) {
	if _, ok := s.session.Values[key]; ok {
		delete(s.session.Values, key)
		s.SetDirty()
	}
}

// Clear deletes all values in the session.
func (s *Session) Clear() {
	s.session.Values = make(map[interface{}]interface{})
	s.SetDirty()
}

// AddFlash adds a flash message to the session.
// A single variadic argument is accepted, and it is optional: it defines the flash key.
// If not defined "_flash" is used by default.
func (s *Session) AddFlash(value interface{}, vars ...string) {
	s.session.AddFlash(value, vars...)
	s.SetDirty()
}

// Flashes returns a slice of flash messages from the session.
// A single variadic argument is accepted, and it is optional: it defines the flash key.
// If not defined "_flash" is used by default.
func (s *Session) Flashes(vars ...string) []interface{} {
	defer s.SetDirty()
	return s.session.Flashes(vars...)
}

// Options sets configuration for a session.
func (s *Session) Options(options *sessions.Options) {
	s.session.Options = options
}

// Save saves all sessions used during the current request.
func (s *Session) Save() (err error) {
	if !s.dirty {
		return
	}

	err = s.session.Save(s.request, s.writer)
	if err == nil {
		s.dirty = false
	}
	return
}

func (s *Session) IsDirty() bool {
	return s.dirty
}

func (s *Session) SetDirty()  {
	s.dirty = true
}