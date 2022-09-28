package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"net/url"
)

var (
	FeatureId    = security.FeatureId("SamlAuthorizeEndpoint", security.FeatureOrderSamlAuthorizeEndpoint)
	SloFeatureId = security.FeatureId("SamlSLOEndpoint", security.FeatureOrderSamlLogout)
)

type Feature struct {
	id           security.FeatureIdentifier
	ssoCondition web.RequestMatcher
	ssoLocation  *url.URL
	metadataPath string
	issuer      security.Issuer
	sloLocation string
}

// New Standard security.Feature entrypoint for authorization, DSL style. Used with security.WebSecurity
func New() *Feature {
	return &Feature{
		id: FeatureId,
	}
}

// NewLogout Standard security.Feature entrypoint for single-logout, DSL style. Used with security.WebSecurity
func NewLogout() *Feature {
	return &Feature{
		id: SloFeatureId,
	}
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return f.id
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

// EnableSLO when sloLocation is not empty, SLO Request handling is added
func (f *Feature) EnableSLO(sloLocation string) *Feature {
	f.sloLocation = sloLocation
	return f
}

func Configure(ws security.WebSecurity) *Feature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure saml authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

func ConfigureLogout(ws security.WebSecurity) *Feature {
	feature := NewLogout()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure saml authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}
