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
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/certs"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	ConfigRootRedisConnection = "redis"
	DefaultDbIndex            = 0
)

type RedisProperties struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addresses utils.CommaSeparatedSlice `json:"addrs"`

	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int `json:"db"`

	// Common options.
	Username string `json:"username"`
	Password string `json:"password"`

	MaxRetries      int           `json:"max-retries"`
	MinRetryBackoff time.Duration `json:"min-retry-backoff"`
	MaxRetryBackoff time.Duration `json:"max-retry-backoff"`

	DialTimeout  time.Duration `json:"dial-timeout"`
	ReadTimeout  time.Duration `json:"read-timeout"`
	WriteTimeout time.Duration `json:"write-timeout"`

	PoolSize           int           `json:"pool-size"`
	MinIdleConns       int           `json:"min-idle-conns"`
	MaxConnAge         time.Duration `json:"max-conn-age"`
	PoolTimeout        time.Duration `json:"pool-timeout"`
	IdleTimeout        time.Duration `json:"idle-timeout"`
	IdleCheckFrequency time.Duration `json:"idle-check-frequency"`

	// TLS Properties for Redis
	TLS TLSProperties `json:"tls"`
	// Only cluster clients.

	MaxRedirects   int  `json:"max-redirects"`
	ReadOnly       bool `json:"read-only"`
	RouteByLatency bool `json:"route-by-latency"`
	RouteRandomly  bool `json:"route-randomly"`

	// The sentinel master name.
	// Only failover clients.
	MasterName       string `json:"master-name"`
	SentinelPassword string `json:"sentinel-password"`
}

type TLSProperties struct {
	Enabled bool                   `json:"enabled"`
	Certs   certs.SourceProperties `json:"certs"`
}

func BindRedisProperties(ctx *bootstrap.ApplicationContext) RedisProperties {
	props := RedisProperties{}
	if err := ctx.Config().Bind(&props, ConfigRootRedisConnection); err != nil {
		panic(errors.Wrap(err, "failed to bind redis.RedisProperties"))
	}
	return props
}
