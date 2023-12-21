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

package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const InmemoryClientsPropertiesPrefix = "security.in-memory"

type PropertiesBasedClient struct {
	ClientId             string   `json:"client-id"`
	Secret               string   `json:"secret"`
	GrantTypes           []string `json:"grant-types"`
	RedirectUris         []string `json:"redirect-uris"`
	Scopes               []string `json:"scopes"`
	AutoApproveScopes    []string `json:"auto-approve-scopes"`
	AccessTokenValidity  string   `json:"access-token-validity"`
	RefreshTokenValidity string   `json:"refresh-token-validity"`
	UseSessionTimeout    bool     `json:"use-session-timeout"`
	TenantRestrictions   []string `json:"tenant-restrictions"`
}

type ClientsProperties struct {
	Clients map[string]PropertiesBasedClient `json:"clients"`
}

func NewClientsProperties() *ClientsProperties {
	return &ClientsProperties {
		Clients: map[string]PropertiesBasedClient{},
	}
}

func BindClientsProperties(ctx *bootstrap.ApplicationContext) ClientsProperties {
	props := NewClientsProperties()
	if err := ctx.Config().Bind(props, InmemoryClientsPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ClientsProperties"))
	}
	return *props
}
