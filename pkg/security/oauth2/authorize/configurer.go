package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

var (
	FeatureId = security.FeatureId("OAuth2Auth", security.FeatureOrderOAuth2Auth)
)

//goland:noinspection GoNameStartsWithPackageName
type OAuth2AuthServerConfigurer struct {

}

func newOAuth2AuthConfigurer() *OAuth2AuthServerConfigurer {
	return &OAuth2AuthServerConfigurer{
	}
}

func (bac *OAuth2AuthServerConfigurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	return nil
}