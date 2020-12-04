package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding/gob"
)

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
	Serialization
******************************/
func GobRegister() {
	gob.Register((*usernamePasswordAuthentication)(nil))
}

/******************************
	security.Authentication
******************************/
type UsernamePasswordAuthentication interface {
	security.Authentication
	IsUsernamePasswordAuthentication() bool
}

// usernamePasswordAuthentication
// Note: all fields should not be used directly. It's exported only because gob only deal with exported field
type usernamePasswordAuthentication struct {
	Account     security.Account
	PermissionList []string
}

func (auth *usernamePasswordAuthentication) Principal() interface{} {
	return auth.Account
}

func (auth *usernamePasswordAuthentication) Permissions() []string {
	return auth.PermissionList
}

func (auth *usernamePasswordAuthentication) Authenticated() bool {
	return true
}

func (auth *usernamePasswordAuthentication) Details() interface{} {
	return auth.Account
}

func (auth *usernamePasswordAuthentication) IsUsernamePasswordAuthentication() bool {
	return true
}

func IsSamePrincipal(username string, currentAuth security.Authentication) bool {
	if currentAuth == nil || !currentAuth.Authenticated() {
		return false
	}

	if account, ok := currentAuth.Principal().(security.Account); ok && username == account.Username() {
		return true
	} else if principal, ok := currentAuth.Principal().(string); ok && username == principal {
		return true
	}
	return false
}