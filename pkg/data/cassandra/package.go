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

package cassandra

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/gocql/gocql"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("Cassandra")

var Module = &bootstrap.Module{
	Name:       "cassandra",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(BindCassandraProperties, NewSession),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func NewSession(p CassandraProperties) *gocql.Session {
	cluster := gocql.NewCluster(p.Hosts()...)
	cluster.Keyspace = p.KeySpaceName
	cluster.Consistency = gocql.ParseConsistency(p.Consistency)
	cluster.Timeout = time.Duration(p.Timeout)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: p.Username,
		Password: p.Password,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		logger.Errorf("unable to create session: %v", err)
	}
	return session
}

func BindCassandraProperties(ctx *bootstrap.ApplicationContext) CassandraProperties {
	p := NewCassandraProperties()
	_ = ctx.Config().Bind(p, CassandraPropertiesPrefix) // Note, we don't panic if this bind is missing
	return *p
}
