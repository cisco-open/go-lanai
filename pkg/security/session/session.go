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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

// Default flashes key.
const flashesKey = "_flash"
const createdTimeKey = "_created"

type Session struct {
	//Used to indicate if the session values has been modified - and should be saved
	dirty bool

	//Updated every time the session is accessed. Used to calculate timeout
	lastAccessed time.Time

	// The id of the session, generated by stores. It should not be used for
	// user data.
	id string
	// values contains the user-data for the session.
	// Because the value is declared as interface, any concrete type that is stored in the values map need to register with gob
	// if used with a store that serializes using gob. See NewRedisStore
	// Should only be set through setter and not set directly
	values map[interface{}]interface{}
	// Should only be modified when session is created.
	options *Options
	isNew   bool
	store   Store
	name    string

	originalAuth security.Authentication
}

type Options struct {
	Path   string
	Domain string

	// Determines how the cookie's "Max-Age" attribute will be set
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	// Defaults to http.SameSiteDefaultMode
	SameSite http.SameSite

	IdleTimeout time.Duration
	AbsoluteTimeout time.Duration
}

func NewSession(store Store, name string) *Session {
	opts := *store.Options() // a copy of the store's option

	return &Session{
		values:  make(map[interface{}]interface{}),
		store:   store,
		name:    name,
		options: &opts,
		isNew:   true,
		dirty:   false,
	}
}

func CreateSession(store Store, name string) *Session {
	s := NewSession(store, name)
	s.lastAccessed = time.Now()
	s.values[createdTimeKey] = time.Now()
	s.id = uuid.New().String()
	return s
}

// NewCookie returns an http.Cookie with the options set. It also sets
// the Expires field calculated based on the MaxAge value, for Internet
// Explorer compatibility.
func NewCookie(name, value string, options *Options, r *http.Request) *http.Cookie {
	cookie := newCookieFromOptions(name, value, options)

	if options.MaxAge > 0 {
		d := time.Duration(options.MaxAge) * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if options.MaxAge < 0 {
		// Set it to the past to expire now.
		cookie.Expires = time.Unix(1, 0)
	}

	protoHeader := r.Header.Get("X-Forwarded-Proto")

	if !options.Secure {
		cookie.Secure = strings.Contains(protoHeader, "https")
	}

	return cookie
}

// newCookieFromOptions returns an http.Cookie with the options set.
func newCookieFromOptions(name, value string, options *Options) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
	}

}

// GetID returns the name used to register the session.
func (s *Session) GetID() string {
	return s.id
}

// Name returns the name used to register the session.
func (s *Session) Name() string {
	return s.name
}

// Get returns the session value associated to the given key.
func (s *Session) Get(key interface{}) interface{} {
	return s.values[key]
}

// Set sets the session value associated to the given key.
func (s *Session) Set(key interface{}, val interface{}) {
	s.values[key] = val
	s.SetDirty()
}

// Delete removes the session value associated to the given key.
func (s *Session) Delete(key interface{}) {
	if _, ok := s.values[key]; ok {
		delete(s.values, key)
		s.SetDirty()
	}
}

// Clear deletes all values in the session.
func (s *Session) Clear() {
	s.values = make(map[interface{}]interface{})
	s.SetDirty()
}

// Flashes returns a slice of flash messages from the session.
//
// A single variadic argument is accepted, and it is optional: it defines
// the flash key. If not defined "_flash" is used by default.
func (s *Session) Flashes(flashKey ...string) []interface{} {
	defer s.SetDirty()

	var flashes []interface{}
	key := flashesKey
	if len(flashKey) > 0 {
		key = flashKey[0]
	}
	if v, ok := s.values[key]; ok {
		// Drop the flashes and return it.
		delete(s.values, key)
		flashes = v.([]interface{})
	}
	return flashes
}

// Flash get the last flash message of given key. It internally uses Flashes
func (s *Session) Flash(key string) (ret interface{}) {
	values := s.Flashes(key)
	if len(values) > 0 {
		ret = values[len(values) - 1]
	}
	return
}

// AddFlash adds a flash message to the session.
//
// A single variadic argument is accepted, and it is optional: it defines
// the flash key. If not defined "_flash" is used by default.
func (s *Session) AddFlash(value interface{}, flashKey ...string) {
	key := flashesKey
	if len(flashKey) > 0 {
		key = flashKey[0]
	}
	var flashes []interface{}
	if v, ok := s.values[key]; ok {
		flashes = v.([]interface{})
	}
	s.values[key] = append(flashes, value)
	s.SetDirty()
}

func (s *Session) ChangeId() error {
	return s.store.ChangeId(s)
}

// Save is a convenience method to save this session. It is the same as calling
// store.Save(request, response, session).
func (s *Session) Save() (err error) {
	if !s.dirty {
		return
	}

	err = s.store.Save(s)
	return
}

func (s *Session) IsDirty() bool {
	return s.dirty
}

func (s *Session) SetDirty()  {
	s.dirty = true
}

func (s *Session) ExpireNow(ctx context.Context) error {
	return s.store.WithContext(ctx).Invalidate(s)
}

func (s *Session) isExpired() bool {
	now := time.Now()
	canExpire, exp := s.expiration()

	if !canExpire {
		return false
	} else {
		return exp.Before(now)
	}
}

func (s *Session) createdOn() time.Time {
	if t, ok := s.values[createdTimeKey]; ok {
		return t.(time.Time)
	} else {
		return time.Time{}
	}
}

func (s *Session) expiration() (canExpire bool, expiration time.Time) {
	var timeoutSetting common.TimeoutSetting = 0

	var idleExpiration, absExpiration time.Time
	if s.options.IdleTimeout > 0 {
		idleExpiration = s.lastAccessed.Add(s.options.IdleTimeout)
		timeoutSetting = timeoutSetting | common.IdleTimeoutEnabled
	}
	if s.options.AbsoluteTimeout > 0 {
		absExpiration = s.createdOn().Add(s.options.AbsoluteTimeout)
		timeoutSetting = timeoutSetting | common.AbsoluteTimeoutEnabled
	}
	return common.CalculateExpiration(timeoutSetting, idleExpiration, absExpiration)
}