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
