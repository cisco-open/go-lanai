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

package apilist

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "management.endpoint.apilist"
)

type FSType int

type Properties struct {
	StaticPath string `json:"static-path"`
}


//NewProperties create a Properties with default values
func NewProperties() *Properties {
	return &Properties{
		StaticPath: "configs/api-list.json",
	}
}

//BindProperties create and bind SessionProperties, with a optional prefix
func BindProperties(ctx *bootstrap.ApplicationContext) Properties {
	props := NewProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind Properties"))
	}
	return *props
}
