package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type OAuth2AuthFeature struct {
	// TODO we may want to override authenticator and other stuff
}

// Standard security.Feature entrypoint
func (f *OAuth2AuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *OAuth2AuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*OAuth2AuthFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authorization: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *OAuth2AuthFeature {
	return &OAuth2AuthFeature{}
}
