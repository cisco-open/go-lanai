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
