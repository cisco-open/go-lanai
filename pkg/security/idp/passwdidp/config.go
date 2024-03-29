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

package passwdidp

import (
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/security/config/authserver"
	"github.com/cisco-open/go-lanai/pkg/security/csrf"
	"github.com/cisco-open/go-lanai/pkg/security/errorhandling"
	"github.com/cisco-open/go-lanai/pkg/security/formlogin"
	"github.com/cisco-open/go-lanai/pkg/security/idp"
	"github.com/cisco-open/go-lanai/pkg/security/passwd"
	"github.com/cisco-open/go-lanai/pkg/security/redirect"
	"github.com/cisco-open/go-lanai/pkg/security/request_cache"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"time"
)

type Options func(opt *option)
type option struct {
	Properties   *PwdAuthProperties
	MFAListeners []passwd.MFAEventListenerFunc
}

func WithProperties(props *PwdAuthProperties) Options {
	return func(opt *option) {
		opt.Properties = props
	}
}

func WithMFAListeners(listeners ...passwd.MFAEventListenerFunc) Options {
	return func(opt *option) {
		opt.MFAListeners = append(opt.MFAListeners, listeners...)
	}
}

// PasswordIdpSecurityConfigurer implements authserver.IdpSecurityConfigurer
type PasswordIdpSecurityConfigurer struct {
	props        *PwdAuthProperties
	mfaListeners []passwd.MFAEventListenerFunc
}

func NewPasswordIdpSecurityConfigurer(opts ...Options) *PasswordIdpSecurityConfigurer {
	opt := option{
		Properties:   NewPwdAuthProperties(),
		MFAListeners: []passwd.MFAEventListenerFunc{},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	return &PasswordIdpSecurityConfigurer{
		props:        opt.Properties,
		mfaListeners: opt.MFAListeners,
	}
}

func (c *PasswordIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authserver.Configuration) {
	// For Authorize endpoint
	condition := idp.RequestWithAuthenticationFlow(idp.InternalIdpForm, config.IdpManager)
	ws = ws.AndCondition(condition)

	if !c.props.Enabled {
		return
	}

	// Note: reset password url is not supported by whitelabel login form, and is hardcoded in MSX UI
	handler := redirect.NewRedirectWithURL(config.Endpoints.Error)
	ws.
		With(session.New().SettingService(config.SessionSettingService)).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(passwd.New().
			MFA(c.props.MFA.Enabled).
			OtpTTL(time.Duration(c.props.MFA.OtpTTL)).
			PasswordEncoder(config.UserPasswordEncoder).
			OtpVerifyLimit(c.props.MFA.OtpMaxAttempts).
			OtpRefreshLimit(c.props.MFA.OtpResendLimit).
			OtpLength(c.props.MFA.OtpLength).
			OtpSecretSize(c.props.MFA.OtpSecretSize).
			MFAEventListeners(c.mfaListeners...),
		).
		With(formlogin.New().
			EnableMFA().
			LoginUrl(c.props.Endpoints.FormLogin).
			LoginProcessUrl(c.props.Endpoints.FormLoginProcess).
			LoginErrorUrl(c.props.Endpoints.FormLoginError).
			MfaUrl(c.props.Endpoints.OtpVerify).
			MfaVerifyUrl(c.props.Endpoints.OtpVerifyProcess).
			MfaRefreshUrl(c.props.Endpoints.OtpVerifyResend).
			MfaErrorUrl(c.props.Endpoints.OtpVerifyError).
			RememberCookieSecured(c.props.RememberMe.UseSecureCookie).
			RememberCookieDomain(c.props.RememberMe.CookieDomain).
			RememberCookieValidity(time.Duration(c.props.RememberMe.CookieValidity)),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		).
		With(csrf.New().
			IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(config.Endpoints.Authorize.Location.Path)),
		).
		With(request_cache.New())
}
