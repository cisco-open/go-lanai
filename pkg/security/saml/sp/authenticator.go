package sp

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"github.com/crewjam/saml"
)

type AssertionCandidate struct {
	Assertion  *saml.Assertion
	DetailsMap map[string]interface{}
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

type SamlAssertionAuthentication interface {
	security.Authentication
	Assertion() *saml.Assertion
}

type samlAssertionAuthentication struct {
	Account       security.Account
	Perms         map[string]interface{}
	DetailsMap    map[string]interface{}
	SamlAssertion *saml.Assertion
}

func (sa *samlAssertionAuthentication) Principal() interface{} {
	return sa.Account
}

func (sa *samlAssertionAuthentication) Permissions() security.Permissions {
	return sa.Perms
}

func (sa *samlAssertionAuthentication) State() security.AuthenticationState {
	return security.StateAuthenticated
}

func (sa *samlAssertionAuthentication) Details() interface{} {
	return sa.DetailsMap
}

func (sa *samlAssertionAuthentication) Assertion() *saml.Assertion {
	return sa.SamlAssertion
}

type Authenticator struct {
	accountStore security.FederatedAccountStore
	idpManager   samlctx.SamlIdentityProviderManager
}

func (a *Authenticator) Authenticate(ctx context.Context, candidate security.Candidate) (security.Authentication, error) {
	assertionCandidate, ok := candidate.(*AssertionCandidate)
	if !ok {
		return nil, nil
	}

	idp, err := a.idpManager.GetIdentityProviderByEntityId(ctx, assertionCandidate.Assertion.Issuer.Value)
	if err != nil {
		return nil, security.NewInternalAuthenticationError("Couldn't find idp matching the assertion")
	}
	samlIdp, ok := idp.(samlctx.SamlIdentityProvider)
	if !ok {
		return nil, security.NewInternalAuthenticationError("Couldn't find idp metadata matching the assertion")
	}

	user, err := a.accountStore.LoadAccountByExternalId(ctx, samlIdp.ExternalIdName(), assertionCandidate.Principal().(string), samlIdp.ExternalIdpName(), samlIdp.GetAutoCreateUserDetails(), assertionCandidate.Assertion)

	if err != nil {
		return nil, security.NewInternalAuthenticationError(err)
	}

	if user.Disabled() {
		return nil, security.NewAccountStatusError("Account Disabled")
	}

	permissions := map[string]interface{}{}
	for _, p := range user.Permissions() {
		permissions[p] = true
	}

	details := assertionCandidate.DetailsMap
	if details == nil {
		details = make(map[string]interface{})
	}
	details[security.DetailsKeyAuthTime] = assertionCandidate.Assertion.IssueInstant
	details[security.DetailsKeyAuthMethod] = security.AuthMethodExternalSaml

	auth := &samlAssertionAuthentication{
		Account:       user,
		SamlAssertion: assertionCandidate.Assertion,
		Perms:         permissions,
		DetailsMap:    details,
	}
	return auth, nil
}
