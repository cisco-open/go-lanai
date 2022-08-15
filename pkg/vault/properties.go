package vault

import "fmt"

const (
	PropertyPrefix = "cloud.vault"
)

type ConnectionProperties struct {
	Host           string        `json:"host"`
	Port           int           `json:"port"`
	Scheme         string        `json:"scheme"`
	Authentication AuthMethod    `json:"authentication"`
	Ssl            SslProperties `json:"ssl"`
	TokenSource    TokenSource   `json:"token-source"`
	Token          string        `json:"token"`
}

func (p *ConnectionProperties) Address() string {
	return fmt.Sprintf("%s://%s:%d", p.Scheme, p.Host, p.Port)
}

type SslProperties struct {
	Cacert     string `json:"ca-cert"`
	ClientCert string `json:"apiClient-cert"`
	ClientKey  string `json:"apiClient-key"`
	Insecure   bool   `json:"insecure"`
}

type TokenSource struct {
	Kubernetes KubernetesConfig `json:"kubernetes"`
}

type KubernetesConfig struct {
	JWTPath string `json:"jwt-path"`
	Role    string `json:"role"`
}
