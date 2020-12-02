package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"net/http"
	"sort"
)

/*****************************
	Abstraction
 *****************************/
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

// AuthenticationSuccessHandler handles authentication success event
// The counterpart of this interface is AuthenticationErrorHandler
type AuthenticationSuccessHandler interface {
	HandleAuthenticationSuccess(context.Context, *http.Request, http.ResponseWriter, Authentication)
}

/*****************************
	Common Impl.
 *****************************/
type CompositeAuthenticator struct {
	authenticators []Authenticator
}

func NewAuthenticator(authenticators ...Authenticator) Authenticator {
	return &CompositeAuthenticator {
		authenticators: authenticators,
	}
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

func (a *CompositeAuthenticator) Add(authenticator Authenticator) *CompositeAuthenticator {
	a.authenticators = append(a.authenticators, authenticator)
	return a
}

func (a *CompositeAuthenticator) Merge(composite *CompositeAuthenticator) *CompositeAuthenticator {
	a.authenticators = append(a.authenticators, composite.authenticators...)
	sort.Slice(a.authenticators, func(i,j int) bool {
		return order.OrderedFirstCompare(a.authenticators[i], a.authenticators[j])
	})
	return a
}

