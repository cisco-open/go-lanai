package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

var (
	FeatureId = security.FeatureId("OAuth2AuthorizeEndpoint", security.FeatureOrderOAuth2AuthorizeEndpoint)
)

//goland:noinspection GoNameStartsWithPackageName
type AuthorizeEndpointConfigurer struct {
}

func newOAuth2AuhtorizeEndpointConfigurer() *AuthorizeEndpointConfigurer {
	return &AuthorizeEndpointConfigurer{
	}
}

func (c *AuthorizeEndpointConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*AuthorizeFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	// configure other features
	errorhandling.Configure(ws).
		AdditionalErrorHandler(f.errorHandler)

	//TODO prepare middlewares
	authorizeMW := NewTokenEndpointMiddleware(func(opts *AuthorizeMWOption) {
		opts.RequestProcessor = f.requestProcessor
	})

	// install middlewares
	preAuth := middleware.NewBuilder("authorize validation").
		ApplyTo(matcher.RouteWithPattern(f.path, http.MethodGet, http.MethodPost)).
		Order(security.MWOrderOAuth2AuthValidation).
		Use(authorizeMW.PreAuthenticateHandlerFunc())

	ep := middleware.NewBuilder("authorize endpoint").
		ApplyTo(matcher.RouteWithPattern(f.path, http.MethodGet, http.MethodPost)).
		Order(security.MWOrderOAuth2Endpoints).
		Use(authorizeMW.AuthroizeHandlerFunc())

	ws.Add(preAuth, ep)

	// add dummy handler
	ws.Add(mapping.Get(f.path).HandlerFunc(security.NoopHandlerFunc))
	ws.Add(mapping.Post(f.path).HandlerFunc(security.NoopHandlerFunc))

	return nil
}

func (c *AuthorizeEndpointConfigurer) validate(f *AuthorizeFeature, ws security.WebSecurity) error {
	if f.path == "" {
		return fmt.Errorf("token endpoint is not set")
	}

	if f.errorHandler == nil {
		f.errorHandler = auth.NewOAuth2ErrorHanlder()
	}

	//if f.granters == nil || len(f.granters) == 0 {
	//	return fmt.Errorf("token granters is not set")
	//}
	return nil
}



