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

package timeoutsupport

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security/session/common"
    "github.com/cisco-open/go-lanai/test/mocks/redismock"
    "github.com/go-redis/redis/v8"
    "github.com/golang/mock/gomock"
    "github.com/onsi/gomega"
    "github.com/stretchr/testify/mock"
    "strconv"
    "testing"
    "time"
)

func TestApplyTimeout_WhenSessionExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sessionId := "my_session_id"

	mockRedis := redismock.NewMockUniversalClient(ctrl)
	//mock that the session doesn't exist (i.e. expired)
	mockRedis.EXPECT().Exists(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId)).Return(redis.NewIntResult(0, nil))

	timeoutApplier := NewRedisTimeoutApplier(mockRedis)

	valid, err := timeoutApplier.ApplyTimeout(context.Background(), sessionId)

	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(valid).To(gomega.BeFalse())
}

func TestApplyTimeout_whenSessionNotExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sessionId := "my_session_id"
	testBeginTime := time.Now()

	//let the session be accessed 10 seconds ago.
	sessionLastAccessedTime := time.Now().Add(-10 * time.Second)
	idleTimeout := 60 * time.Second
	origAbsTimeoutTime := sessionLastAccessedTime.Add(120 * time.Second)
	origIdleTimeoutTime := sessionLastAccessedTime.Add(idleTimeout)

	mockRedis := redismock.NewMockUniversalClient(ctrl)
	//mock that the session exists
	mockRedis.EXPECT().Exists(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId)).
		Return(redis.NewIntResult(1, nil))
	mockRedis.EXPECT().HMGet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionIdleTimeoutDuration, common.SessionAbsTimeoutTime).
		Return(redis.NewSliceResult([]interface{}{idleTimeout.String(), strconv.FormatInt(origAbsTimeoutTime.Unix(), 10)}, nil))

	//expects last accessed time to be updated
	lastAccessTimeMatcher := mock.MatchedBy(func(epoch int64) bool {
		return epoch >= testBeginTime.Unix()
	})
	mockRedis.EXPECT().HSet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionLastAccessedField, lastAccessTimeMatcher).
		Return(redis.NewIntResult(1, nil))

	expirationMatcher := mock.MatchedBy(func(expiration time.Time) bool {
		return expiration.After(origIdleTimeoutTime) && expiration.Before(origAbsTimeoutTime)
	})
	mockRedis.EXPECT().ExpireAt(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), expirationMatcher).Return(redis.NewBoolResult(true, nil))

	timeoutApplier := NewRedisTimeoutApplier(mockRedis)

	valid, err := timeoutApplier.ApplyTimeout(context.Background(), sessionId)

	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(valid).To(gomega.BeTrue())
}

func TestApplyTimeout_whenIdleTimeoutDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sessionId := "my_session_id"
	testBeginTime := time.Now()

	//let the session be accessed 10 seconds ago.
	sessionLastAccessedTime := time.Now().Add(-10 * time.Second)
	origAbsTimeoutTime := sessionLastAccessedTime.Add(120 * time.Second)

	mockRedis := redismock.NewMockUniversalClient(ctrl)
	//mock that the session exists
	mockRedis.EXPECT().Exists(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId)).
		Return(redis.NewIntResult(1, nil))
	mockRedis.EXPECT().HMGet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionIdleTimeoutDuration, common.SessionAbsTimeoutTime).
		Return(redis.NewSliceResult([]interface{}{nil, strconv.FormatInt(origAbsTimeoutTime.Unix(), 10)}, nil))

	//expects last accessed time to be updated
	lastAccessTimeMatcher := mock.MatchedBy(func(epoch int64) bool {
		return epoch >= testBeginTime.Unix()
	})
	mockRedis.EXPECT().HSet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionLastAccessedField, lastAccessTimeMatcher).
		Return(redis.NewIntResult(1, nil))

	expirationMatcher := mock.MatchedBy(func(expiration time.Time) bool {
		return expiration.Unix() == origAbsTimeoutTime.Unix()
	})
	mockRedis.EXPECT().ExpireAt(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), expirationMatcher).Return(redis.NewBoolResult(true, nil))

	timeoutApplier := NewRedisTimeoutApplier(mockRedis)

	valid, err := timeoutApplier.ApplyTimeout(context.Background(), sessionId)

	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(valid).To(gomega.BeTrue())
}

func TestApplyTimeout_whenAbsTimeoutDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sessionId := "my_session_id"
	testBeginTime := time.Now()

	//let the session be accessed 10 seconds ago.
	sessionLastAccessedTime := time.Now().Add(-10 * time.Second)
	idleTimeout := 60 * time.Second
	origIdleTimeoutTime := sessionLastAccessedTime.Add(idleTimeout)

	mockRedis := redismock.NewMockUniversalClient(ctrl)
	//mock that the session exists
	mockRedis.EXPECT().Exists(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId)).
		Return(redis.NewIntResult(1, nil))
	mockRedis.EXPECT().HMGet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionIdleTimeoutDuration, common.SessionAbsTimeoutTime).
		Return(redis.NewSliceResult([]interface{}{idleTimeout.String(), nil}, nil))

	//expects last accessed time to be updated
	lastAccessTimeMatcher := mock.MatchedBy(func(epoch int64) bool {
		return epoch >= testBeginTime.Unix()
	})
	mockRedis.EXPECT().HSet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionLastAccessedField, lastAccessTimeMatcher).
		Return(redis.NewIntResult(1, nil))

	expirationMatcher := mock.MatchedBy(func(expiration time.Time) bool {
		return expiration.After(origIdleTimeoutTime)
	})
	mockRedis.EXPECT().ExpireAt(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), expirationMatcher).Return(redis.NewBoolResult(true, nil))

	timeoutApplier := NewRedisTimeoutApplier(mockRedis)

	valid, err := timeoutApplier.ApplyTimeout(context.Background(), sessionId)

	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(valid).To(gomega.BeTrue())
}

func TestApplyTimeout_whenBothTimeoutDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sessionId := "my_session_id"
	testBeginTime := time.Now()

	//let the session be accessed 10 seconds ago.

	mockRedis := redismock.NewMockUniversalClient(ctrl)
	//mock that the session exists
	mockRedis.EXPECT().Exists(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId)).
		Return(redis.NewIntResult(1, nil))
	mockRedis.EXPECT().HMGet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionIdleTimeoutDuration, common.SessionAbsTimeoutTime).
		Return(redis.NewSliceResult([]interface{}{nil, nil}, nil))

	//expects last accessed time to be updated
	lastAccessTimeMatcher := mock.MatchedBy(func(epoch int64) bool {
		return epoch >= testBeginTime.Unix()
	})
	mockRedis.EXPECT().HSet(gomock.Any(), fmt.Sprintf("LANAI:SESSION:SESSION:%s", sessionId), common.SessionLastAccessedField, lastAccessTimeMatcher).
		Return(redis.NewIntResult(1, nil))

	timeoutApplier := NewRedisTimeoutApplier(mockRedis)

	valid, err := timeoutApplier.ApplyTimeout(context.Background(), sessionId)

	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(valid).To(gomega.BeTrue())
}