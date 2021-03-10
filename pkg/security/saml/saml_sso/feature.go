package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"net/url"
)

type Feature struct {
	ssoCondition web.RequestMatcher
	ssoLocation  *url.URL
	metadataPath string
	issuer       security.Issuer
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func NewEndpoint() *Feature {
	return &Feature{}
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *Feature) SsoCondition(condition web.RequestMatcher) *Feature {
	f.ssoCondition = condition
	return f
}

func (f *Feature) SsoLocation(location *url.URL) *Feature {
	f.ssoLocation = location
	return f
}

func (f *Feature) MetadataPath(path string) *Feature {
	f.metadataPath = path
	return f
}

func (f *Feature) Issuer(issuer security.Issuer) *Feature {
	f.issuer = issuer
	return f
}

func Configure(ws security.WebSecurity) *Feature {
	feature := NewEndpoint()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure saml authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}