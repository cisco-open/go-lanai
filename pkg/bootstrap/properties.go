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

package bootstrap

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type Properties struct {
	Application ApplicationProperties `json:"application"`
	Cloud       CloudProperties       `json:"cloud"`
}

type ApplicationProperties struct {
	Name     string            `json:"name"`
	Profiles ProfileProperties `json:"profiles"`
}

type ProfileProperties struct {
	Active     utils.CommaSeparatedSlice `json:"active"`
	Additional utils.CommaSeparatedSlice `json:"additional"`
}

type CloudProperties struct {
	Gateway GatewayProperties `json:"gateway"`
}

type GatewayProperties struct {
	Service string `json:"service"`
	Scheme  string `json:"scheme"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}
