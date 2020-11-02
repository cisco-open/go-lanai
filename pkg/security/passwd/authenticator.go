package passwd

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"fmt"
)

/******************************
	security.Authenticator
******************************/
type Authenticator struct {
	store security.AccountStore
	passwdEncoder PasswordEncoder
}

func NewAuthenticator(store security.AccountStore, passwdEncoder PasswordEncoder) *Authenticator{
	if passwdEncoder == nil {
		passwdEncoder = NewNoopPasswordEncoder()
	}
	return &Authenticator{
		store: store,
		passwdEncoder: passwdEncoder,
	}
}

func (a *Authenticator) Authenticate(candidate security.Candidate) (security.Authentication, error) {
	upp, ok := candidate.(*UsernamePasswordPair)
	if !ok {
		return nil, nil
	}

	// Search user in the slice of allowed credentials
	user, err := a.store.LoadAccountByUsername(upp.Username)
	if err != nil {
		return nil, security.NewUsernameNotFoundError(fmt.Sprintf("Cannot find account with username %v", upp.Principal()))
	}

	// TODO check account status

	// Check password
	if upp.Username != user.Username() || !a.passwdEncoder.Matches(upp.Password, user.Password()) {
		return nil, security.NewBadCredentialsError("incorrect credentials")
	}

	// TODO post password check

	auth := usernamePasswordAuthentication{
		Account: user,
		PermissionList: []string {
			//TODO
			"authenticated",
		},
	}
	return &auth, nil
}
