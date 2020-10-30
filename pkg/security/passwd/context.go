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
// UsernamePasswordAuthentication
// Note: all fields should not be used directly. It's exported only because gob only deal with exported field
type UsernamePasswordAuthentication struct {
	Account     security.Account
	PermissionList []string
}

func (auth *UsernamePasswordAuthentication) Principal() interface{} {
	return auth.Account
}

func (auth *UsernamePasswordAuthentication) Permissions() []string {
	return auth.PermissionList
}

func (auth *UsernamePasswordAuthentication) Authenticated() bool {
	return true
}

func (auth *UsernamePasswordAuthentication) Details() interface{} {
	return auth.Account
}