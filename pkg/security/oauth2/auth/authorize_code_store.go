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

package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"fmt"
	"time"
)

const (
	defaultAuthCodeLength = 32
	authCodePrefix = "AC"
)

var (
	authCodeValidity = 5 * time.Minute
)

/**********************
	Abstraction
 **********************/

type AuthorizationCodeStore interface {
	GenerateAuthorizationCode(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (string, error)
	ConsumeAuthorizationCode(ctx context.Context, authCode string, onetime bool) (oauth2.Authentication, error)
}

/**********************
	Redis Impl
 **********************/

// RedisAuthorizationCodeStore store authorization code in Redis
type RedisAuthorizationCodeStore struct {
	redisClient redis.Client
}

func NewRedisAuthorizationCodeStore(ctx context.Context, cf redis.ClientFactory, dbIndex int) *RedisAuthorizationCodeStore {
	client, e := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = dbIndex
	})
	if e != nil {
		panic(e)
	}

	return &RedisAuthorizationCodeStore{
		redisClient: client,
	}
}

func (s *RedisAuthorizationCodeStore) GenerateAuthorizationCode(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (string, error) {
	// code_challenge_method and code_challenge is stored in both parameters and extensions.
	// so no need to save them separately
	request := r.OAuth2Request()
	userAuth := ConvertToOAuthUserAuthentication(user)
	toSave := oauth2.NewAuthentication(func(conf *oauth2.AuthOption) {
		conf.Request = request
		conf.UserAuth = userAuth
	})
	code := utils.RandomStringWithCharset(defaultAuthCodeLength, utils.CharsetAlphabetic)

	if e := s.save(ctx, code, toSave); e != nil {
		return "", oauth2.NewInternalError(e)
	}
	return code, nil
}

func (s *RedisAuthorizationCodeStore) ConsumeAuthorizationCode(ctx context.Context, authCode string, onetime bool) (oauth2.Authentication, error) {
	key := s.authCodeRedisKey(authCode)
	cmd := s.redisClient.Get(ctx, key)
	if cmd.Err() != nil {
		return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("code [%s] is not valid", authCode))
	}

	toLoad := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = oauth2.NewOAuth2Request()
		opt.UserAuth = oauth2.NewUserAuthentication()
		opt.Details = map[string]interface{}{}
	})

	e := json.Unmarshal([]byte(cmd.Val()), &toLoad)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("code [%s] is not valid", authCode), e)
	}

	if onetime {
		if cmd := s.redisClient.Del(ctx, key); cmd.Err() != nil {
			logger.WithContext(ctx).Warnf("authorization code was not removed: %v", cmd.Err())
		}
	}
	return toLoad, nil
}

/**********************
	Helpers
 **********************/
func (s *RedisAuthorizationCodeStore) save(ctx context.Context, code string, oauth oauth2.Authentication) error {
	key := s.authCodeRedisKey(code)
	toSave, e := json.Marshal(oauth)
	if e != nil {
		return e
	}

	cmd := s.redisClient.Set(ctx, key, toSave, authCodeValidity)
	return cmd.Err()
}

func (s *RedisAuthorizationCodeStore) authCodeRedisKey(code string) string {
	return fmt.Sprintf("%s:%s", authCodePrefix, code)
}

