package security

import "fmt"

type Candidate interface {
	Principal() interface{}
	Credentials() interface{}
	Details() interface{}
}

type Authenticator interface {
	// Authenticate function takes the Candidate and authenticate it.
	// if the Candidate type is not supported, return nil,nil
	// if the Candidate is rejected, non-nil error, and the returned Authentication is ignored
	Authenticate(Candidate) (Authentication, error)
}

type CompositeAuthenticator []Authenticator

func NewAuthenticator(authenticators ...Authenticator) Authenticator {
	return CompositeAuthenticator(authenticators)
}

func (a CompositeAuthenticator) Authenticate(candidate Candidate) (auth Authentication, err error) {
	for _,authenticator := range a {
		auth, err = authenticator.Authenticate(candidate)
		if auth != nil || err != nil {
			return
		}
	}
	return nil, NewAuthenticatorNotAvailableError(fmt.Sprintf("unable to find authenticator for cadidate %T", candidate))
}