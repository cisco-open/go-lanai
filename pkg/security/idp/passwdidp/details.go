package passwdidp

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"

type PasswdIdpDetails struct {
	Domain           string
}


type PasswdIdpOptions func(opt *PasswdIdpDetails)

// PasswdIdentityProvider implements idp.IdentityProvider and idp.AuthenticationFlowAware
type PasswdIdentityProvider struct {
	PasswdIdpDetails
}

func NewIdentityProvider(opts ...PasswdIdpOptions) *PasswdIdentityProvider {
	opt := PasswdIdpDetails{}
	for _, f := range opts {
		f(&opt)
	}
	return &PasswdIdentityProvider{
		PasswdIdpDetails: opt,
	}
}

func (s PasswdIdentityProvider) AuthenticationFlow() idp.AuthenticationFlow {
	return idp.InternalIdpForm
}

func (s PasswdIdentityProvider) Domain() string {
	return s.PasswdIdpDetails.Domain
}
