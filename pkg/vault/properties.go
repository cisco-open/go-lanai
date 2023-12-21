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

package vault

import "fmt"

const (
	PropertyPrefix = "cloud.vault"
)

type ConnectionProperties struct {
	Host              string           `json:"host"`
	Port              int              `json:"port"`
	Scheme            string           `json:"scheme"`
	Authentication    AuthMethod       `json:"authentication"`
	SSL               SSLProperties    `json:"ssl"`
	Kubernetes        KubernetesConfig `json:"kubernetes"`
	Token             string           `json:"token"`
}

func (p ConnectionProperties) Address() string {
	return fmt.Sprintf("%s://%s:%d", p.Scheme, p.Host, p.Port)
}

type SSLProperties struct {
	CaCert     string `json:"ca-cert"`
	ClientCert string `json:"client-cert"`
	ClientKey  string `json:"client-key"`
	Insecure   bool   `json:"insecure"`
}

type KubernetesConfig struct {
	JWTPath string `json:"jwt-path"`
	Role    string `json:"role"`
}
