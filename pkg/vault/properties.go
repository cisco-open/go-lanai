package vault

import "fmt"

const (
	PropertyPrefix = "cloud.vault"
)

type ConnectionProperties struct {
	Enabled     string `json:"enabled"` //TODO
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Scheme      string `json:"scheme"`
	Authentication string `json:"authentication"`
	Ssl			SslProperties `json:"ssl"`
	Token 		string `json:"token"`
}

func (p *ConnectionProperties) Address() string {
	return fmt.Sprintf("%s://%s:%d", p.Scheme, p.Host, p.Port)
}

type SslProperties struct {
	Cacert     string `json:"ca-cert"`
	ClientCert string `json:"client-cert"`
	ClientKey  string `json:"client-key"`
	Insecure   bool   `json:"insecure"`
}