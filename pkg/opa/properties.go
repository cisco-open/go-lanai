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

package opa

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

const PropertiesPrefix = "security.opa"

type Properties struct {
	Server  BundleServerProperties            `json:"server"`
	Bundles map[string]BundleSourceProperties `json:"bundles"`
	Logging LoggingProperties                 `json:"logging"`
}

type BundleServerProperties struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	PollingProperties
}

type BundleSourceProperties struct {
	Path string `json:"path"`
	PollingProperties
}

type LoggingProperties struct {
	LogLevel          log.LoggingLevel `json:"level"`
	DecisionLogsLevel log.LoggingLevel `json:"decision-logs-level"`
}

type PollingProperties struct {
	PollingMinDelay    *utils.Duration `json:"polling-min-delay,omitempty"`    // min amount of time to wait between successful poll attempts
	PollingMaxDelay    *utils.Duration `json:"polling-max-delay,omitempty"`    // max amount of time to wait between poll attempts
	LongPollingTimeout *utils.Duration `json:"long-polling-timeout,omitempty"` // max amount of time the server should wait before issuing a timeout if there's no update available
}

func NewProperties() *Properties {
	return &Properties{}
}
