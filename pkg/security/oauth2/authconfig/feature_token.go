package authconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type OAuth2TokenEndpointFeature struct {
	// TODO we may want to override authenticator and other stuff
}

// Standard security.Feature entrypoint
func (f *OAuth2TokenEndpointFeature) Identifier() security.FeatureIdentifier {
	return OAuth2TokenFeatureId
}

func Configure(ws security.WebSecurity) *OAuth2TokenEndpointFeature {
	feature := NewTokenEndpoint()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*OAuth2TokenEndpointFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authconfig: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func NewTokenEndpoint() *OAuth2TokenEndpointFeature {
	return &OAuth2TokenEndpointFeature{}
}
