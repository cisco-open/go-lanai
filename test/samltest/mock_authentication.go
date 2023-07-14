package samltest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/crewjam/saml"
)

type MockedSamlAssertionAuthentication struct {
	Account       security.Account
	DetailsMap    map[string]interface{}
	SamlAssertion *saml.Assertion
}

func (sa *MockedSamlAssertionAuthentication) Principal() interface{} {
	return sa.Account
}

func (sa *MockedSamlAssertionAuthentication) Permissions() security.Permissions {
	perms := security.Permissions{}
	for _, perm := range sa.Account.Permissions() {
		perms[perm] = struct{}{}
	}
	return perms
}

func (sa *MockedSamlAssertionAuthentication) State() security.AuthenticationState {
	return security.StateAuthenticated
}

func (sa *MockedSamlAssertionAuthentication) Details() interface{} {
	return sa.DetailsMap
}

func (sa *MockedSamlAssertionAuthentication) Assertion() *saml.Assertion {
	return sa.SamlAssertion
}
