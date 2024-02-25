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

package tokenauth

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
)

var (
	FeatureId = security.FeatureId("OAuth2TokenAuth", security.FeatureOrderOAuth2TokenAuth)
)

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthConfigurer struct {
	tokenStoreReader oauth2.TokenStoreReader
}

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthOptions func(opt *TokenAuthOption)

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthOption struct {
	TokenStoreReader oauth2.TokenStoreReader
}

func NewTokenAuthConfigurer(opts ...TokenAuthOptions) *TokenAuthConfigurer {
	opt := TokenAuthOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &TokenAuthConfigurer{
		tokenStoreReader: opt.TokenStoreReader,
	}
}

func (c *TokenAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*TokenAuthFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	// configure other features
	errorhandling.Configure(ws).
		AdditionalErrorHandler(f.errorHandler)
	// use ScopesApproved(...) for scope based access decision maker

	// setup authenticator
	authenticator := NewAuthenticator(func(opt *AuthenticatorOption) {
		opt.TokenStoreReader = c.tokenStoreReader
	})
	ws.Authenticator().(*security.CompositeAuthenticator).Add(authenticator)

	// prepare middlewares
	successHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler)
	if !ok {
		successHandler = security.NewAuthenticationSuccessHandler()
	}
	mw := NewTokenAuthMiddleware(func(opt *TokenAuthMWOption) {
		opt.Authenticator = ws.Authenticator()
		opt.SuccessHandler = successHandler
		opt.PostBodyEnabled = f.postBodyEnabled
	})

	// install middlewares
	tokenAuth := middleware.NewBuilder("token authentication").
		Order(security.MWOrderOAuth2TokenAuth).
		Use(mw.AuthenticateHandlerFunc())

	ws.Add(tokenAuth)
	return nil
}

func (c *TokenAuthConfigurer) validate(f *TokenAuthFeature, _ security.WebSecurity) error {
	if c.tokenStoreReader == nil {
		return fmt.Errorf("token store reader is not pre-configured")
	}

	if f.errorHandler == nil {
		f.errorHandler = NewOAuth2ErrorHanlder()
	}

	//if f.granters == nil || len(f.granters) == 0 {
	//	return fmt.Errorf("token granters is not set")
	//}
	return nil
}



