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

package clientauth

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/access"
    "github.com/cisco-open/go-lanai/pkg/security/basicauth"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
    "github.com/cisco-open/go-lanai/pkg/security/passwd"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
)

var (
	FeatureId = security.FeatureId("OAuth2ClientAuth", security.FeatureOrderOAuth2ClientAuth)
)

//goland:noinspection GoNameStartsWithPackageName
type ClientAuthConfigurer struct {
}

func newClientAuthConfigurer() *ClientAuthConfigurer {
	return &ClientAuthConfigurer{
	}
}

func (c *ClientAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*ClientAuthFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	// configure other features
	passwd.Configure(ws).
		AccountStore(c.clientAccountStore(f)).
		PasswordEncoder(f.clientSecretEncoder).
		MFA(false)
	// no entry point, everything handled by access denied handler
	basicauth.Configure(ws).
		EntryPoint(nil)
	access.Configure(ws).
		Request(matcher.AnyRequest()).
		Authenticated()
	errorhandling.Configure(ws).
		AdditionalErrorHandler(f.errorHandler)

	// add middleware to translate authentication error to oauth2 error
	mw := NewClientAuthMiddleware(func(opt *MWOption) {
		opt.Authenticator = ws.Authenticator()
		opt.SuccessHandler = ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler)
	})
	ws.Add(middleware.NewBuilder("client auth error translator").
		Order(security.MWOrderPreAuth).
		Use(mw.ErrorTranslationHandlerFunc()),
	)

	// add middleware to support form based client auth
	if f.allowForm {
		ws.Add(middleware.NewBuilder("form client auth").
			Order(security.MWOrderFormAuth).
			Use(mw.ClientAuthFormHandlerFunc()),
		)
	}

	return nil
}

func (c *ClientAuthConfigurer) validate(f *ClientAuthFeature, ws security.WebSecurity) error {
	if f.clientStore == nil {
		return fmt.Errorf("client store for client authentication is not set")
	}

	if f.clientSecretEncoder == nil {
		f.clientSecretEncoder = passwd.NewNoopPasswordEncoder()
	}

	if f.errorHandler == nil {
		f.errorHandler = auth.NewOAuth2ErrorHandler()
	}
	return nil
}

func (c *ClientAuthConfigurer) clientAccountStore(f *ClientAuthFeature) *auth.OAuth2ClientAccountStore {
	return auth.WrapOAuth2ClientStore(f.clientStore)
}



