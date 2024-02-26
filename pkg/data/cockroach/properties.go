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

package cockroach

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/certs"
	"github.com/pkg/errors"
)

const (
	CockroachPropertiesPrefix = "data.cockroach"
)

type CockroachProperties struct {
	//Enabled       bool                               `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SslMode  string `json:"sslmode"`
	Tls      TLS    `json:"tls"`
}

type TLS struct {
	Enable bool                   `json:"enabled"`
	Certs  certs.SourceProperties `json:"certs"`
}

// NewCockroachProperties create a CockroachProperties with default values
func NewCockroachProperties() *CockroachProperties {
	return &CockroachProperties{
		Host:     "localhost",
		Port:     26257,
		Username: "root",
		Password: "root",
		SslMode:  "disable",
	}
}

// BindCockroachProperties create and bind SessionProperties, with a optional prefix
func BindCockroachProperties(ctx *bootstrap.ApplicationContext) CockroachProperties {
	props := NewCockroachProperties()
	if err := ctx.Config().Bind(props, CockroachPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind CockroachProperties"))
	}
	return *props
}
