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

import (
	"context"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/kubernetes"
)

type KubernetesClient struct {
	config KubernetesConfig
}

func (c *KubernetesClient) Login(client *api.Client) (string, error) {
	var options []kubernetes.LoginOption
	// defaults to using /var/run/secrets/kubernetes.io/serviceaccount/token if no options set
	if c.config.JWTPath != "" {
		options = append(options, kubernetes.WithServiceAccountTokenPath(c.config.JWTPath))
	}
	k8sAuth, err := kubernetes.NewKubernetesAuth(
		c.config.Role,
		options...,
	)
	if err != nil {
		return "", err
	}
	authInfo, err := client.Auth().Login(context.Background(), k8sAuth)
	if err != nil {
		return "", err
	}

	return authInfo.Auth.ClientToken, nil
}

func TokenKubernetesAuthentication(kubernetesConfig KubernetesConfig) *KubernetesClient {
	return &KubernetesClient{
		config: kubernetesConfig,
	}
}
