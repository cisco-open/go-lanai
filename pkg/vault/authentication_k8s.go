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
