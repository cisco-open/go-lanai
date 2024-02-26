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

package resserver

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/common"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/tokenauth"
	"go.uber.org/fx"
)

type ResourceServerConfigurer func(*Configuration)

type resServerConfigDI struct {
	fx.In
	AppContext           *bootstrap.ApplicationContext
	RedisClientFactory   redis.ClientFactory
	CryptoProperties     jwt.CryptoProperties
	TimeoutSupport 	 	 oauth2.TimeoutApplier `optional:"true"`
	Configurer           ResourceServerConfigurer
}

type resServerOut struct {
	fx.Out
	Config *Configuration
	TokenStore oauth2.TokenStoreReader
}

//goland:noinspection GoExportedFuncWithUnexportedType,HttpUrlsUsage
func ProvideResServerDI(di resServerConfigDI) resServerOut {
	config := Configuration{
		appContext:         di.AppContext,
		cryptoProperties:   di.CryptoProperties,
		redisClientFactory: di.RedisClientFactory,
		timeoutSupport:     di.TimeoutSupport,
		RemoteEndpoints: RemoteEndpoints{
			Token:      "http://authserver/v2/token",
			CheckToken: "http://authserver/v2/check_token",
			UserInfo:   "http://authserver/v2/userinfo",
			JwkSet:     "http://authserver/v2/jwks",
		},
	}
	di.Configurer(&config)
	return resServerOut{
		Config: &config,
		TokenStore: config.SharedTokenStoreReader(),
	}
}

type resServerDI struct {
	fx.In
	Config               *Configuration
	SecurityRegistrar    security.Registrar
	DiscoveryCustomizers *discovery.Customizers `optional:"true"`
}

// ConfigureResourceServer configuration entry point
func ConfigureResourceServer(di resServerDI) {
	// SMCR only applicable when discovery is on
	if di.DiscoveryCustomizers != nil {
		di.DiscoveryCustomizers.Add(security.CompatibilityDiscoveryCustomizer)
	}

	// reigester token auth feature
	configurer := tokenauth.NewTokenAuthConfigurer(func(opt *tokenauth.TokenAuthOption) {
		opt.TokenStoreReader = di.Config.tokenStoreReader()
	})
	di.SecurityRegistrar.(security.FeatureRegistrar).RegisterFeature(tokenauth.FeatureId, configurer)
}

/****************************
	configuration
 ****************************/

type RemoteEndpoints struct {
	Token      string
	CheckToken string
	UserInfo   string
	JwkSet     string
}

type Configuration struct {
	// configurable items
	RemoteEndpoints  RemoteEndpoints
	TokenStoreReader oauth2.TokenStoreReader
	JwkStore         jwt.JwkStore

	// not directly configurable items
	appContext                *bootstrap.ApplicationContext
	redisClientFactory        redis.ClientFactory
	cryptoProperties          jwt.CryptoProperties
	sharedTokenAuthenticator  security.Authenticator
	sharedErrorHandler        *tokenauth.OAuth2ErrorHandler
	sharedContextDetailsStore security.ContextDetailsStore
	sharedJwtDecoder          jwt.JwtDecoder
	timeoutSupport 			  oauth2.TimeoutApplier
}

func (c *Configuration) SharedTokenStoreReader() oauth2.TokenStoreReader {
	return c.tokenStoreReader()
}

func (c *Configuration) errorHandler() *tokenauth.OAuth2ErrorHandler {
	if c.sharedErrorHandler == nil {
		c.sharedErrorHandler = tokenauth.NewOAuth2ErrorHanlder()
	}
	return c.sharedErrorHandler
}

func (c *Configuration) contextDetailsStore() security.ContextDetailsStore {
	if c.sharedContextDetailsStore == nil {
		c.sharedContextDetailsStore = common.NewRedisContextDetailsStore(c.appContext, c.redisClientFactory, c.timeoutSupport)
	}
	return c.sharedContextDetailsStore
}

func (c *Configuration) tokenStoreReader() oauth2.TokenStoreReader {
	if c.TokenStoreReader == nil {
		c.TokenStoreReader = common.NewJwtTokenStoreReader(func(opt *common.JTSROption) {
			opt.DetailsStore = c.contextDetailsStore()
			opt.Decoder = c.jwtDecoder()
		})
	}
	return c.TokenStoreReader
}

func (c *Configuration) jwkStore() jwt.JwkStore {
	if c.JwkStore == nil {
		c.JwkStore = jwt.NewFileJwkStore(c.cryptoProperties)
	}
	return c.JwkStore
}

func (c *Configuration) jwtDecoder() jwt.JwtDecoder {
	if c.sharedJwtDecoder == nil {
		c.sharedJwtDecoder = jwt.NewRS256JwtDecoder(c.jwkStore(), c.cryptoProperties.Jwt.KeyName)
	}
	return c.sharedJwtDecoder
}

func (c *Configuration) tokenAuthenticator() security.Authenticator {
	if c.sharedTokenAuthenticator == nil {
		c.sharedTokenAuthenticator = tokenauth.NewAuthenticator(func(opt *tokenauth.AuthenticatorOption) {
			opt.TokenStoreReader = c.tokenStoreReader()
		})
	}
	return c.sharedTokenAuthenticator
}