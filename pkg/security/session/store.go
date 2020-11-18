package session

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"encoding/base32"
	"encoding/gob"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const sessionValueField = "values"
const sessionOptionField = "options"
const sessionLastAccessedField = "lastAccessed"

type Store interface {
	// Get should return a cached session.
	Get(id string, name string) (*Session, error)

	// New should create and return a new session.
	//
	// Note that New should never return a nil session, even in the case of
	// an error if using the Registry infrastructure to cache the session.
	New(name string) (*Session, error)

	// Save should persist session to the underlying store implementation.
	Save(s *Session) error

	//Delete the session from store. It will also update the http response header to clear the corresponding cookie
	Delete(s *Session) error

	Options() *Options
}

type RedisStore struct {
	options    *Options
	connection *redis.Connection
}

type notFoundError struct{
	id string
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("redis store: session not found %s", e.id)
}

func NewRedisStore(connection *redis.Connection, options ...func(*Options)) Store {
	//defaults
	o := &Options{
		Path:   "/",
		MaxAge: 86400 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
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
	return s.options
}

func (s *RedisStore) Get(id string, name string) (*Session, error) {
	if id != "" {
		session := NewSession(s, name)
		session.ID = id
		err := s.load(session) //TODO: should session expired be handled here or at higher level?
		if err == nil {
			session.IsNew = false
		}

		if _, ok := err.(*notFoundError); ok {
			return session, nil
		} else {
			return session, err
		}
	} else {
		return s.New(name)
	}
}

// New will create a new session.
func (s *RedisStore) New(name string) (*Session, error) {
	session := NewSession(s, name)
	random := securecookie.GenerateRandomKey(32)

	if random == nil {
		return nil, errors.New("system failed to generate random value for session id")
	}

	session.ID = strings.TrimRight(
		base32.StdEncoding.EncodeToString(random), "=")

	return session, nil
}

// Save adds a single session to the response.
func (s *RedisStore) Save(session *Session) error {
	if session.ID == "" {
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
	return nil
}

func (s *RedisStore) load(session *Session) error {
	key := fmt.Sprintf("%s:%s", session.Name(), session.ID)

	cmd := s.connection.HGetAll(context.Background(), key)

	result, err := cmd.Result()

	if err != nil {
		return err
	}

	if len(result) == 0 {
		return &notFoundError{id: session.ID}
	}

	for k, v := range result {
		if k == sessionValueField {
			err = Deserialize(strings.NewReader(v), &session.Values)
		} else if k == sessionOptionField {
			err = Deserialize(strings.NewReader(v), &session.Options)
		} else if k == sessionLastAccessedField {
			timeStamp, e := strconv.ParseInt(v, 10, 0)
			session.lastAccessed = time.Unix(timeStamp, 0)
			err = e
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RedisStore) save(session *Session) error {
	//TODO: calculate ttl

	key := fmt.Sprintf("%s:%s", session.Name(), session.ID)
	var args []interface{}

	if values, err := Serialize(session.Values); err == nil {
		args = append(args, sessionValueField, values)
	}

	if options, err := Serialize(session.Options); err == nil {
		args = append(args, sessionOptionField, options)
	}

	args = append(args, sessionLastAccessedField, session.lastAccessed.Unix())

	cmd := s.connection.HSet(context.Background(), key, args...)

	return cmd.Err()
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

