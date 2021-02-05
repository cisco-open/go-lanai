package samllogin

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"

var (
	FeatureId = security.FeatureId("saml_login", security.FeatureOrderSamlLogin)
)

type Feature struct {
	metadataPath   string
	acsPath        string
	sloPath        string
	//The path to send the user to when authentication error is encountered
	errorPath      string
	successHandler security.AuthenticationSuccessHandler
}

func New() *Feature {
	return &Feature{
		metadataPath: "/saml/metadata",
		acsPath:      "/saml/SSO",
		sloPath:      "/saml/slo",
		errorPath: 	  "/error",
	}
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}