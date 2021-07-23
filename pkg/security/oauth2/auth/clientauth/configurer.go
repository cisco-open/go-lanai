package clientauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
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

	// configure other features
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
		AdditionalErrorHandler(f.errorHandler)

	// add middleware to translate authentication error to oauth2 error
	mw := NewClientAuthMiddleware(func(opt *ClientAuthMWOption) {
		opt.Authenticator = ws.Authenticator()
		opt.SuccessHandler = ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler)
	})
	ws.Add(middleware.NewBuilder("client auth error translator").
		Order(security.MWOrderPreAuth).
		Use(mw.ErrorTranslationHandlerFunc()),
	)

	// add middleware to support form based client auth
	if f.allowForm {
		ws.Add(middleware.NewBuilder("form client auth").
			Order(security.MWOrderFormAuth).
			Use(mw.ClientAuthFormHandlerFunc()),
		)
	}

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
		f.errorHandler = auth.NewOAuth2ErrorHandler()
	}
	return nil
}

func (c *ClientAuthConfigurer) clientAccountStore(f *ClientAuthFeature) *auth.OAuth2ClientAccountStore {
	return auth.WrapOAuth2ClientStore(f.clientStore)
}



