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
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/config/compatibility"
	"github.com/cisco-open/go-lanai/pkg/security/idp"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/grants"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/openid"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/revoke"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/common"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/tokenauth"
	"github.com/cisco-open/go-lanai/pkg/security/passwd"
	samlctx "github.com/cisco-open/go-lanai/pkg/security/saml"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"go.uber.org/fx"
	"net/url"
)

const (
	OrderAuthorizeSecurityConfigurer  = 0
	OrderLogoutSecurityConfigurer     = 50
	OrderClientAuthSecurityConfigurer = 100
	OrderTokenAuthSecurityConfigurer  = 200
)

type AuthorizationServerConfigurer func(*Configuration)

type configDI struct {
	fx.In
	AppContext         *bootstrap.ApplicationContext
	Properties         AuthServerProperties
	Configurer         AuthorizationServerConfigurer
	RedisClientFactory redis.ClientFactory
	ServerProperties   web.ServerProperties
	SessionProperties  security.SessionProperties
	CryptoProperties   jwt.CryptoProperties
	SessionStore       session.Store
	TimeoutSupport     oauth2.TimeoutApplier `optional:"true"`
	ApprovalStore      auth.ApprovalStore    `optional:"true"`
}

type authServerOut struct {
	fx.Out
	Config                  *Configuration
	CompatibilityCustomizer discovery.ServiceRegistrationCustomizer `group:"discovery"`
}

//goland:noinspection GoExportedFuncWithUnexportedType
func ProvideAuthServerDI(di configDI) authServerOut {
	config := Configuration{
		appContext:         di.AppContext,
		redisClientFactory: di.RedisClientFactory,
		sessionStore:       di.SessionStore,
		properties:         di.Properties,
		serverProperties:   di.ServerProperties,
		sessionProperties:  di.SessionProperties,
		cryptoProperties:   di.CryptoProperties,
		Issuer:             newIssuer(&di.Properties.Issuer, &di.ServerProperties),
		timeoutSupport:     di.TimeoutSupport,
		ApprovalStore:      di.ApprovalStore,
		Endpoints: Endpoints{
			Authorize: ConditionalEndpoint{
				Location:  &url.URL{Path: di.Properties.Endpoints.Authorize},
				Condition: matcher.NotRequest(matcher.RequestWithForm(oauth2.ParameterGrantType, samlctx.GrantTypeSamlSSO)),
			},
			Approval:   di.Properties.Endpoints.Approval,
			Token:      di.Properties.Endpoints.Token,
			CheckToken: di.Properties.Endpoints.CheckToken,
			UserInfo:   di.Properties.Endpoints.UserInfo,
			JwkSet:     di.Properties.Endpoints.JwkSet,
			Error:      di.Properties.Endpoints.Error,
			Logout:     di.Properties.Endpoints.Logout,
			LoggedOut:  di.Properties.Endpoints.LoggedOut,
			SamlSso: ConditionalEndpoint{
				Location:  &url.URL{Path: di.Properties.Endpoints.Authorize, RawQuery: fmt.Sprintf("%s=%s", oauth2.ParameterGrantType, samlctx.GrantTypeSamlSSO)},
				Condition: matcher.RequestWithForm(oauth2.ParameterGrantType, samlctx.GrantTypeSamlSSO),
			},
			SamlMetadata:    di.Properties.Endpoints.SamlMetadata,
			TenantHierarchy: di.Properties.Endpoints.TenantHierarchy,
		},
		OpenIDSSOEnabled: true,
	}
	di.Configurer(&config)
	return authServerOut{
		Config:                  &config,
		CompatibilityCustomizer: compatibility.CompatibilityDiscoveryCustomizer{},
	}
}

type initDI struct {
	fx.In
	Config            *Configuration
	WebRegistrar      *web.Registrar
	SecurityRegistrar security.Registrar
}

// ConfigureAuthorizationServer is the Configuration entry point
func ConfigureAuthorizationServer(di initDI) {
	// Securities
	di.SecurityRegistrar.Register(&ClientAuthEndpointsConfigurer{config: di.Config})
	di.SecurityRegistrar.Register(&TokenAuthEndpointsConfigurer{config: di.Config})
	for _, configuer := range di.Config.idpConfigurers {
		di.SecurityRegistrar.Register(&AuthorizeEndpointConfigurer{config: di.Config, delegate: configuer})
	}
	di.SecurityRegistrar.Register(&LogoutEndpointConfigurer{config: di.Config, delegates: di.Config.idpConfigurers})

	// Additional endpoints and other web configurations
	di.WebRegistrar.WarnDuplicateMiddlewares(true,
		di.Config.Endpoints.Authorize.Location.Path,
		di.Config.Endpoints.SamlSso.Location.Path,
		di.Config.Endpoints.Approval,
		di.Config.Endpoints.Logout,
	)
	registerEndpoints(di.WebRegistrar, di.Config)
}

/****************************
	configuration
 ****************************/

type ConditionalEndpoint struct {
	Location  *url.URL
	Condition web.RequestMatcher
}

type Endpoints struct {
	Authorize       ConditionalEndpoint
	Approval        string
	Token           string
	CheckToken      string
	UserInfo        string
	JwkSet          string
	Logout          string
	LoggedOut       string
	Error           string
	SamlSso         ConditionalEndpoint
	SamlMetadata    string
	TenantHierarchy string
}

type Configuration struct {
	// configurable items
	SessionSettingService session.SettingService
	ClientStore           oauth2.OAuth2ClientStore
	ClientSecretEncoder   passwd.PasswordEncoder
	Endpoints             Endpoints
	UserAccountStore      security.AccountStore
	TenantStore           security.TenantStore
	ProviderStore         security.ProviderStore
	UserPasswordEncoder   passwd.PasswordEncoder
	TokenStore            auth.TokenStore
	JwkStore              jwt.JwkStore
	IdpManager            idp.IdentityProviderManager
	Issuer                security.Issuer
	OpenIDSSOEnabled      bool
	SamlIdpSigningMethod  string
	ApprovalStore         auth.ApprovalStore
	CustomTokenGranter    []auth.TokenGranter
	CustomTokenEnhancer   []auth.TokenEnhancer
	CustomAuthRegistry    auth.AuthorizationRegistry

	// not directly configurable items
	appContext                *bootstrap.ApplicationContext
	redisClientFactory        redis.ClientFactory
	sessionStore              session.Store
	properties                AuthServerProperties
	serverProperties          web.ServerProperties
	sessionProperties         security.SessionProperties
	cryptoProperties          jwt.CryptoProperties
	idpConfigurers            []IdpSecurityConfigurer
	sharedContextDetailsStore security.ContextDetailsStore
	sharedAuthRegistry        auth.AuthorizationRegistry
	sharedAccessRevoker       auth.AccessRevoker
	sharedErrorHandler        *auth.OAuth2ErrorHandler
	sharedTokenGranter        auth.TokenGranter
	sharedAuthService         auth.AuthorizationService
	sharedPasswdAuthenticator security.Authenticator
	sharedJwtEncoder          jwt.JwtEncoder
	sharedJwtDecoder          jwt.JwtDecoder
	sharedDetailsFactory      *common.ContextDetailsFactory
	sharedARProcessor         auth.AuthorizeRequestProcessor
	sharedAuthHandler         auth.AuthorizeHandler
	sharedAuthCodeStore       auth.AuthorizationCodeStore
	sharedTokenAuthenticator  security.Authenticator
	timeoutSupport            oauth2.TimeoutApplier
}

func (c *Configuration) AddIdp(configurer IdpSecurityConfigurer) {
	c.idpConfigurers = append(c.idpConfigurers, configurer)
}

func newIssuer(props *IssuerProperties, serverProps *web.ServerProperties) security.Issuer {
	contextPath := props.ContextPath
	if contextPath == "" {
		contextPath = serverProps.ContextPath
	}
	return security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
		*opt = security.DefaultIssuerDetails{
			Protocol:    props.Protocol,
			Domain:      props.Domain,
			Port:        props.Port,
			ContextPath: contextPath,
			IncludePort: props.IncludePort,
		}
	})
}

func (c *Configuration) clientSecretEncoder() passwd.PasswordEncoder {
	if c.ClientSecretEncoder == nil {
		c.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.ClientSecretEncoder
}

func (c *Configuration) userPasswordEncoder() passwd.PasswordEncoder {
	if c.UserPasswordEncoder == nil {
		c.UserPasswordEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.UserPasswordEncoder
}

func (c *Configuration) errorHandler() *auth.OAuth2ErrorHandler {
	if c.sharedErrorHandler == nil {
		c.sharedErrorHandler = auth.NewOAuth2ErrorHandler()
	}
	return c.sharedErrorHandler
}

func (c *Configuration) tokenGranter() auth.TokenGranter {
	if c.sharedTokenGranter == nil {
		granters := []auth.TokenGranter{
			grants.NewAuthorizationCodeGranter(c.authorizationService(), c.authorizeCodeStore()),
			grants.NewClientCredentialsGranter(c.authorizationService()),
			grants.NewRefreshGranter(c.authorizationService(), c.tokenStore()),
			grants.NewSwitchUserGranter(c.authorizationService(), c.tokenAuthenticator(), c.UserAccountStore),
			grants.NewSwitchTenantGranter(c.authorizationService(), c.tokenAuthenticator(), c.UserAccountStore),
		}

		// password granter is optional
		if c.passwordGrantAuthenticator() != nil {
			passwordGranter := grants.NewPasswordGranter(c.authorizationService(), c.passwordGrantAuthenticator())
			granters = append(granters, passwordGranter)
		}

		for _, custom := range c.CustomTokenGranter {
			switch v := custom.(type) {
			case auth.AuthorizationServiceInjector:
				v.Inject(c.authorizationService())
			default:
				// do nothing
			}
		}
		granters = append(granters, c.CustomTokenGranter...)

		c.sharedTokenGranter = auth.NewCompositeTokenGranter(granters...)
	}
	return c.sharedTokenGranter
}

func (c *Configuration) passwordGrantAuthenticator() security.Authenticator {
	if c.sharedPasswdAuthenticator == nil && c.UserAccountStore != nil {
		authenticator, err := passwd.NewAuthenticatorBuilder(
			passwd.New().
				AccountStore(c.UserAccountStore).
				PasswordEncoder(c.userPasswordEncoder()).
				MFA(false),
		).Build(context.Background())

		if err == nil {
			c.sharedPasswdAuthenticator = authenticator
		}
	}
	return c.sharedPasswdAuthenticator
}

func (c *Configuration) contextDetailsStore() security.ContextDetailsStore {
	if c.sharedContextDetailsStore == nil {
		c.sharedContextDetailsStore = common.NewRedisContextDetailsStore(c.appContext, c.redisClientFactory, c.timeoutSupport)
	}
	return c.sharedContextDetailsStore
}

func (c *Configuration) authorizationRegistry() auth.AuthorizationRegistry {
	if c.sharedAuthRegistry == nil {
		if c.CustomAuthRegistry != nil {
			c.sharedAuthRegistry = c.CustomAuthRegistry
		} else {
			c.sharedAuthRegistry = c.contextDetailsStore().(auth.AuthorizationRegistry)
		}
	}
	return c.sharedAuthRegistry
}

func (c *Configuration) tokenStore() auth.TokenStore {
	if c.TokenStore == nil {
		c.TokenStore = auth.NewJwtTokenStore(func(opt *auth.JTSOption) {
			opt.DetailsStore = c.contextDetailsStore()
			opt.Encoder = c.jwtEncoder()
			opt.Decoder = c.jwtDecoder()
			opt.AuthRegistry = c.authorizationRegistry()
		})
	}
	return c.TokenStore
}

func (c *Configuration) authorizationService() auth.AuthorizationService {
	if c.sharedAuthService == nil {
		c.sharedAuthService = auth.NewDefaultAuthorizationService(func(conf *auth.DASOption) {
			conf.TokenStore = c.tokenStore()
			conf.DetailsFactory = c.contextDetailsFactory()
			conf.Issuer = c.Issuer
			conf.ClientStore = c.ClientStore
			conf.AccountStore = c.UserAccountStore
			conf.TenantStore = c.TenantStore
			conf.ProviderStore = c.ProviderStore
			if c.OpenIDSSOEnabled {
				openidEnhancer := openid.NewOpenIDTokenEnhancer(func(opt *openid.EnhancerOption) {
					opt.Issuer = c.Issuer
					opt.JwtEncoder = c.jwtEncoder()
				})
				conf.PostTokenEnhancers = append(conf.PostTokenEnhancers, openidEnhancer)
			}
			conf.TokenEnhancers = append(conf.TokenEnhancers, c.CustomTokenEnhancer...)
		})
	}

	return c.sharedAuthService
}

func (c *Configuration) jwkStore() jwt.JwkStore {
	if c.JwkStore == nil {
		c.JwkStore = jwt.NewFileJwkStore(c.cryptoProperties)
	}
	return c.JwkStore
}

func (c *Configuration) jwtEncoder() jwt.JwtEncoder {
	if c.sharedJwtEncoder == nil {
		c.sharedJwtEncoder = jwt.NewSignedJwtEncoder(jwt.SignWithJwkStore(c.jwkStore(), c.cryptoProperties.Jwt.KeyName))
	}
	return c.sharedJwtEncoder
}

func (c *Configuration) jwtDecoder() jwt.JwtDecoder {
	if c.sharedJwtDecoder == nil {
		c.sharedJwtDecoder = jwt.NewSignedJwtDecoder(jwt.VerifyWithJwkStore(c.jwkStore(), c.cryptoProperties.Jwt.KeyName))
	}
	return c.sharedJwtDecoder
}

func (c *Configuration) contextDetailsFactory() *common.ContextDetailsFactory {
	if c.sharedDetailsFactory == nil {
		c.sharedDetailsFactory = common.NewContextDetailsFactory()
	}
	return c.sharedDetailsFactory
}

func (c *Configuration) authorizeRequestProcessor() auth.AuthorizeRequestProcessor {
	if c.sharedARProcessor == nil {
		processors := []auth.ChainedAuthorizeRequestProcessor{
			auth.NewStandardAuthorizeRequestProcessor(func(opt *auth.StdARPOption) {
				opt.ClientStore = c.ClientStore
				opt.AccountStore = c.UserAccountStore
			}),
		}
		if c.OpenIDSSOEnabled {
			p := openid.NewOpenIDAuthorizeRequestProcessor(func(opt *openid.ARPOption) {
				opt.Issuer = c.Issuer
				opt.JwtDecoder = c.jwtDecoder()
			})
			processors = append([]auth.ChainedAuthorizeRequestProcessor{p}, processors...)
		}
		c.sharedARProcessor = auth.NewAuthorizeRequestProcessor(processors...)
	}
	return c.sharedARProcessor
}

func (c *Configuration) authorizeHandler() auth.AuthorizeHandler {
	if c.sharedAuthHandler == nil {
		//TODO OIDC Implicit flow extension
		c.sharedAuthHandler = auth.NewAuthorizeHandler(func(opt *auth.AuthHandlerOption) {
			//opt.Extensions = OIDC extensions
			opt.ApprovalPageTmpl = "authorize.tmpl"
			opt.ApprovalUrl = c.Endpoints.Approval
			opt.AuthService = c.authorizationService()
			opt.AuthCodeStore = c.authorizeCodeStore()
		})
	}
	return c.sharedAuthHandler
}

func (c *Configuration) authorizeCodeStore() auth.AuthorizationCodeStore {
	if c.sharedAuthCodeStore == nil {
		c.sharedAuthCodeStore = auth.NewRedisAuthorizationCodeStore(c.appContext, c.redisClientFactory, c.sessionProperties.DbIndex)
	}
	return c.sharedAuthCodeStore
}

func (c *Configuration) tokenAuthenticator() security.Authenticator {
	if c.sharedTokenAuthenticator == nil {
		c.sharedTokenAuthenticator = tokenauth.NewAuthenticator(func(opt *tokenauth.AuthenticatorOption) {
			opt.TokenStoreReader = c.tokenStore()
		})
	}
	return c.sharedTokenAuthenticator
}

func (c *Configuration) accessRevoker() auth.AccessRevoker {
	if c.sharedAccessRevoker == nil {
		c.sharedAccessRevoker = revoke.NewDefaultAccessRevoker(func(opt *revoke.RevokerOption) {
			opt.AuthRegistry = c.authorizationRegistry()
			opt.SessionStore = c.sessionStore
			opt.TokenStoreReader = c.tokenStore()
		})
	}
	return c.sharedAccessRevoker
}

func (c *Configuration) approvalStore() auth.ApprovalStore {
	return c.ApprovalStore
}
