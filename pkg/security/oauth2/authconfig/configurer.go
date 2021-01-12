package authconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
)

var (
	OAuth2AuthorizeFeatureId = security.FeatureId("OAuth2Auth", security.FeatureOrderOAuth2TokenEndpoint)
	OAuth2TokenFeatureId = security.FeatureId("OAuth2Auth", security.FeatureOrderOAuth2AuthEndpoint)
)

//goland:noinspection GoNameStartsWithPackageName
type OAuth2AuthServerConfigurer struct {

}

func newOAuth2AuthConfigurer() *OAuth2AuthServerConfigurer {
	return &OAuth2AuthServerConfigurer{
	}
}

func (bac *OAuth2AuthServerConfigurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	//TODO
	tokenMw := auth.NewTokenEndpointMiddleware()

	tokenMapping := middleware.NewBuilder("otp verify").
		Order(security.MWOrderOAuth2Endpoints).
		Use(tokenMw.TokenHandlerFunc()).
		Build()

	ws.Add(tokenMapping)

	// add dummy
	ws.Add(mapping.Post("/v2/token").HandlerFunc(tokenMw.EndpointHandlerFunc()).Build())

	return nil
}