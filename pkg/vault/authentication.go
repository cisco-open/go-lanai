package vault

const (
	TokenAuthentication = "token"
)

//ClientAuthentication interface represents a vault auth method https://www.vaultproject.io/docs/auth
type ClientAuthentication interface {
	Login() (token string, err error)
}

//TODO: token should be lease aware
type TokenClientAuthentication string

func (d TokenClientAuthentication) Login() (token string, err error){
	return string(d), nil
}