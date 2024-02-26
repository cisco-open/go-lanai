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
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/security/csrf"
	"github.com/cisco-open/go-lanai/pkg/security/errorhandling"
	"github.com/cisco-open/go-lanai/pkg/security/logout"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/authorize"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/clientauth"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/openid"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/revoke"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/token"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/tokenauth"
	"github.com/cisco-open/go-lanai/pkg/security/redirect"
	"github.com/cisco-open/go-lanai/pkg/security/request_cache"
	"github.com/cisco-open/go-lanai/pkg/security/saml/idp"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
)

/***************************
	additional abstractions
 ***************************/

// IdpSecurityConfigurer interface for IDPs to implement for customizing "authorize" process
type IdpSecurityConfigurer interface {
	Configure(ws security.WebSecurity, config *Configuration)
}

// IdpLogoutSecurityConfigurer additional interface that IdpSecurityConfigurer could choose to implement for
// customizing "logout" process
// Note: IdpLogoutSecurityConfigurer is only invoked once per instance, the given security.WebSecurity are shared
//
//	between IDPs. Therefore, implementing class should not change Route or Condition on the given "ws"
type IdpLogoutSecurityConfigurer interface {
	ConfigureLogout(ws security.WebSecurity, config *Configuration)
}

/***************************
	security configurers
 ***************************/

// ClientAuthEndpointsConfigurer implements security.Configurer and order.Ordered
// responsible to configure misc using client auth
type ClientAuthEndpointsConfigurer struct {
	config *Configuration
}

func (c *ClientAuthEndpointsConfigurer) Order() int {
	return OrderClientAuthSecurityConfigurer
}

func (c *ClientAuthEndpointsConfigurer) Configure(ws security.WebSecurity) {
	// For Token endpoint
	ws.Route(matcher.RouteWithPattern(c.config.Endpoints.Token)).
		Route(matcher.RouteWithPattern(c.config.Endpoints.CheckToken)).
		Route(matcher.RouteWithPattern(fmt.Sprintf("%s/*", c.config.Endpoints.TenantHierarchy))).
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

// TokenAuthEndpointsConfigurer implements security.Configurer and order.Ordered
// responsible to configure misc using token auth
type TokenAuthEndpointsConfigurer struct {
	config *Configuration
}

func (c *TokenAuthEndpointsConfigurer) Order() int {
	return OrderTokenAuthSecurityConfigurer
}

func (c *TokenAuthEndpointsConfigurer) Configure(ws security.WebSecurity) {
	// For Token endpoint
	ws.Route(matcher.RouteWithPattern(c.config.Endpoints.UserInfo)).
		With(tokenauth.New().
			EnablePostBody(),
		).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New())
}

// AuthorizeEndpointConfigurer implements security.Configurer and order.Ordered
// responsible to configure "authorize" endpoint
type AuthorizeEndpointConfigurer struct {
	config   *Configuration
	delegate IdpSecurityConfigurer
}

func (c *AuthorizeEndpointConfigurer) Order() int {
	return OrderAuthorizeSecurityConfigurer
}

func (c *AuthorizeEndpointConfigurer) Configure(ws security.WebSecurity) {
	path := c.config.Endpoints.Authorize.Location.Path
	condition := c.config.Endpoints.Authorize.Condition
	ws.Route(matcher.RouteWithPattern(path)).
		With(authorize.NewEndpoint().
			Path(path).
			Condition(condition).
			ApprovalPath(c.config.Endpoints.Approval).
			RequestProcessor(c.config.authorizeRequestProcessor()).
			ErrorHandler(c.config.errorHandler()).
			AuthorizeHanlder(c.config.authorizeHandler()),
		).
		Route(matcher.RouteWithPattern(c.config.Endpoints.SamlSso.Location.Path)).
		With(samlidp.New().
			Issuer(c.config.Issuer).
			SsoCondition(c.config.Endpoints.SamlSso.Condition).
			SsoLocation(c.config.Endpoints.SamlSso.Location).
			MetadataPath(c.config.Endpoints.SamlMetadata).
			EnableSLO(c.config.Endpoints.Logout).
			SigningMethod(c.config.SamlIdpSigningMethod))

	c.delegate.Configure(ws, c.config)
}

// LogoutEndpointConfigurer implements security.Configurer and order.Ordered
// responsible to configure "logout" endpoint
type LogoutEndpointConfigurer struct {
	config    *Configuration
	delegates []IdpSecurityConfigurer
}

func (c *LogoutEndpointConfigurer) Order() int {
	return OrderLogoutSecurityConfigurer
}

func (c *LogoutEndpointConfigurer) Configure(ws security.WebSecurity) {
	// Logout Handler
	// Note: we disable default logout errHandler here because we don't want to unauthenticate user when PUT or DELETE is used
	logoutHandler := revoke.NewTokenRevokingLogoutHandler(func(opt *revoke.HanlderOption) {
		opt.Revoker = c.config.accessRevoker()
	})
	logoutSuccessHandler := revoke.NewTokenRevokeSuccessHandler(func(opt *revoke.SuccessOption) {
		opt.ClientStore = c.config.ClientStore
		opt.WhitelabelErrorPath = c.config.Endpoints.Error
		opt.RedirectWhitelist = utils.NewStringSet(c.config.properties.RedirectWhitelist...)
		opt.WhitelabelLoggedOutPath = c.config.Endpoints.LoggedOut
	})

	oidcLogoutHandler := openid.NewOidcLogoutHandler(func(opt *openid.HandlerOption) {
		opt.Dec = c.config.sharedJwtDecoder
		opt.Issuer = c.config.Issuer
		opt.ClientStore = c.config.ClientStore
	})
	oidcLogoutSuccessHandler := openid.NewOidcSuccessHandler(func(opt *openid.SuccessOption) {
		opt.ClientStore = c.config.ClientStore
		opt.WhitelabelErrorPath = c.config.Endpoints.Error
	})

	oidcEntryPoint := openid.NewOidcEntryPoint(func(opt *openid.EpOption) {
		opt.WhitelabelErrorPath = c.config.Endpoints.Error
	})

	errHandler := redirect.NewRedirectWithURL(c.config.Endpoints.Error)
	ws.With(session.New().SettingService(c.config.SessionSettingService)).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New().
			AccessDeniedHandler(errHandler),
		).
		With(csrf.New().
			IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(c.config.Endpoints.Logout)),
		).
		With(request_cache.New()).
		With(logout.New().
			LogoutUrl(c.config.Endpoints.Logout).
			// By using this instead of AddLogoutHandler, the default logout handler is disabled.
			LogoutHandlers(logoutHandler, oidcLogoutHandler).
			AddSuccessHandler(logoutSuccessHandler).
			AddSuccessHandler(oidcLogoutSuccessHandler).
			AddEntryPoint(oidcEntryPoint),
		).
		With(samlidp.NewLogout().
			Issuer(c.config.Issuer).
			SsoCondition(c.config.Endpoints.SamlSso.Condition).
			SsoLocation(c.config.Endpoints.SamlSso.Location).
			MetadataPath(c.config.Endpoints.SamlMetadata).
			EnableSLO(c.config.Endpoints.Logout),
		)

	for _, configurer := range c.delegates {
		if lc, ok := configurer.(IdpLogoutSecurityConfigurer); ok {
			lc.ConfigureLogout(ws, c.config)
		}
	}
}
