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
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/authmock"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/redismock"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/gob"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"strconv"
	"testing"
	"time"
)

func TestChangeSessionHandler_HandleAuthenticationSuccess(t *testing.T) {
	handler := &ChangeSessionHandler{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := redismock.NewMockUniversalClient(ctrl)

	sessionStore := NewRedisStore(mockRedis)
	s, _ := sessionStore.New(common.DefaultName)
	s.isNew = false //if session is new it won't get changed

	//The actual request is not important
	c := webtest.NewGinContext(context.Background(),
		"POST", "/something", nil,
	)
	if e := Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}
	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)

	originalId := s.id

	mockRedis.EXPECT().Rename(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, key, newkey string) *redis.StatusCmd {
		if key != "LANAI:SESSION"+":"+common.DefaultName+":"+originalId {
			t.Errorf("expected original key")
		}

		if newkey == "LANAI:SESSION"+":"+common.DefaultName+":"+originalId {
			t.Errorf("expected changed key")
		}

		return redis.NewStatusCmd(ctx, key, newkey)
	})

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)

	resp := webtest.GinContextRecorder(c).Result()
	if resp.Header.Get("Set-Cookie") == "" {
		t.Errorf("Should set new session in response header")
	}
}

func TestConcurrentSessionHandler_HandleAuthenticationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := redismock.NewMockUniversalClient(ctrl)

	sessionStore := NewRedisStore(mockRedis)

	//this handler allows 1 concurrent sessions
	handler := &ConcurrentSessionHandler{
		sessionStore: sessionStore,
		sessionSettingService: NewDefaultSettingService(security.SessionProperties{
			MaxConcurrentSession: 1,
		}),
	}

	s, _ := sessionStore.New(common.DefaultName)

	//The actual request is not important
	c := webtest.NewGinContext(context.Background(),
		"POST", "/something", nil,
	)
	if e := Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}

	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	principalName := "principal1"
	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)
	mockTo.EXPECT().Principal().Return(principalName).AnyTimes()

	mockRedis.EXPECT().SAdd(gomock.Any(), "LANAI:SESSION:INDEX:"+common.DefaultName+":"+principalName, s.id).Return(redis.NewIntCmd(context.Background()))

	existingId := "1"
	mockRedis.EXPECT().SScan(gomock.Any(), "LANAI:SESSION:INDEX:"+common.DefaultName+":"+principalName, uint64(0), "", int64(0)).
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
	hset[common.SessionLastAccessedField] = strconv.FormatInt(time.Now().Unix(), 10)
	hset[sessionOptionField] = string(optionBytes)

	mockRedis.EXPECT().
		HGetAll(gomock.Any(), gomock.Eq("LANAI:SESSION:" + common.DefaultName + ":" + s.id)).
		Return(redis.NewStringStringMapResult(hset, nil))

	//Mock existing session
	sessionValues = make(map[interface{}]interface{})
	sessionValues[createdTimeKey] = time.Now().Add(-time.Second * 30)

	//need to register these two type so that they can be serialized.
	//can't use mock_security.NewMockAuthentication here because it can't be serialized - no exported fields
	gob.Register((*testAuthentication)(nil))
	gob.Register((*testUser)(nil))
	existingSessionAuth := &testAuthentication{
		Account: &testUser{
			User: principalName,
			Pass: "test_pass",
		},
		PermissionList: map[string]interface{}{},
	}
	sessionValues[sessionKeySecurity] = existingSessionAuth

	valueBytes, err = Serialize(sessionValues)
	if err != nil {
		t.Errorf("not able to serialize session values %v", err)
	}
	hset = make(map[string]string)
	hset[sessionValueField] = string(valueBytes)
	hset[common.SessionLastAccessedField] = strconv.FormatInt(time.Now().Unix(), 10)
	hset[sessionOptionField] = string(optionBytes)

	mockRedis.EXPECT().
		HGetAll(gomock.Any(), gomock.Eq("LANAI:SESSION:" + common.DefaultName + ":" + existingId)).
		Return(redis.NewStringStringMapResult(hset, nil))

	mockRedis.EXPECT().Del(gomock.Any(), "LANAI:SESSION:" + common.DefaultName + ":" + existingId).Return(redis.NewIntCmd(context.Background()))
	mockRedis.EXPECT().SRem(gomock.Any(), "LANAI:SESSION:INDEX:"+ common.DefaultName +":"+principalName, existingId).Return(redis.NewIntCmd(context.Background()))

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)
}

func TestDeleteSessionOnLogoutHandler_HandleAuthenticationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := redismock.NewMockUniversalClient(ctrl)
	sessionStore := NewRedisStore(mockRedis)

	s, _ := sessionStore.New(common.DefaultName)

	//The actual request is not important
	c := webtest.NewGinContext(context.Background(),
		"POST", "/something", nil,
	)
	if e := Set(c, s); e != nil {
		t.Errorf("failed to set session into context")
	}

	principalName := "principal1"
	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAuthenticated)
	mockFrom.EXPECT().Principal().Return(principalName).AnyTimes()

	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAnonymous)

	existingSessionAuth := &testAuthentication{
		Account: &testUser{
			User: principalName,
			Pass: "test_pass",
		},
		PermissionList: map[string]interface{}{},
	}
	s.values[sessionKeySecurity] = existingSessionAuth

	handler := DeleteSessionOnLogoutHandler{
		sessionStore: sessionStore,
	}

	mockRedis.EXPECT().Del(gomock.Any(), "LANAI:SESSION:" + common.DefaultName + ":" + s.id).Return(redis.NewIntCmd(context.Background()))
	mockRedis.EXPECT().SRem(gomock.Any(), "LANAI:SESSION:INDEX:"+ common.DefaultName+":"+principalName, s.id).Return(redis.NewIntCmd(context.Background()))

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)
}