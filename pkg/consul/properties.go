package consul

import "fmt"

type ConnectionProperties struct {
	Host           string           `json:"host"`
	Port           int              `json:"port"`
	Scheme         string        `json:"scheme"`
	SSL            SSLProperties `json:"ssl"`
	Authentication AuthMethod    `json:"authentication"`
	Kubernetes     KubernetesConfig `json:"kubernetes"`
	Token          string           `json:"token"`
}

func (c ConnectionProperties) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type SSLProperties struct {
	CaCert     string `json:"ca-cert"`
	ClientCert string `json:"client-cert"`
	ClientKey  string `json:"client-key"`
	Insecure   bool   `json:"insecure"`
}

type KubernetesConfig struct {
	JWTPath string `json:"jwt-path"`
	Method  string `json:"method"`
}
