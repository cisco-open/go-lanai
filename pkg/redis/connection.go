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

package redis

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// KeepTTL is an option for Set command to keep key's existing TTL.
// For example:
//
//	rdb.Set(ctx, key, value, redis.KeepTTL)
const KeepTTL = redis.KeepTTL

// ConnOptions options for connectivity by manipulating redis.UniversalOptions
type ConnOptions func(opt *redis.UniversalOptions) error

func GetUniversalOptions(p *RedisProperties, opts ...ConnOptions) (*redis.UniversalOptions, error) {
	universal := &redis.UniversalOptions{
		Addrs:              p.Addresses,
		DB:                 p.DB,
		Username:           p.Username,
		Password:           p.Password,
		MaxRetries:         p.MaxRetries,
		MinRetryBackoff:    p.MinRetryBackoff,
		MaxRetryBackoff:    p.MaxRetryBackoff,
		DialTimeout:        p.DialTimeout,
		ReadTimeout:        p.ReadTimeout,
		WriteTimeout:       p.WriteTimeout,
		PoolSize:           p.PoolSize,
		MinIdleConns:       p.MinIdleConns,
		MaxConnAge:         p.MaxConnAge,
		PoolTimeout:        p.PoolTimeout,
		IdleTimeout:        p.IdleTimeout,
		IdleCheckFrequency: p.IdleCheckFrequency,
		// Only cluster clients.
		MaxRedirects:   p.MaxRedirects,
		ReadOnly:       p.ReadOnly,
		RouteByLatency: p.RouteByLatency,
		RouteRandomly:  p.RouteRandomly,

		// The sentinel master name.
		// Only failover clients.
		MasterName:       p.MasterName,
		SentinelPassword: p.SentinelPassword,
	}

	for _, fn := range opts {
		if e := fn(universal); e != nil {
			return nil, e
		}
	}
	return universal, nil
}

func withDB(dbIndex int) ConnOptions {
	return func(opt *redis.UniversalOptions) error {
		opt.DB = dbIndex
		return nil
	}
}

func withTLS(ctx context.Context, certsMgr certs.Manager, p *certs.SourceProperties) ConnOptions {
	return func(opt *redis.UniversalOptions) error {
		if certsMgr == nil {
			return fmt.Errorf("TLS auth is enabled for Redis, but certificate manager is not available")
		}
		src, err := certsMgr.Source(ctx, certs.WithSourceProperties(p))
		if err != nil {
			return errors.Wrapf(err, "failed to initialize redis connection: %v", err)
		}

		opt.TLSConfig, err = src.TLSConfig(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize redis connection: %v", err)
		}
		return nil
	}
}

type Client interface {
	redis.UniversalClient
}

type client struct {
	redis.UniversalClient
}
