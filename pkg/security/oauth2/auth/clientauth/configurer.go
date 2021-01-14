package clientauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
)

var (
	FeatureId = security.FeatureId("OAuth2ClientAuth", security.FeatureOrderOAuth2ClientAuth)
)

//goland:noinspection GoNameStartsWithPackageName
type ClientAuthConfigurer struct {
}

func newClientAuthConfigurer() *ClientAuthConfigurer {
	return &ClientAuthConfigurer{
	}
}

func (c *ClientAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*ClientAuthFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	passwd.Configure(ws).
		AccountStore(c.clientAccountStore(f)).
		PasswordEncoder(f.clientSecretEncoder).
		MFA(false)
	// no entry point, everything handled by access denied handler
	basicauth.Configure(ws).
		EntryPoint(nil)
	access.Configure(ws).
		Request(matcher.AnyRequest()).
		Authenticated()
	errorhandling.Configure(ws).
		AccessDeniedHandler(f.errorHandler).
		AuthenticationErrorHandler(f.errorHandler)

	return nil
}

func (c *ClientAuthConfigurer) validate(f *ClientAuthFeature, ws security.WebSecurity) error {
	if f.clientStore == nil {
		return fmt.Errorf("client store for client authentication is not set")
	}

	if f.clientSecretEncoder == nil {
		f.clientSecretEncoder = passwd.NewNoopPasswordEncoder()
	}

	if f.errorHandler == nil {
		f.errorHandler = auth.NewOAuth2ErrorHanlder()
	}
	return nil
}

func (c *ClientAuthConfigurer) clientAccountStore(f *ClientAuthFeature) *auth.OAuth2ClientAccountStore {
	return auth.WrapOAuth2ClientStore(f.clientStore)
}



