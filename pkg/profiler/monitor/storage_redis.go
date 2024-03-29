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

package monitor

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/utils"
    goRedis "github.com/go-redis/redis/v8"
    "time"
)

const (
	redisDB    = 6
	prefixData = "D"
	ttl = 5 * time.Second
)

type redisDataStorage struct {
	client redis.Client
	identifier string
}

func NewRedisDataStorage(ctx context.Context, cf redis.ClientFactory) *redisDataStorage {
	client, e := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = redisDB
	})
	if e != nil {
		panic(e)
	}

	return &redisDataStorage{
		client: client,
		identifier: utils.RandomString(10),
	}
}

func (s *redisDataStorage) Read(ctx context.Context, groups...DataGroup) (map[DataGroup]RawEntries, error) {
	pipeliner := s.client.Pipeline()
	defer func() { _ = pipeliner.Close()}()

	for _, group := range groups {
		key := s.groupKey(group)
		pipeliner.LRange(ctx, key, 0, -1)
	}

	cmds, e := pipeliner.Exec(ctx)
	if e != nil {
		return nil, e
	}
	data := map[DataGroup]RawEntries{}
	for i, cmd := range cmds {
		if cmd.Err() != nil {
			return nil, fmt.Errorf("%s failed: %v", cmd.Name(), cmd.Err())
		}
		switch result := cmd.(type) {
		case *goRedis.StringSliceCmd:
			// Note: redis store our data in a reversed order
			data[groups[i]] = s.reverse(result.Val())
		}
	}
	return data, nil
}

func (s *redisDataStorage) AppendAll(ctx context.Context, data map[DataGroup]interface{}, cap int64) error {
	pipeliner := s.client.Pipeline()
	defer func() { _ = pipeliner.Close()}()

	for group, entry := range data {
		key := s.groupKey(group)
		if entry != nil {
			pipeliner.LPush(ctx, key, entry)
			pipeliner.LTrim(ctx, key, 0, cap - 1)
		}
		pipeliner.Expire(ctx, key, ttl)
	}

	cmds, e := pipeliner.Exec(ctx)
	if e != nil {
		return e
	}
	for _, cmd := range cmds {
		if cmd.Err() != nil {
			return fmt.Errorf("%s failed: %v", cmd.Name(), cmd.Err())
		}
	}
	return e
}


func (s *redisDataStorage) Append(ctx context.Context, group DataGroup, entry interface{}, cap int64) error {
	key := s.groupKey(group)
	// Note: we ignore Redis errors from each command
	_, e := s.client.Pipelined(ctx, func(pipeliner goRedis.Pipeliner) error {
		if entry != nil {
			pipeliner.LPush(ctx, key, entry)
			pipeliner.LTrim(ctx, key, 0, cap - 1)
		}
		pipeliner.Expire(ctx, key, ttl)
		return nil
	})
	return e
}

func (s *redisDataStorage) groupKey(group DataGroup) string {
	return fmt.Sprintf(`%s:%s:%s`, prefixData, s.identifier, group)
}

func (s *redisDataStorage) reverse(data []string) []string {
	size := len(data)
	for i := 0; i < size / 2; i++ {
		j := size - i - 1
		v := data[i]
		data[i] = data[j]
		data[j] = v
	}
	return data
}
