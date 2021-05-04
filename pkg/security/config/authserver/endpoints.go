package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	utils_matcher "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"fmt"
)

func registerEndpoints(registrar *web.Registrar, config *Configuration) {
	jwks := misc.NewJwkSetEndpoint(config.jwkStore())
	ct := misc.NewCheckTokenEndpoint(config.Issuer, config.tokenStore())
	ui := misc.NewUserInfoEndpoint(config.Issuer, config.UserAccountStore, config.jwtEncoder())
	th := misc.NewEndpoint()

	mappings := []interface{} {
		template.New().Get("/error").HandlerFunc(errorhandling.ErrorWithStatus).Build(),

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

		rest.New("tenant hierarchy parent").Get(fmt.Sprintf("%s/%s", config.Endpoints.TenantHierarchy, "parent")).
			EndpointFunc(th.GetParent).EncodeResponseFunc(misc.StringResponseEncoder()).Build(),
		rest.New("tenant hierarchy children").Get(fmt.Sprintf("%s/%s", config.Endpoints.TenantHierarchy, "children")).
			EndpointFunc(th.GetChildren).Build(),
		rest.New("tenant hierarchy ancestors").Get(fmt.Sprintf("%s/%s", config.Endpoints.TenantHierarchy, "ancestors")).
			EndpointFunc(th.GetAncestors).Build(),
		rest.New("tenant hierarchy descendants").Get(fmt.Sprintf("%s/%s", config.Endpoints.TenantHierarchy, "descendants")).
			EndpointFunc(th.GetDescendants).Build(),
		rest.New("tenant hierarchy root").Get(fmt.Sprintf("%s/%s", config.Endpoints.TenantHierarchy, "root")).
			EndpointFunc(th.GetRoot).EncodeResponseFunc(misc.StringResponseEncoder()).Build(),
	}
	registrar.Register(mappings...)
}

func acceptJwtMatcher() web.RequestMatcher {
	return matcher.RequestWithHeader("Accept", "application/jwt", true)
}

func notAcceptJwtMatcher() web.RequestMatcher {
	return utils_matcher.Not(matcher.RequestWithHeader("Accept", "application/jwt", true))
}

