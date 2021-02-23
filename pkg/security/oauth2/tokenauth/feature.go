package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type TokenAuthFeature struct {
	errorHandler     *OAuth2ErrorHandler
}

// Standard security.Feature entrypoint
func (f *TokenAuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *TokenAuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*TokenAuthFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *TokenAuthFeature {
	return &TokenAuthFeature{
	}
}

/** Setters **/
func (f *TokenAuthFeature) ErrorHandler(errorHandler *OAuth2ErrorHandler) *TokenAuthFeature {
	f.errorHandler = errorHandler
	return f
}
