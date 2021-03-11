package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/authorize"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/clientauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/token"
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

/***************************
	addtional abstractions
 ***************************/
type IdpSecurityConfigurer interface {
	Configure(ws security.WebSecurity, config *Configuration)
}

/***************************
	security configurers
 ***************************/
// ClientAuthEndpointsConfigurer implements security.Configurer
// responsible to configure misc using client auth
type ClientAuthEndpointsConfigurer struct {
	config *Configuration
}

func (c *ClientAuthEndpointsConfigurer) Configure(ws security.WebSecurity) {
	// For Token endpoint
	ws.Route(matcher.RouteWithPattern(c.config.Endpoints.Token)).
		Route(matcher.RouteWithPattern(c.config.Endpoints.CheckToken)).
		With(clientauth.New().
			ClientStore(c.config.ClientStore).
			ClientSecretEncoder(c.config.clientSecretEncoder()).
			ErrorHandler(c.config.errorHandler()).
			AllowForm(true), // AllowForm also implicitly enables Public Client
		).
		// uncomment following if we want CheckToken to not allow public client
		//With(access.Configure(ws).
		//	Request(matcher.RequestWithPattern(c.config.Endpoints.CheckToken)).
		//	AllowIf(access.HasPermissionsWithExpr("!public_client")),
		//).
		With(token.NewEndpoint().
			Path(c.config.Endpoints.Token).
			AddGranter(c.config.tokenGranter()),
		)
}

// AuthorizeEndpointConfigurer implements security.Configurer
// responsible to configure //todo
type AuthorizeEndpointConfigurer struct {
	config *Configuration
	delegate IdpSecurityConfigurer
}

func (c *AuthorizeEndpointConfigurer) Configure(ws security.WebSecurity) {
	oauth2_path := c.config.Endpoints.Authorize.Location.Path
	oauth2_condition := c.config.Endpoints.Authorize.Condition
	ws.Route(matcher.RouteWithPattern(oauth2_path)).
		With(authorize.NewEndpoint().
			Path(oauth2_path).
			Condition(oauth2_condition).
			ApprovalPath(c.config.Endpoints.Approval).
			RequestProcessors(c.config.authorizeRequestProcessor()).
			ErrorHandler(c.config.errorHandler()).
			AuthorizeHanlder(c.config.authorizeHanlder()),
		).
		Route(matcher.RouteWithPattern(c.config.Endpoints.SamlSso.Location.Path)).
		With(saml_auth.NewEndpoint().
			Issuer(c.config.Issuer).
			SsoCondition(c.config.Endpoints.SamlSso.Condition).
			SsoLocation(c.config.Endpoints.SamlSso.Location).
			MetadataPath(c.config.Endpoints.SamlMetadata))

	c.delegate.Configure(ws, c.config)

	logout.Configure(ws).LogoutUrl(c.config.Endpoints.Logout)
}