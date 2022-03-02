package vault

//ClientAuthentication interface represents a vault auth method https://www.vaultproject.io/docs/auth
type ClientAuthentication interface {
	Login() (token string, err error)
}

func newClientAuthentication(p *ConnectionProperties) ClientAuthentication {
	var clientAuthentication ClientAuthentication
	switch p.Authentication {
	case Token:
		fallthrough
	default:
		clientAuthentication = TokenClientAuthentication(p.Token)
	}
	return clientAuthentication
}

type TokenClientAuthentication string

func (d TokenClientAuthentication) Login() (token string, err error){
	return string(d), nil
}

