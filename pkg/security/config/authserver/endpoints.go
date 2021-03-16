package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	utils_matcher "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
)

func registerEndpoints(registrar *web.Registrar, config *Configuration) {
	jwks := misc.NewJwkSetEndpoint(config.jwkStore())
	ct := misc.NewCheckTokenEndpoint(config.Issuer, config.tokenStore())
	ui := misc.NewUserInfoEndpoint(config.Issuer, config.UserAccountStore, config.jwtEncoder())

	mappings := []interface{} {
		rest.New("jwks").Get(config.Endpoints.JwkSet).EndpointFunc(jwks.JwkSet).Build(),
		rest.New("check_token").Post(config.Endpoints.CheckToken).EndpointFunc(ct.CheckToken).Build(),
		rest.New("userinfo GET").Get(config.Endpoints.UserInfo).
			Condition(acceptJwtMatcher()).
			EncodeResponseFunc(misc.JwtResponseEncoder()).
			EndpointFunc(ui.JwtUserInfo).Build(),
		rest.New("userinfo GET").Get(config.Endpoints.UserInfo).
			Condition(notAcceptJwtMatcher()).EndpointFunc(ui.PlainUserInfo).Build(),
		rest.New("userinfo POST").Post(config.Endpoints.UserInfo).
			Condition(acceptJwtMatcher()).
			EncodeResponseFunc(misc.JwtResponseEncoder()).
			EndpointFunc(ui.JwtUserInfo).Build(),
		rest.New("userinfo POST").Post(config.Endpoints.UserInfo).
			Condition(notAcceptJwtMatcher()).
			EndpointFunc(ui.PlainUserInfo).Build(),
	}
	registrar.Register(mappings...)
}

func acceptJwtMatcher() web.RequestMatcher {
	return matcher.RequestWithHeader("Accept", "application/jwt", true)
}

func notAcceptJwtMatcher() web.RequestMatcher {
	return utils_matcher.Not(matcher.RequestWithHeader("Accept", "application/jwt", true))
}

