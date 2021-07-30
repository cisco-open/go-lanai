package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

var (
	FeatureId = security.FeatureId("OAuth2TokenAuth", security.FeatureOrderOAuth2TokenAuth)
)

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthConfigurer struct {
	tokenStoreReader oauth2.TokenStoreReader
}

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthOptions func(opt *TokenAuthOption)

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthOption struct {
	TokenStoreReader oauth2.TokenStoreReader
}

func NewTokenAuthConfigurer(opts ...TokenAuthOptions) *TokenAuthConfigurer {
	opt := TokenAuthOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &TokenAuthConfigurer{
		tokenStoreReader: opt.TokenStoreReader,
	}
}

func (c *TokenAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*TokenAuthFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	// configure other features
	errorhandling.Configure(ws).
		AdditionalErrorHandler(f.errorHandler)
	// TODO scope based access decesion maker

	// setup authenticator
	authenticator := NewAuthenticator(func(opt *AuthenticatorOption) {
		opt.TokenStoreReader = c.tokenStoreReader
	})
	ws.Authenticator().(*security.CompositeAuthenticator).Add(authenticator)

	// prepare middlewares
	successHanlder, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler)
	if !ok {
		successHanlder = security.NewAuthenticationSuccessHandler()
	}
	mw := NewTokenAuthMiddleware(func(opt *TokenAuthMWOption) {
		opt.Authenticator = ws.Authenticator()
		opt.SuccessHandler = successHanlder
		opt.PostBodyEnabled = f.postBodyEnabled
	})

	// install middlewares
	tokenAuth := middleware.NewBuilder("token authentication").
		Order(security.MWOrderOAuth2TokenAuth).
		Use(mw.AuthenticateHandlerFunc())

	ws.Add(tokenAuth)
	return nil
}

func (c *TokenAuthConfigurer) validate(f *TokenAuthFeature, _ security.WebSecurity) error {
	if c.tokenStoreReader == nil {
		return fmt.Errorf("token store reader is not pre-configured")
	}

	if f.errorHandler == nil {
		f.errorHandler = NewOAuth2ErrorHanlder()
	}

	//if f.granters == nil || len(f.granters) == 0 {
	//	return fmt.Errorf("token granters is not set")
	//}
	return nil
}



