package clientauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type ClientAuthFeature struct {
	clientStore         oauth2.OAuth2ClientStore
	clientSecretEncoder passwd.PasswordEncoder
	errorHandler        *auth.OAuth2ErrorHandler
}

// Standard security.Feature entrypoint
func (f *ClientAuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *ClientAuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*ClientAuthFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authconfig: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *ClientAuthFeature {
	return &ClientAuthFeature{
	}
}

/** Setters **/
func (f *ClientAuthFeature) ClientStore(clientStore oauth2.OAuth2ClientStore) *ClientAuthFeature {
	f.clientStore = clientStore
	return f
}

func (f *ClientAuthFeature) ClientSecretEncoder(clientSecretEncoder passwd.PasswordEncoder) *ClientAuthFeature {
	f.clientSecretEncoder = clientSecretEncoder
	return f
}

func (f *ClientAuthFeature) ErrorHandler(errorHandler *auth.OAuth2ErrorHandler) *ClientAuthFeature {
	f.errorHandler = errorHandler
	return f
}