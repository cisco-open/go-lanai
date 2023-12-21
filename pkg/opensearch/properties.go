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

package opensearch

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"embed"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "data.opensearch"
)

//go:embed defaults-opensearch.yml
var defaultConfigFS embed.FS

type Properties struct {
	Addresses []string `json:"addresses"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	TLS       TLS      `json:"tls"`
}

type TLS struct {
	Enable bool                   `json:"enable"`
	Certs  certs.SourceProperties `json:"certs"`
}

func NewOpenSearchProperties() *Properties {
	return &Properties{} // None by default, they should all be defined in the defaults-opensearch.yml
}

func BindOpenSearchProperties(ctx *bootstrap.ApplicationContext) *Properties {
	props := NewOpenSearchProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind OpenSearchProperties"))
	}
	return props
}
