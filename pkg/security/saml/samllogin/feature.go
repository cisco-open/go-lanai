package samllogin

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"

var (
	FeatureId       = security.FeatureId("saml_login", security.FeatureOrderSamlLogin)
	LogoutFeatureId = security.FeatureId("saml_logout", security.FeatureOrderLogout)
)

type Feature struct {
	id           security.FeatureIdentifier
	metadataPath string
	acsPath      string
	sloPath      string
	//The path to send the user to when authentication error is encountered
	errorPath      string
	successHandler security.AuthenticationSuccessHandler
	issuer         security.Issuer
}

func new(id security.FeatureIdentifier) *Feature {
	return &Feature{
		id:           id,
		metadataPath: "/saml/metadata",
		acsPath:      "/saml/SSO",
		sloPath:      "/saml/slo",
		errorPath:    "/error",
	}
}

func New() *Feature {
	return new(FeatureId)
}

func NewLogout() *Feature {
	return new(LogoutFeatureId)
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return f.id
}

func (f *Feature) Issuer(issuer security.Issuer) *Feature {
	f.issuer = issuer
	return f
}

func (f *Feature) ErrorPath(path string) *Feature {
	f.errorPath = path
	return f
}
