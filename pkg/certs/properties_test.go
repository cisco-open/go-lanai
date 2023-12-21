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

package certs_test

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

// No test cases here, just bunch of test properties. See Manager tests

type TestRedisProperties struct {
	Address string            `json:"addrs"`
	TLS     TestTLSProperties `json:"tls"`
}

type TestDataProperties struct {
	Host string            `json:"host"`
	Port int               `json:"port"`
	TLS  TestTLSProperties `json:"tls"`
}

type TestKafkaProperties struct {
	Brokers utils.CommaSeparatedSlice `json:"brokers"`
	TLS     TestTLSProperties         `json:"tls"`
}

type TestTLSProperties struct {
	Enabled bool                   `json:"enabled"`
	Certs   certs.SourceProperties `json:"certs"`
}

type TestVaultSourceProperties struct {
	MinTLSVersion    string         `json:"min-version"`
	Path             string         `json:"path"`
	Role             string         `json:"role"`
	CN               string         `json:"cn"`
	IpSans           string         `json:"ip-sans"`
	AltNames         string         `json:"alt-names"`
	TTL              utils.Duration `json:"ttl"`
	MinRenewInterval utils.Duration `json:"min-renew-interval"`
	CachePath        string         `json:"cache-path"`
}

type TestFileSourceProperties struct {
	MinTLSVersion string `json:"min-version"`
	CACertFile    string `json:"ca-cert-file"`
	CertFile      string `json:"cert-file"`
	KeyFile       string `json:"key-file"`
	KeyPass       string `json:"key-pass"`
}
