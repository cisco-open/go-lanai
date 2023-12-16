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

func (d TokenClientAuthentication) Login(client *api.Client) (token string, err error) {
	return string(d), nil
}
