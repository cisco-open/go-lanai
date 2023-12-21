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

package consul

import "github.com/hashicorp/consul/api"

// ClientAuthentication
// TODO review ClientAuthentication and KubernetesClient
type ClientAuthentication interface {
	Login(client *api.Client) (token string, err error)
}

func newClientAuthentication(p *ConnectionProperties) ClientAuthentication {
	var clientAuthentication ClientAuthentication
	switch p.Authentication {
	case Kubernetes:
		clientAuthentication = TokenKubernetesAuthentication(p.Kubernetes)
	case Token:
		fallthrough
	default:
		clientAuthentication = TokenClientAuthentication(p.Token)
	}
	return clientAuthentication
}

type TokenClientAuthentication string

func (d TokenClientAuthentication) Login(_ *api.Client) (token string, err error) {
	return string(d), nil
}
