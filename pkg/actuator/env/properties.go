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

package env

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
)

const (
	EnvPropertiesPrefix = "management.endpoint.env"
)

type EnvProperties struct {
	// KeysToSanitize holds list of regular expressions
	KeysToSanitize utils.StringSet `json:"keys-to-sanitize"`
}

//NewSessionProperties create a SessionProperties with default values
func NewEnvProperties() *EnvProperties {
	return &EnvProperties{
		KeysToSanitize: utils.NewStringSet(
			`.*password.*`, `.*secret.*`, `key`,
			`.*credentials.*`, `vcap_services`, `sun.java.command`,
		),
	}
}

//BindHealthProperties create and bind SessionProperties, with a optional prefix
func BindEnvProperties(ctx *bootstrap.ApplicationContext) EnvProperties {
	props := NewEnvProperties()
	if err := ctx.Config().Bind(props, EnvPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind EnvProperties"))
	}
	return *props
}

