package token

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type TokenFeature struct {
	path string
	granters []auth.TokenGranter
}

// Standard security.Feature entrypoint
func (f *TokenFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *TokenFeature {
	feature := NewEndpoint()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*TokenFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authconfig: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func NewEndpoint() *TokenFeature {
	return &TokenFeature{
	}
}

/** Setters **/
func (f *TokenFeature) Path(path string) *TokenFeature {
	f.path = path
	return f
}

func (f *TokenFeature) AddGranter(granter auth.TokenGranter) *TokenFeature {
	if composite, ok := granter.(*auth.CompositeTokenGranter); ok {
		f.granters = append(f.granters, composite.Delegates()...)
	} else {
		f.granters = append(f.granters, granter)
	}

	return f
}