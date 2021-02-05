package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/authorize"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/clientauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/token"
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
// responsible to configure endpoints using client auth
type ClientAuthEndpointsConfigurer struct {
	config *Configuration
}

func (c *ClientAuthEndpointsConfigurer) Configure(ws security.WebSecurity) {
	// TODO Complete this
	// For Token endpoint
	ws.Route(matcher.RouteWithPattern(c.config.Endpoints.Token)).
		With(clientauth.New().
			ClientStore(c.config.ClientStore).
			ClientSecretEncoder(c.config.clientSecretEncoder()).
			ErrorHandler(c.config.errorHandler()),
		).
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
	path := c.config.Endpoints.Authorize
	ws.Route(matcher.RouteWithPattern(path)).
		With(authorize.NewEndpoint().
			Path(path).
			RequestProcessors(c.config.authorizeRequestProcessor()).
			ErrorHandler(c.config.errorHandler()),
		)
	c.delegate.Configure(ws, c.config)
}
