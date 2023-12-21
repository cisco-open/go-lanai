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

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"os"
)

type KubernetesClient struct {
	config KubernetesConfig
}

func (c *KubernetesClient) Login(client *api.Client) (string, error) {
	// defaults to using /var/run/secrets/kubernetes.io/serviceaccount/token if no options set
	if c.config.JWTPath == "" {
		c.config.JWTPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}
	jwtToken, err := readTokenFromFile(c.config.JWTPath)
	if err != nil {
		return "", err
	}
	options := &api.ACLLoginParams{
		AuthMethod:  c.config.Method,
		BearerToken: jwtToken,
	}
	authToken, _, err := client.ACL().Login(options, nil)
	if err != nil {
		return "", err
	}
	logger.Info("Successfully obtained Consul token using k8s auth")
	return authToken.SecretID, nil
}

func TokenKubernetesAuthentication(kubernetesConfig KubernetesConfig) *KubernetesClient {
	return &KubernetesClient{
		config: kubernetesConfig,
	}
}

func readTokenFromFile(filepath string) (string, error) {
	jwt, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("unable to read file containing service account token: %w", err)
	}
	return string(jwt), nil
}
