package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
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
	authRouteMatcher := matcher.RouteWithPattern(f.path, http.MethodGet, http.MethodPost)
	approveRouteMatcher := matcher.RouteWithPattern(f.approvalPath, http.MethodPost)
	approveRequestMatcher := matcher.RequestWithPattern(f.approvalPath, http.MethodPost).
		And(matcher.RequestHasPostParameter(oauth2.ParameterUserApproval))

	authorizeMW := NewAuthorizeEndpointMiddleware(func(opts *AuthorizeMWOption) {
		opts.RequestProcessor = f.requestProcessor
		opts.AuthorizeHanlder = f.authorizeHanlder
		opts.ApprovalMatcher = approveRequestMatcher
	})

	// install middlewares
	preAuth := middleware.NewBuilder("authorize validation").
		ApplyTo(authRouteMatcher.Or(approveRouteMatcher)).
		Order(security.MWOrderOAuth2AuthValidation).
		Use(authorizeMW.PreAuthenticateHandlerFunc())

	ws.Add(preAuth)

	// install authorize endpoint
	epGet := mapping.Get(f.path).Name("authorize GET").
		HandlerFunc(authorizeMW.AuthroizeHandlerFunc())
	epPost := mapping.Post(f.path).Name("authorize Post").
		HandlerFunc(authorizeMW.AuthroizeHandlerFunc())

	ws.Route(authRouteMatcher).Add(epGet, epPost)

	// install approve endpoint
	approve := mapping.Post(f.approvalPath).Name("approve endpoint").
		HandlerFunc(authorizeMW.ApproveOrDenyHandlerFunc())

	ws.Route(approveRouteMatcher).Add(approve)

	return nil
}

func (c *AuthorizeEndpointConfigurer) validate(f *AuthorizeFeature, ws security.WebSecurity) error {
	if f.path == "" {
		return fmt.Errorf("authorize endpoint path is not set")
	}

	if f.errorHandler == nil {
		f.errorHandler = auth.NewOAuth2ErrorHanlder()
	}

	if f.authorizeHanlder == nil {
		return fmt.Errorf("auhtorize handler is not set")
	}

	//if f.granters == nil || len(f.granters) == 0 {
	//	return fmt.Errorf("token granters is not set")
	//}
	return nil
}



