package session

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"encoding/gob"
	"fmt"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

type testUser struct {
	User string
	Pass string
}

func (u *testUser) Username() string {
	return u.User
}

func (u *testUser) Password() string {
	return u.Pass
}

type testAuthentication struct {
	Account     security.Account
	PermissionList []string
}

func (auth *testAuthentication) Principal() interface{} {
	return auth.Account
}

func (auth *testAuthentication) Permissions() []string {
	return auth.PermissionList
}

func (auth *testAuthentication) Authenticated() bool {
	return true
}

func (auth *testAuthentication) Details() interface{} {
	return auth.Account
}

func (auth *testAuthentication) IsUsernamePasswordAuthentication() bool {
	return true
}

func TestSerialization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mocks.NewMockUniversalClient(ctrl)
	connection := &redis.Connection{
		UniversalClient: mock,
	}

	store := NewRedisStore(connection)
	s := NewSession(store, DefaultName)

	auth := testAuthentication{
		Account: &testUser{
			User: "test_user",
			Pass: "test_pass",
		},
		PermissionList: []string{"perm_a", "perm_b"},
	}

	s.Values["auth"] = auth

	gob.Register((*testAuthentication)(nil))
	gob.Register((*testUser)(nil))

	serialized, err := Serialize(s.Values)

	if err != nil {
		t.Errorf("Cannot serialize %v", err)
	}

	s = NewSession(store, DefaultName)
	err = Deserialize(bytes.NewReader(serialized), &s.Values)

	if err != nil {
		t.Errorf("Cannot deserialize %v", err)
	}

	deserializedAuth, ok := s.Values["auth"]

	if !ok {
		t.Errorf("auth is not deserialized")
	}

	userAuth, ok := deserializedAuth.(passwd.UsernamePasswordAuthentication)

	if !ok {
		t.Errorf("auth is not UsernamePasswordAuthentication")
	}

	principal := userAuth.Principal()

	account, ok := principal.(security.Account)

	if !ok {
		t.Errorf("principal is not account")
	}

	if account.Password() != "test_pass" || account.Username() != "test_user" {
		t.Errorf("account username password does not match")
	}

	if len(userAuth.Permissions()) != 2 {
		t.Errorf("permissions does not have the right length")
	}
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mocks.NewMockUniversalClient(ctrl)
	connection := &redis.Connection{
		UniversalClient: mock,
	}

	mock.EXPECT().
		HGetAll(gomock.Any(), gomock.Eq("session_name:session_id")).
		Return(&goRedis.StringStringMapCmd{})

	store := NewRedisStore(connection)
	session, err := store.Get("session_id", "session_name")

	if err != nil {
		t.Error("does not expect error when session is not found")
	}

	if session == nil {
		t.Error("expect a new session to be returned")
	}

	if !session.IsNew {
		t.Error("session should be new")
	}

	if session.IsDirty() {
		t.Error("session shouldn't be dirty")
	}

	if session.isExpired() {
		t.Error("session should not be expired")
	}

	if session.createdOn().IsZero() || session.lastAccessed.IsZero() {
		t.Error("session should have created on and last accessed time")
	}
}

func TestSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mocks.NewMockUniversalClient(ctrl)
	connection := &redis.Connection{
		UniversalClient: mock,
	}

	store := NewRedisStore(connection)
	session, _ := store.New("session_name")
	session.Set("TEST", "TEST_Value")

	if !session.IsDirty() {
		t.Error("session should be dirty")
	}

	mock.EXPECT().HSet(gomock.Any(),
		fmt.Sprintf("%s:%s", session.name, session.ID),
		sessionValueField, gomock.Any(),
		sessionOptionField, gomock.Any(),
		sessionLastAccessedField, gomock.Any()).Return(&goRedis.IntCmd{})

	originalLastAccessed := session.lastAccessed
	var expiresAt time.Time
	mock.EXPECT().ExpireAt(gomock.Any(),
		fmt.Sprintf("%s:%s", session.name, session.ID),
		gomock.Any()).
		Do(func(ctx context.Context, key string, tm time.Time){
			expiresAt = tm
		}).
		Return(&goRedis.BoolCmd{})


	store.Save(session)

	if !session.lastAccessed.After(originalLastAccessed) && session.lastAccessed.Before(time.Now()) {
		t.Error("session last accessed time should be updated")
	}

	if !expiresAt.Equal(session.lastAccessed.Add(session.Options.IdleTimeout)) {
		t.Error("session should expire at idle time out")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}
}