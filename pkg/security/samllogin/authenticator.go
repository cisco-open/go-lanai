package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/crewjam/saml"
)

type AssertionCandidate struct {
	Assertion *saml.Assertion
	DetailsMap map[interface{}]interface{}
}

func (a *AssertionCandidate) Principal() interface{} {
	if a.Assertion.Subject == nil || a.Assertion.Subject.NameID == nil {
		return nil
	}
	return a.Assertion.Subject.NameID.Value
}

func (a *AssertionCandidate) Credentials() interface{} {
	return a.Assertion
}

func (a *AssertionCandidate) Details() interface{} {
	return a.DetailsMap
}

type samlAssertionAuthentication struct {
	Acct       security.Account
	Perms      map[string]interface{}
	DetailsMap map[interface{}]interface{}
}

func (sa *samlAssertionAuthentication) Principal() interface{} {
	return sa.Acct
}

func (sa *samlAssertionAuthentication)  Permissions() map[string]interface{} {
	return sa.Perms
}

func (sa *samlAssertionAuthentication)  State() security.AuthenticationState {
	return security.StateAuthenticated
}

func (sa *samlAssertionAuthentication)  Details() interface{} {
	return sa.DetailsMap
}



type Authenticator struct {
	accountStore      security.FederatedAccountStore
	idpManager IdentityProviderManager
}

func (a *Authenticator) Authenticate(_ context.Context, candidate security.Candidate) (security.Authentication, error) {
	assertionCandidate, ok := candidate.(*AssertionCandidate)
	if !ok {
		return nil, nil
	}

	idp := a.idpManager.GetIdentityProviderByEntityId(assertionCandidate.Assertion.Issuer.Value)

	user, err := a.accountStore.LoadAccountByExternalId(idp.ExternalIdName, assertionCandidate.Principal().(string), idp.ExternalIdpName)

	if err != nil {
		return nil, security.NewInternalAuthenticationError("Couldn't load federated account", err)
	}

	permissions := map[string]interface{}{}
	for _,p := range user.Permissions() {
		permissions[p] = true
	}

	details := assertionCandidate.DetailsMap
	if details == nil {
		details = map[interface{}]interface{}{}
	}

	auth := &samlAssertionAuthentication{
		Acct: user,
		Perms: permissions,
		DetailsMap: details,
	}
	return auth, nil
}
