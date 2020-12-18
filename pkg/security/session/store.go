package session

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const redisNameSpace = "LANAI:SESSION" //This is to avoid confusion with records from other frameworks.
const sessionValueField = "values"
const sessionOptionField = "options"
const sessionLastAccessedField = "lastAccessed"

type Store interface {
	// Get should return a cached session.
	Get(id string, name string) (*Session, error)

	// New should create and return a new session.
	New(name string) (*Session, error)

	// Save should persist session to the underlying store implementation.
	Save(s *Session) error

	//Delete the session from store. It will also update the http response header to clear the corresponding cookie
	Delete(s *Session) error

	Options() *Options

	ChangeId(s *Session) error

	AddToPrincipalIndex(principal string, session *Session) error
	RemoveFromPrincipalIndex(principal string, sessions *Session) error
	FindByPrincipalName(principal string, sessionName string) ([]*Session, error)
}

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
	options    *Options
	connection redis.Client
}

func NewRedisStore(connection redis.Client, options ...func(*Options)) *RedisStore {
	gob.Register(time.Time{})

	//defaults
	o := &Options{
		Path:   "/",
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
		IdleTimeout: 900*time.Second,
		AbsoluteTimeout: 1800*time.Second,
	}

	for _, opt := range options {
		opt(o)
	}

	s := &RedisStore{
		options: o,
		connection: connection,
	}
	return s
}

func (s *RedisStore) Options() *Options {
	//TODO: the timeout values should be dynamically read every time to allow change
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
	session := NewSession(s, name)

	session.lastAccessed = time.Now()
	session.values[createdTimeKey] = time.Now()

	session.id = uuid.New().String()

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
	}
	return err
}

func (s *RedisStore) Delete(session *Session) error {
	cmd := s.connection.Del(context.Background(), getRedisSessionKey(session.Name(), session.GetID()))
	return cmd.Err()
}

func (s *RedisStore) FindByPrincipalName(principal string, sessionName string) ([]*Session, error) {
	//iterate through the set members using default count
	cursor:= uint64(0)
	var ids []string
	for ok := true; ok ; ok = cursor!= 0 {
		cmd := s.connection.SScan(context.Background(), getRedisPrincipalIndexKey(principal, sessionName), cursor, "", 0)
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
	s.connection.SRem(context.Background(), getRedisPrincipalIndexKey(principal, sessionName), expired...)

	return found, nil
}

func (s *RedisStore) AddToPrincipalIndex(principal string, session *Session) error {
	cmd := s.connection.SAdd(context.Background(), getRedisPrincipalIndexKey(principal, session.Name()), session.GetID())
	return cmd.Err()
}

func (s *RedisStore) RemoveFromPrincipalIndex(principal string, session *Session) error {
	cmd := s.connection.SRem(context.Background(), getRedisPrincipalIndexKey(principal, session.Name()), session.GetID())
	return cmd.Err()
}


func (s *RedisStore) ChangeId(session *Session) error {
	newId := uuid.New().String()
	cmd := s.connection.Rename(context.Background(), getRedisSessionKey(session.Name(), session.GetID()), getRedisSessionKey(session.Name(), newId))
	err := cmd.Err()
	if err != nil {
		return err
	}
	session.id = newId
	return nil
}

func (s *RedisStore) load(id string, name string) (*Session, error) {
	key := getRedisSessionKey(name, id)

	cmd := s.connection.HGetAll(context.Background(), key)

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
		} else if k == sessionLastAccessedField {
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
	key := getRedisSessionKey(session.Name(), session.GetID())
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
		} else {
			return err
		}
	}

	args = append(args, sessionLastAccessedField, session.lastAccessed.Unix())
	hsetCmd := s.connection.HSet(context.Background(), key, args...)
	if hsetCmd.Err() != nil {
		return hsetCmd.Err()
	}

	exp := session.expiration()
	expCmd := s.connection.ExpireAt(context.Background(), key, exp)
	return expCmd.Err()
}

func getRedisSessionKey(name string, id string) string {
	return fmt.Sprintf("%s:%s:%s", redisNameSpace, name, id)
}

func getRedisPrincipalIndexKey(principal string, sessionName string) string {
	return fmt.Sprintf("%s:INDEX:%s:%s", redisNameSpace, sessionName, principal)
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

