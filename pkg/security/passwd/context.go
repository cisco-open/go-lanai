package passwd

import "cto-github.cisco.com/livdu/jupiter/pkg/security"

/************************
	security.Candidate
************************/
// UsernamePasswordPair is the supported security.Candidate of this authenticator
type UsernamePasswordPair struct {
	Username string
	Password string
}

// security.Candidate
func (upp *UsernamePasswordPair) Principal() interface{} {
	return upp.Username
}

// security.Candidate
func (upp *UsernamePasswordPair) Credentials() interface{} {
	return upp.Password
}

// security.Candidate
func (upp *UsernamePasswordPair) Details() interface{} {
	return nil
}

/******************************
	security.Authentication
******************************/
type UsernamePasswordAuthentication struct {
	principal security.Account
	permissions []string
}

func (auth *UsernamePasswordAuthentication) Principal() interface{} {
	return auth.principal
}

func (auth *UsernamePasswordAuthentication) Permissions() []string {
	return auth.permissions
}

func (auth *UsernamePasswordAuthentication) Authenticated() bool {
	return true
}

func (auth *UsernamePasswordAuthentication) Details() interface{} {
	return auth.principal
}