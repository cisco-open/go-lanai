package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
)

func registerEndpoints(registrar *web.Registrar, config *Configuration) {
	jwks := misc.NewJwkSetEndpoint(config.jwkStore())

	mappings := []interface{} {
		rest.New("jwks").Get(config.Endpoints.JwkSet).EndpointFunc(jwks.JwkSet).Build(),
	}
	registrar.Register(mappings...)
}
