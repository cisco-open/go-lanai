package session

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/redismock"
	"encoding/gob"
	"fmt"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"strconv"
	"testing"
	"time"
)

type testUser struct {
	User string
	Pass string
}

func (u *testUser) ID() interface{} {
	return u.User
}

func (u *testUser) Type() security.AccountType {
	return security.AccountTypeDefault
}

func (u *testUser) Username() string {
	return u.User
}

func (u *testUser) Credentials() interface{} {
	return u.Pass
}

func (u *testUser) Permissions() []string {
	var p []string
	return p
}

func (u *testUser) Disabled() bool {
	return false
}

func (u *testUser) Locked() bool {
	return false
}

func (u *testUser) UseMFA() bool {
	return false
}

func (c *testUser) CacheableCopy() security.Account {
	copy := testUser{
		User: c.User,
		Pass: "",
	}
	return &copy
}


type testAuthentication struct {
	Account     security.Account
	PermissionList map[string]interface{}
}

func (auth *testAuthentication) Principal() interface{} {
	return auth.Account
}

func (auth *testAuthentication) Permissions() security.Permissions {
	return auth.PermissionList
}

func (auth *testAuthentication) State() security.AuthenticationState {
	return security.StateAuthenticated
}

func (auth *testAuthentication) Details() interface{} {
	return auth.Account
}

func (auth *testAuthentication) Username() string {
	return auth.Account.Username()
}

func (auth *testAuthentication) IsMFAPending() bool {
	return false
}

func (auth *testAuthentication) OTPIdentifier() string {
	return ""
}

func TestSerialization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	store := NewRedisStore(connection)
	s := NewSession(store, common.DefaultName)

	auth := testAuthentication{
		Account: &testUser{
			User: "test_user",
			Pass: "test_pass",
		},
		PermissionList: map[string]interface{}{"perm_a": true, "perm_b": true},
	}

	s.values["auth"] = auth

	gob.Register((*testAuthentication)(nil))
	gob.Register((*testUser)(nil))

	serialized, err := Serialize(s.values)

	if err != nil {
		t.Errorf("Cannot serialize %v", err)
	}

	s = NewSession(store, common.DefaultName)
	err = Deserialize(bytes.NewReader(serialized), &s.values)

	if err != nil {
		t.Errorf("Cannot deserialize %v", err)
	}

	deserializedAuth, ok := s.values["auth"]

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

	if account.Credentials() != "test_pass" || account.Username() != "test_user" {
		t.Errorf("account username password does not match")
	}

	if len(userAuth.Permissions()) != 2 {
		t.Errorf("permissions does not have the right length")
	}
}

func Test_Get_Not_Exist_Session_Should_Create_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	mock.EXPECT().
		HGetAll(gomock.Any(), gomock.Eq("LANAI:SESSION:session_name:session_id")).
		Return(&goRedis.StringStringMapCmd{})

	store := NewRedisStore(connection)

	session, err := store.Get("session_id", "session_name")

	if err != nil {
		t.Error("does not expect error when session is not found")
	}

	if session == nil {
		t.Error("expect a new session to be returned")
		return
	}

	if !session.isNew {
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

func Test_Get_Exist_Session_Should_Return_Existing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock
	store := NewRedisStore(connection)

	var sessionValues = make(map[interface{}]interface{})
	sessionValues[createdTimeKey] = time.Now()
	valueBytes, err := Serialize(sessionValues)
	if err != nil {
		t.Errorf("not able to serialize session values %v", err)
	}

	options := &Options{
		IdleTimeout: 900 * time.Second,
		AbsoluteTimeout: 1800 * time.Second,
	}
	optionBytes, err := Serialize(options)
	if err != nil {
		t.Errorf("not able to serialize session values %v", err)
	}

	var hset = make(map[string]string)
	hset[sessionValueField] = string(valueBytes)
	hset[common.SessionLastAccessedField] = strconv.FormatInt(time.Now().Unix(), 10)
	hset[sessionOptionField] = string(optionBytes)

	mock.EXPECT().
		HGetAll(gomock.Any(), gomock.Eq("LANAI:SESSION:session_name:session_id")).
		Return(goRedis.NewStringStringMapResult(hset, nil))

	session, err := store.Get("session_id", "session_name")

	if err != nil {
		t.Error("does not expect error when session is not found")
	}

	if session == nil {
		t.Error("expect a new session to be returned")
		return
	}

	if session.isNew {
		t.Error("session should not be new")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}

	if session.isExpired() {
		t.Error("session should not be expired")
	}

	if session.createdOn().IsZero() || session.lastAccessed.IsZero() {
		t.Error("session should have created on and last accessed time")
	}
}

func TestSaveNewSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	store := NewRedisStore(connection)
	session, _ := store.New("session_name")

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}

	mock.EXPECT().HSet(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		sessionValueField, gomock.Any(),
		sessionOptionField, gomock.Any(),
		common.SessionIdleTimeoutDuration, gomock.Any(),
		common.SessionAbsTimeoutTime, gomock.Any(),
		common.SessionLastAccessedField, gomock.Any()).Return(&goRedis.IntCmd{})

	originalLastAccessed := session.lastAccessed
	var expiresAt time.Time
	mock.EXPECT().ExpireAt(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		gomock.Any()).
		Do(func(ctx context.Context, key string, tm time.Time){
			expiresAt = tm
		}).
		Return(&goRedis.BoolCmd{})


	_ = store.Save(session)

	if !session.lastAccessed.After(originalLastAccessed) && session.lastAccessed.Before(time.Now()) {
		t.Error("session last accessed time should be updated")
	}

	if !expiresAt.Equal(session.lastAccessed.Add(session.options.IdleTimeout)) {
		t.Error("session should expire at idle time out")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}
}

func TestSaveNewSessionWithIdleTimeoutDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	store := NewRedisStore(connection, func(o *StoreOption){
		o.IdleTimeout = -1 * time.Second
	})
	session, _ := store.New("session_name")

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}

	mock.EXPECT().HSet(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		sessionValueField, gomock.Any(),
		sessionOptionField, gomock.Any(),
		common.SessionAbsTimeoutTime, gomock.Any(),
		common.SessionLastAccessedField, gomock.Any()).Return(&goRedis.IntCmd{})

	originalLastAccessed := session.lastAccessed
	var expiresAt time.Time
	mock.EXPECT().ExpireAt(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		gomock.Any()).
		Do(func(ctx context.Context, key string, tm time.Time){
			expiresAt = tm
		}).
		Return(&goRedis.BoolCmd{})

	_ = store.Save(session)

	if !session.lastAccessed.After(originalLastAccessed) && session.lastAccessed.Before(time.Now()) {
		t.Error("session last accessed time should be updated")
	}

	if !expiresAt.Equal(session.createdOn().Add(session.options.AbsoluteTimeout)) {
		t.Error("session should expire at absolute time out")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}
}

func TestSaveNewSessionWithAbsTimeoutDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	store := NewRedisStore(connection, func(o *StoreOption){
		o.AbsoluteTimeout = -1 * time.Second
	})
	session, _ := store.New("session_name")

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}

	mock.EXPECT().HSet(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		sessionValueField, gomock.Any(),
		sessionOptionField, gomock.Any(),
		common.SessionIdleTimeoutDuration, gomock.Any(),
		common.SessionLastAccessedField, gomock.Any()).Return(&goRedis.IntCmd{})

	originalLastAccessed := session.lastAccessed
	var expiresAt time.Time
	mock.EXPECT().ExpireAt(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		gomock.Any()).
		Do(func(ctx context.Context, key string, tm time.Time){
			expiresAt = tm
		}).
		Return(&goRedis.BoolCmd{})

	_ = store.Save(session)

	if !session.lastAccessed.After(originalLastAccessed) && session.lastAccessed.Before(time.Now()) {
		t.Error("session last accessed time should be updated")
	}

	if !expiresAt.Equal(session.lastAccessed.Add(session.options.IdleTimeout)) {
		t.Error("session should expire at idle time out")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}
}

func TestSaveNewSessionWithTimeoutDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	store := NewRedisStore(connection, func(o *StoreOption){
		o.IdleTimeout = -1 * time.Second
		o.AbsoluteTimeout = -1 * time.Second
	})
	session, _ := store.New("session_name")

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}

	mock.EXPECT().HSet(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		sessionValueField, gomock.Any(),
		sessionOptionField, gomock.Any(),
		common.SessionLastAccessedField, gomock.Any()).Return(&goRedis.IntCmd{})

	originalLastAccessed := session.lastAccessed

	_ = store.Save(session)

	if !session.lastAccessed.After(originalLastAccessed) && session.lastAccessed.Before(time.Now()) {
		t.Error("session last accessed time should be updated")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}
}

func TestSaveDirtySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	store := NewRedisStore(connection)
	session, _ := store.New("session_name")

	//set this session so that it's not a new session (similar to a session that is loaded from store)
	//and also set an attribute so that it's dirty
	session.isNew = false
	session.Set("TEST", "TEST_Value")

	if !session.IsDirty() {
		t.Error("session should be dirty")
	}

	//since it's dirty, but it's not new, we only expect two fields to be serialized and saved
	mock.EXPECT().HSet(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		sessionValueField, gomock.Any(),
		common.SessionLastAccessedField, gomock.Any()).Return(&goRedis.IntCmd{})

	originalLastAccessed := session.lastAccessed
	var expiresAt time.Time
	mock.EXPECT().ExpireAt(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		gomock.Any()).
		Do(func(ctx context.Context, key string, tm time.Time){
			expiresAt = tm
		}).
		Return(&goRedis.BoolCmd{})


	_ = store.Save(session)

	if !session.lastAccessed.After(originalLastAccessed) && session.lastAccessed.Before(time.Now()) {
		t.Error("session last accessed time should be updated")
	}

	if !expiresAt.Equal(session.lastAccessed.Add(session.options.IdleTimeout)) {
		t.Error("session should expire at idle time out")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty after being saved")
	}
}

//The session is not dirty and is not new, so we only expect last accessed time to be updated
func TestSaveAccessedSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := redismock.NewMockUniversalClient(ctrl)
	connection := mock

	store := NewRedisStore(connection)
	session, _ := store.New("session_name")

	//set this session so that it's not a new session (similar to a session that is loaded from store)
	session.isNew = false

	if session.IsDirty() {
		t.Error("session should not be dirty")
	}

	//since it's not new and not dirty, we only expect one field to be serialized and saved
	mock.EXPECT().HSet(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		common.SessionLastAccessedField, gomock.Any()).Return(&goRedis.IntCmd{})

	originalLastAccessed := session.lastAccessed
	var expiresAt time.Time
	mock.EXPECT().ExpireAt(gomock.Any(),
		fmt.Sprintf("LANAI:SESSION:%s:%s", session.name, session.id),
		gomock.Any()).
		Do(func(ctx context.Context, key string, tm time.Time){
			expiresAt = tm
		}).
		Return(&goRedis.BoolCmd{})


	_ = store.Save(session)

	if !session.lastAccessed.After(originalLastAccessed) && session.lastAccessed.Before(time.Now()) {
		t.Error("session last accessed time should be updated")
	}

	if !expiresAt.Equal(session.lastAccessed.Add(session.options.IdleTimeout)) {
		t.Error("session should expire at idle time out")
	}

	if session.IsDirty() {
		t.Error("session should not be dirty after being saved")
	}
}

//TODO: add save accessed session but with either idle, abs timeout disabled or both disabled

