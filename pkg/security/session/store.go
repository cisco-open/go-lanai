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
    "bytes"
    "context"
    "encoding/gob"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/session/common"
    "github.com/google/uuid"
    "github.com/pkg/errors"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"
)

const (
	sessionValueField  = "values"
	sessionOptionField = "options"
)

const (
	globalSettingIdleTimeout = "IDLE_SESSION_TIMEOUT_SECS"
	globalSettingAbsTimeout  = "ABSOLUTE_SESSION_TIMEOUT_SECS"
)

type Store interface {
	// Get should return a cached session.
	Get(id string, name string) (*Session, error)

	// New should create and return a new session.
	New(name string) (*Session, error)

	// Save should persist session to the underlying store implementation.
	Save(s *Session) error

	// Invalidate sessions from store.
	// It will also remove associations between sessions and its stored principal via RemoveFromPrincipalIndex
	Invalidate(sessions ...*Session) error

	Options() *Options

	ChangeId(s *Session) error

	AddToPrincipalIndex(principal string, session *Session) error

	RemoveFromPrincipalIndex(principal string, sessions *Session) error

	FindByPrincipalName(principal string, sessionName string) ([]*Session, error)

	// InvalidateByPrincipalName invalidate all sessions associated with given principal name
	InvalidateByPrincipalName(principal, sessionName string) error

	// WithContext make a shallow copy of the store with given ctx
	WithContext(ctx context.Context) Store
}

// RedisStore
/**
	Session is implemented as a HSET in redis.
	Session is expired using Redis TTL. The TTL is slightly longer than the expiration time so that when the session is
	expired, we can still get the session details (if necessary).

	Currently we don't have a need for "on session expired" event. If we do have a need for this in the future,
	we can use https://redis.io/topics/notifications and listen on the TTL event. Once caveat of using the redis
	notification is that it may not generate a event until the key is being accessed. So if we want to have deterministic
	behaviour on when the event is fired, we would need to implement a scheduler ourselves that access these keys that
	are expired which will force redis to generate the event.

	For each session:
	1. HSET with key in the form of "LANAI:SESSION:SESSION:{sessionId}"
	This stores session.values, session.options and session.lastAccessedTime as separate fields in the hash set

	2. SET with key in the form of "LANAI:SESSION:INDEX:SESSION:{principal}"
	This stores the set of session Id for this user. The session Id stored in this set may have been expired or deleted.

	if we don't clean up the keys in this set, then that means on each successful login, we need to go through the
	content of this set and find the corresponding session - sscan for the set entries, and then hgetall for each entry.
	and filter the expired entries.

	if we want to clean up the keys in this set, we need to do so with scheduled tasks. We cannot depend on the redis expiring
	event, because when we get the event, the session is not available anymore. Therefore we need to introduce other data structures
	to keep track of the expiration separately.

	The worst case scenario if we don't clean up this set
	is when a user opens multiple session without logging out - this will result in these sessions remain in the set even
	when they expires. This will result in a penalty when the user logs on the next time.
	But if we don't expect a user to have millions of concurrent sessions, this should be insignificant.
	If concurrent user limit is set, we don't expect the number of entries to be more than the concurrent user limit
    which should be reasonably small.

	If concurrent user limit is not set, it can grow large, but that is not a problem due to expiration - i.e. the
	set can grow unbounded before any expiration event occurs. And the remedy to that is to apply a concurrent user limit.

	The application can use redis SCAN family of commands to make sure that redis is not blocked by a single user's request.
*/
type RedisStore struct {
	ctx           context.Context
	options       *Options
	connection    redis.Client
	settingReader security.GlobalSettingReader
}

type StoreOptions func(opt *StoreOption)

type StoreOption struct {
	Options
	SettingReader security.GlobalSettingReader
}

func NewRedisStore(redisClient redis.Client, options ...StoreOptions) *RedisStore {
	gob.Register(time.Time{})

	//defaults
	opt := StoreOption{
		Options: Options{
			Path:            "/",
			HttpOnly:        true,
			SameSite:        http.SameSiteNoneMode,
			IdleTimeout:     900 * time.Second,
			AbsoluteTimeout: 1800 * time.Second,
		},
	}

	for _, fn := range options {
		fn(&opt)
	}
	return &RedisStore{
		ctx:           context.Background(),
		options:       &opt.Options,
		connection:    redisClient,
		settingReader: opt.SettingReader,
	}
}

func (s *RedisStore) WithContext(ctx context.Context) Store {
	cp := *s
	cp.ctx = ctx
	return &cp
}

func (s *RedisStore) Options() *Options {
	return s.options
}

func (s *RedisStore) Get(id string, name string) (*Session, error) {
	if id != "" {
		session, err := s.load(id, name)

		if err != nil {
			return nil, err
		}

		if session == nil {
			return s.New(name)
		} else {
			return session, nil
		}
	} else {
		return s.New(name)
	}
}

// New will create a new session.
func (s *RedisStore) New(name string) (*Session, error) {
	session := CreateSession(s, name)
	if idle, ok := s.readTimeoutSetting(s.ctx, globalSettingIdleTimeout); ok {
		session.options.IdleTimeout = idle
	}
	if abs, ok := s.readTimeoutSetting(s.ctx, globalSettingAbsTimeout); ok {
		session.options.AbsoluteTimeout = abs
	}
	return session, nil
}

// Save adds a single session to the persistence layer
func (s *RedisStore) Save(session *Session) error {
	if session.id == "" {
		return errors.New("session id is empty")
	}

	session.lastAccessed = time.Now()
	err := s.save(session)
	if err == nil {
		session.dirty = false
		session.isNew = false
	}
	return err
}

func (s *RedisStore) Invalidate(sessions ...*Session) error {
	for _, session := range sessions {
		if cmd := s.connection.Del(s.ctx, common.GetRedisSessionKey(session.Name(), session.GetID())); cmd.Err() != nil {
			return cmd.Err()
		}

		// remove principal index is an optional step
		if pName, e := getPrincipalName(session); e == nil && pName != "" {
			//ignore error here since even if it can't be deleted from this index, it'll be cleaned up
			// on read since the session itself is already deleted successfully
			_ = s.RemoveFromPrincipalIndex(pName, session)
		}
	}
	return nil
}

func (s *RedisStore) InvalidateByPrincipalName(principal, sessionName string) error {
	sessions, e := s.FindByPrincipalName(principal, sessionName)
	if e != nil {
		return e
	}
	return s.Invalidate(sessions...)
}

func (s *RedisStore) FindByPrincipalName(principal string, sessionName string) ([]*Session, error) {
	//iterate through the set members using default count
	cursor := uint64(0)
	var ids []string
	for ok := true; ok; ok = cursor != 0 {
		cmd := s.connection.SScan(s.ctx, getRedisPrincipalIndexKey(principal, sessionName), cursor, "", 0)
		keys, next, err := cmd.Result()
		cursor = next
		if err != nil {
			return nil, err
		}
		ids = append(ids, keys...)
	}

	var found []*Session
	var expired []interface{}
	for _, id := range ids {
		session, err := s.load(id, sessionName)

		if err != nil {
			return nil, err
		}

		if session == nil {
			expired = append(expired, id)
		} else {
			found = append(found, session)
		}
	}

	//clean up the expired entries from the index
	if len(expired) > 0 {
		s.connection.SRem(s.ctx, getRedisPrincipalIndexKey(principal, sessionName), expired...)
	}

	return found, nil
}

func (s *RedisStore) AddToPrincipalIndex(principal string, session *Session) error {
	cmd := s.connection.SAdd(s.ctx, getRedisPrincipalIndexKey(principal, session.Name()), session.GetID())
	return cmd.Err()
}

func (s *RedisStore) RemoveFromPrincipalIndex(principal string, session *Session) error {
	cmd := s.connection.SRem(s.ctx, getRedisPrincipalIndexKey(principal, session.Name()), session.GetID())
	return cmd.Err()
}

func (s *RedisStore) ChangeId(session *Session) error {
	newId := uuid.New().String()
	cmd := s.connection.Rename(s.ctx, common.GetRedisSessionKey(session.Name(), session.GetID()), common.GetRedisSessionKey(session.Name(), newId))
	err := cmd.Err()
	if err != nil {
		return err
	}
	session.id = newId
	return nil
}

func (s *RedisStore) load(id string, name string) (*Session, error) {
	key := common.GetRedisSessionKey(name, id)

	cmd := s.connection.HGetAll(s.ctx, key)

	result, err := cmd.Result()

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	session := NewSession(s, name)
	session.id = id

	for k, v := range result {
		if k == sessionValueField {
			err = Deserialize(strings.NewReader(v), &session.values)
		} else if k == sessionOptionField {
			err = Deserialize(strings.NewReader(v), &session.options)
		} else if k == common.SessionLastAccessedField {
			timeStamp, e := strconv.ParseInt(v, 10, 0)
			session.lastAccessed = time.Unix(timeStamp, 0)
			err = e
		}

		if err != nil {
			return nil, err
		}
	}
	session.isNew = false

	if session.isExpired() {
		return nil, nil
	} else {
		return session, nil
	}
}

func (s *RedisStore) save(session *Session) error {
	key := common.GetRedisSessionKey(session.Name(), session.GetID())
	var args []interface{}

	if session.IsDirty() || session.isNew {
		if values, err := Serialize(session.values); err == nil {
			args = append(args, sessionValueField, values)
		} else {
			return err
		}
	}

	if session.isNew {
		if options, err := Serialize(session.options); err == nil {
			args = append(args, sessionOptionField, options)

			//stored separate for easy retrieval
			if session.options.IdleTimeout > 0 {
				args = append(args, common.SessionIdleTimeoutDuration, session.options.IdleTimeout.String())
			}
			if session.options.AbsoluteTimeout > 0 {
				args = append(args, common.SessionAbsTimeoutTime, session.createdOn().Add(session.options.AbsoluteTimeout).Unix())
			}
		} else {
			return err
		}
	}

	args = append(args, common.SessionLastAccessedField, session.lastAccessed.Unix())
	hsetCmd := s.connection.HSet(s.ctx, key, args...)
	if hsetCmd.Err() != nil {
		return hsetCmd.Err()
	}

	canExpire, exp := session.expiration()
	if canExpire {
		expCmd := s.connection.ExpireAt(s.ctx, key, exp)
		return expCmd.Err()
	}

	return nil
}

func (s *RedisStore) readTimeoutSetting(ctx context.Context, key string) (time.Duration, bool) {
	if s.settingReader == nil {
		return  0, false
	}
	var secs int
	if e := s.settingReader.Read(ctx, key, &secs); e != nil {
		return 0, false
	}
	return time.Second * time.Duration(secs), true
}

func getRedisPrincipalIndexKey(principal string, sessionName string) string {
	return fmt.Sprintf("%s:INDEX:%s:%s", common.RedisNameSpace, sessionName, principal)
}

func Serialize(src interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(src); err != nil {
		return nil, errors.Wrap(err, "Cannot serialize value")
	}
	return buf.Bytes(), nil
}

func Deserialize(src io.Reader, dst interface{}) error {
	dec := gob.NewDecoder(src)
	if err := dec.Decode(dst); err != nil {
		return errors.Wrap(err, "Cannot serialize value")
	}
	return nil
}

func getPrincipalName(session *Session) (string, error) {
	auth, ok := session.Get(sessionKeySecurity).(security.Authentication)
	if !ok {
		return "", nil
	}

	return security.GetUsername(auth)
}
