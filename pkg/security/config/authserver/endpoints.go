// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package authserver

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/misc"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/openid"
    utils_matcher "github.com/cisco-open/go-lanai/pkg/utils/matcher"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
    "github.com/cisco-open/go-lanai/pkg/web/template"
)

func registerEndpoints(registrar *web.Registrar, config *Configuration) {
	jwks := misc.NewJwkSetEndpoint(config.jwkStore())
	ct := misc.NewCheckTokenEndpoint(config.Issuer, config.tokenStore())
	ui := misc.NewUserInfoEndpoint(config.Issuer, config.UserAccountStore, config.jwtEncoder())
	th := misc.NewTenantHierarchyEndpoint()

	mappings := []interface{}{
		template.New().Get(config.Endpoints.Error).HandlerFunc(errorhandling.ErrorWithStatus).Build(),

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

	// openid additional
	if config.OpenIDSSOEnabled {
		opConf := prepareWellKnownEndpoint(config)
		mappings = append(mappings,
			rest.New("openid-config").Get(openid.WellKnownEndpointOPConfig).
				EndpointFunc(opConf.OpenIDConfig).Build(),
		)
	}
	registrar.MustRegister(mappings...)
}

func acceptJwtMatcher() web.RequestMatcher {
	return matcher.RequestWithHeader("Accept", "application/jwt", true)
}

func notAcceptJwtMatcher() web.RequestMatcher {
	return utils_matcher.Not(matcher.RequestWithHeader("Accept", "application/jwt", true))
}

func prepareWellKnownEndpoint(config *Configuration) *misc.WellKnownEndpoint {
	extra := map[string]interface{}{
		openid.OPMetadataAuthEndpoint:       config.Endpoints.Authorize.Location.Path,
		openid.OPMetadataTokenEndpoint:      config.Endpoints.Token,
		openid.OPMetadataUserInfoEndpoint:   config.Endpoints.UserInfo,
		openid.OPMetadataJwkSetURI:          config.Endpoints.JwkSet,
		openid.OPMetadataEndSessionEndpoint: config.Endpoints.Logout,
	}
	return misc.NewWellKnownEndpoint(config.Issuer, config.IdpManager, extra)
}
