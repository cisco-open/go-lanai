package consul

import "fmt"

type ConnectionProperties struct {
	Host           string           `json:"host"`
	Port           int              `json:"port"`
	Scheme         string           `json:"scheme"`
	Ssl            SslProperties    `json:"ssl"`
	Authentication AuthMethod       `json:"authentication"`
	Kubernetes     KubernetesConfig `json:"kubernetes"`
	Token          string           `json:"token"`
}

func (c ConnectionProperties) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type SslProperties struct {
	Cacert     string `json:"ca-cert"`
	ClientCert string `json:"apiClient-cert"`
	ClientKey  string `json:"apiClient-key"`
	Insecure   bool   `json:"insecure"`
}

type KubernetesConfig struct {
	JWTPath string `json:"jwt-path"`
	Method  string `json:"method"`
}
