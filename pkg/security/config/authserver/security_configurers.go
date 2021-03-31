package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/authorize"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/clientauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/revoke"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/token"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
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

// TokenAuthEndpointsConfigurer implements security.Configurer
// responsible to configure misc using token auth
type TokenAuthEndpointsConfigurer struct {
	config *Configuration
}

func (c *TokenAuthEndpointsConfigurer) Configure(ws security.WebSecurity) {
	// For Token endpoint
	ws.Route(matcher.RouteWithPattern(c.config.Endpoints.UserInfo)).
		With(tokenauth.New()).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New())
}

// AuthorizeEndpointConfigurer implements security.Configurer
// responsible to configure "authorize" endpoint
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

	// Logout Handler
	// Note: we disable default logout handler here because we don't want to unauthenticate user when PUT or DELETE is used
	logoutHandler := revoke.NewTokenRevokingLogoutHanlder(func(opt *revoke.HanlderOption) {
		opt.Revoker = c.config.accessRevoker()
	})
	logoutSuccessHandler := revoke.NewTokenRevokeSuccessHandler(func(opt *revoke.SuccessOption) {
		opt.ClientStore = c.config.ClientStore
		opt.WhitelabelErrorPath = "/error"
		opt.RedirectWhitelist = c.config.properties.RedirectWhitelist
	})
	logout.Configure(ws).
		LogoutUrl(c.config.Endpoints.Logout).
		LogoutHandlers(logoutHandler).
		SuccessHandler(logoutSuccessHandler)
}