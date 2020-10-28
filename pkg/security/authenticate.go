package security

import "fmt"

type Authenticator interface {
	// Authenticate function takes the Candidate and authenticate it.
	// if the Candidate type is not supported, return nil,nil
	// if the Candidate is rejected, non-nil error, and the returned Authentication is ignored
	Authenticate(Candidate) (Authentication,error)
}

type CompositeAuthenticator struct {
	authenticators []Authenticator
}

func (a *CompositeAuthenticator) Authenticate(candidate Candidate) (auth Authentication, err error) {
	for _,authenticator := range a.authenticators {
		auth, err = authenticator.Authenticate(candidate)
		if auth != nil || err != nil {
			return
		}
	}
	return nil, NewAuthenticatorNotAvailableError(fmt.Sprintf("unable to find authenticator for cadidate %T", candidate))
}