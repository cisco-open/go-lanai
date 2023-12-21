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

package extsamlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "security.idp.saml"
)

type SamlAuthProperties struct {
	Enabled   bool                       `json:"enabled"`
	Endpoints SamlAuthEndpointProperties `json:"endpoints"`
}

type SamlAuthEndpointProperties struct {}

func NewSamlAuthProperties() *SamlAuthProperties {
	return &SamlAuthProperties{}
}

func BindSamlAuthProperties(ctx *bootstrap.ApplicationContext) SamlAuthProperties {
	props := NewSamlAuthProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SamlAuthProperties"))
	}
	return *props
}
