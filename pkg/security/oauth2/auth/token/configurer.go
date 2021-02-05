package token

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

var (
	FeatureId = security.FeatureId("OAuth2AuthToken", security.FeatureOrderOAuth2TokenEndpoint)
)

//goland:noinspection GoNameStartsWithPackageName
type TokenEndpointConfigurer struct {
}

func newOAuth2TokenEndpointConfigurer() *TokenEndpointConfigurer {
	return &TokenEndpointConfigurer{
	}
}

func (c *TokenEndpointConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*TokenFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	//TODO prepare middlewares
	tokenMw := NewTokenEndpointMiddleware(func(opts *TokenEndpointOptions) {
		opts.Granter = auth.NewCompositeTokenGranter(f.granters...)
	})

	// install middlewares
	tokenMapping := middleware.NewBuilder("token endpoint").
		ApplyTo(matcher.RouteWithPattern(f.path, http.MethodPost)).
		Order(security.MWOrderOAuth2Endpoints).
		Use(tokenMw.TokenHandlerFunc())

	ws.Add(tokenMapping)

	// add dummy handler
	ws.Add(mapping.Post(f.path).HandlerFunc(security.NoopHandlerFunc))

	return nil
}

func (c *TokenEndpointConfigurer) validate(f *TokenFeature, ws security.WebSecurity) error {
	if f.path == "" {
		return fmt.Errorf("token endpoint is not set")
	}

	if f.granters == nil || len(f.granters) == 0 {
		return fmt.Errorf("token granters is not set")
	}
	return nil
}



