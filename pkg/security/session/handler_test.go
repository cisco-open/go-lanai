package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mock_redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mock_security"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestChangeSessionHandler_HandleAuthenticationSuccess(t *testing.T) {
	handler := &ChangeSessionHandler{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := mock_redis.NewMockUniversalClient(ctrl)

	sessionStore := NewRedisStore(mockRedis)
	s, _ := sessionStore.New(DefaultName)
	s.isNew = false //if session is new it won't get changed

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	//The actual request is not important
	c.Request = httptest.NewRequest("POST", "/something", nil)

	mockFrom := mock_security.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	mockTo := mock_security.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)

	originalId := s.id

	mockRedis.EXPECT().Rename(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, key, newkey string) *redis.StatusCmd {
		if key != "LANAI:SESSION"+":"+DefaultName+":"+originalId {
			t.Errorf("expected original key")
		}

		if newkey == "LANAI:SESSION"+":"+DefaultName+":"+originalId {
			t.Errorf("expected changed key")
		}

		return redis.NewStatusCmd(ctx, key, newkey)
	})

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)

	resp := recorder.Result()
	if resp.Header.Get("Set-Cookie") == "" {
		t.Errorf("Should set new session in response header")
	}
}

func TestConcurrentSessionHandler_HandleAuthenticationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := mock_redis.NewMockUniversalClient(ctrl)

	sessionStore := NewRedisStore(mockRedis)

	//this handler allows 1 concurrent sessions
	handler := &ConcurrentSessionHandler{
		sessionStore: sessionStore,
		getMaxSessions: func() int {
			return 1
		},
	}

	s, _ := sessionStore.New(DefaultName)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	//The actual request is not important
	c.Request = httptest.NewRequest("POST", "/something", nil)

	mockFrom := mock_security.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	principalName := "principal1"
	mockTo := mock_security.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)
	mockTo.EXPECT().Principal().Return(principalName).AnyTimes()

	mockRedis.EXPECT().SAdd(gomock.Any(), "LANAI:SESSION:INDEX:"+DefaultName+":"+principalName, s.id).Return(redis.NewIntCmd(context.Background()))

	existingId := "1"
	mockRedis.EXPECT().SScan(gomock.Any(), "LANAI:SESSION:INDEX:"+DefaultName+":"+principalName, uint64(0), "", int64(0)).
		Return(redis.NewScanCmdResult([]string{s.id, existingId}, 0, nil))

	//Mock current session
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
	hset[sessionLastAccessedField] = strconv.FormatInt(time.Now().Unix(), 10)
	hset[sessionOptionField] = string(optionBytes)

	mockRedis.EXPECT().
		HGetAll(gomock.Any(), gomock.Eq("LANAI:SESSION:" + DefaultName + ":" + s.id)).
		Return(redis.NewStringStringMapResult(hset, nil))

	//Mock existing session
	sessionValues = make(map[interface{}]interface{})
	sessionValues[createdTimeKey] = time.Now().Add(-time.Second * 30)
	valueBytes, err = Serialize(sessionValues)
	if err != nil {
		t.Errorf("not able to serialize session values %v", err)
	}
	hset = make(map[string]string)
	hset[sessionValueField] = string(valueBytes)
	hset[sessionLastAccessedField] = strconv.FormatInt(time.Now().Unix(), 10)
	hset[sessionOptionField] = string(optionBytes)

	mockRedis.EXPECT().
		HGetAll(gomock.Any(), gomock.Eq("LANAI:SESSION:" + DefaultName + ":" + existingId)).
		Return(redis.NewStringStringMapResult(hset, nil))

	mockRedis.EXPECT().Del(gomock.Any(), "LANAI:SESSION:" + DefaultName + ":" + existingId).Return(redis.NewIntCmd(context.Background()))
	// FIXME this caused issue after recent session update
	//mockRedis.EXPECT().SRem(gomock.Any(), "LANAI:SESSION:INDEX:"+DefaultName+":"+principalName, existingId).Return(redis.NewIntCmd(context.Background()))

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)
}

func TestDeleteSessionOnLogoutHandler_HandleAuthenticationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := mock_redis.NewMockUniversalClient(ctrl)
	sessionStore := NewRedisStore(mockRedis)

	s, _ := sessionStore.New(DefaultName)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	//The actual request is not important
	c.Request = httptest.NewRequest("POST", "/something", nil)

	principalName := "principal1"
	mockFrom := mock_security.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAuthenticated)
	mockFrom.EXPECT().Principal().Return(principalName).AnyTimes()

	mockTo := mock_security.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAnonymous)

	handler := DeleteSessionOnLogoutHandler{
		sessionStore: sessionStore,
	}

	mockRedis.EXPECT().Del(gomock.Any(), "LANAI:SESSION:" + DefaultName + ":" + s.id).Return(redis.NewIntCmd(context.Background()))
	// FIXME this caused issue after recent session update
	//mockRedis.EXPECT().SRem(gomock.Any(), "LANAI:SESSION:INDEX:"+DefaultName+":"+principalName, s.id).Return(redis.NewIntCmd(context.Background()))

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)
}