package vault

import "github.com/hashicorp/vault/api"

//ClientAuthentication interface represents a vault auth method https://www.vaultproject.io/docs/auth
type ClientAuthentication interface {
	Login(client *api.Client) (token string, err error)
}

func newClientAuthentication(p *ConnectionProperties) ClientAuthentication {
	var clientAuthentication ClientAuthentication
	switch p.TokenSource.Source {
	case Kubernetes:
		clientAuthentication = TokenKubernetesAuthentication(p.TokenSource.Kubernetes)
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
