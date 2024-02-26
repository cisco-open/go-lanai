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
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/security/session/common"
	"strconv"
	"time"
)

type RedisTimeoutApplier struct {
	sessionName string
	client      redis.Client
}

func NewRedisTimeoutApplier(client redis.Client) *RedisTimeoutApplier {
	return &RedisTimeoutApplier{
		sessionName: common.DefaultName,
		client:      client,
	}
}

func (r *RedisTimeoutApplier) ApplyTimeout(ctx context.Context, sessionId string) (valid bool, err error) {
	key := common.GetRedisSessionKey(r.sessionName, sessionId)

	//check if session exists
	existCmd := r.client.Exists(ctx, key)
	if existCmd.Err() != nil {
		valid = false
		err = existCmd.Err()
		return
	} else {
		valid = existCmd.Val() == 1
	}

	if !valid {
		return
	}

	hmGetCmd := r.client.HMGet(ctx, key, common.SessionIdleTimeoutDuration, common.SessionAbsTimeoutTime)
	if hmGetCmd.Err() != nil {
		err = hmGetCmd.Err()
		return
	}
	result, _ := hmGetCmd.Result()

	var timeoutSetting common.TimeoutSetting = 0
	var idleExpiration, absExpiration time.Time
	now := time.Now()

	if result[0] != nil {
		idleTimeout, e := time.ParseDuration(result[0].(string))
		if e != nil {
			err = e
			return
		}
		idleExpiration = now.Add(idleTimeout)
		timeoutSetting = timeoutSetting | common.IdleTimeoutEnabled
	}

	if result[1] != nil {
		absTimeoutUnixTime, e := strconv.ParseInt(result[1].(string), 10, 0)
		if e != nil {
			err = e
			return
		}
		absExpiration = time.Unix(absTimeoutUnixTime, 0)
		timeoutSetting = timeoutSetting | common.AbsoluteTimeoutEnabled
	}

	canExpire, expiration := common.CalculateExpiration(timeoutSetting, idleExpiration, absExpiration)

	//update session last accessed time
	hsetCmd := r.client.HSet(ctx, key, common.SessionLastAccessedField, now.Unix())
	if hsetCmd.Err() != nil {
		err = hsetCmd.Err()
		return
	}
	if canExpire {
		expireCmd := r.client.ExpireAt(ctx, key, expiration)
		err = expireCmd.Err()
	}
	return
}
