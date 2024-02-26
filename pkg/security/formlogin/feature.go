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

package formlogin

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "time"
)

/*********************************
	Feature Impl
 *********************************/

//goland:noinspection GoNameStartsWithPackageName
type FormLoginFeature struct {
	successHandler         security.AuthenticationSuccessHandler
	failureHandler         security.AuthenticationErrorHandler
	loginUrl               string
	loginProcessUrl        string
	loginErrorUrl          string
	usernameParam          string
	passwordParam          string
	rememberCookieDomain   string
	rememberCookieSecured  bool
	rememberCookieValidity time.Duration
	rememberParam          string

	mfaEnabled    bool
	mfaUrl        string
	mfaVerifyUrl  string
	mfaRefreshUrl string
	mfaErrorUrl   string
	otpParam      string
}

func (f *FormLoginFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *FormLoginFeature) LoginUrl(loginUrl string) *FormLoginFeature {
	f.loginUrl = loginUrl
	return f
}

func (f *FormLoginFeature) LoginProcessUrl(loginProcessUrl string) *FormLoginFeature {
	f.loginProcessUrl = loginProcessUrl
	return f
}

func (f *FormLoginFeature) LoginErrorUrl(loginErrorUrl string) *FormLoginFeature {
	f.loginErrorUrl = loginErrorUrl
	return f
}

func (f *FormLoginFeature) UsernameParameter(usernameParam string) *FormLoginFeature {
	f.usernameParam = usernameParam
	return f
}

func (f *FormLoginFeature) PasswordParameter(passwordParam string) *FormLoginFeature {
	f.passwordParam = passwordParam
	return f
}

func (f *FormLoginFeature) RememberParameter(rememberParam string) *FormLoginFeature {
	f.rememberParam = rememberParam
	return f
}

func (f *FormLoginFeature) RememberCookieDomain(v string) *FormLoginFeature {
	f.rememberCookieDomain = v
	return f
}

func (f *FormLoginFeature) RememberCookieSecured(v bool) *FormLoginFeature {
	f.rememberCookieSecured = v
	return f
}

func (f *FormLoginFeature) RememberCookieValidity(v time.Duration) *FormLoginFeature {
	f.rememberCookieValidity = v
	return f
}

// SuccessHandler overrides LoginSuccessUrl
func (f *FormLoginFeature) SuccessHandler(successHandler security.AuthenticationSuccessHandler) *FormLoginFeature {
	f.successHandler = successHandler
	return f
}

// FailureHandler overrides LoginErrorUrl
func (f *FormLoginFeature) FailureHandler(failureHandler security.AuthenticationErrorHandler) *FormLoginFeature {
	f.failureHandler = failureHandler
	return f
}

func (f *FormLoginFeature) EnableMFA() *FormLoginFeature {
	f.mfaEnabled = true
	return f
}

func (f *FormLoginFeature) MfaUrl(mfaUrl string) *FormLoginFeature {
	f.mfaUrl = mfaUrl
	return f
}

func (f *FormLoginFeature) MfaVerifyUrl(mfaVerifyUrl string) *FormLoginFeature {
	f.mfaVerifyUrl = mfaVerifyUrl
	return f
}

func (f *FormLoginFeature) MfaRefreshUrl(mfaRefreshUrl string) *FormLoginFeature {
	f.mfaRefreshUrl = mfaRefreshUrl
	return f
}

func (f *FormLoginFeature) MfaErrorUrl(mfaErrorUrl string) *FormLoginFeature {
	f.mfaErrorUrl = mfaErrorUrl
	return f
}

func (f *FormLoginFeature) OtpParameter(otpParam string) *FormLoginFeature {
	f.otpParam = otpParam
	return f
}

/*********************************
	Constructors and Configure
 *********************************/

func Configure(ws security.WebSecurity) *FormLoginFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*FormLoginFeature)
	}
	panic(fmt.Errorf("unable to configure form login: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// New is Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *FormLoginFeature {
	return &FormLoginFeature{
		loginUrl:         "/login",
		loginProcessUrl:  "/login",
		loginErrorUrl:    "/login?error=true",
		usernameParam:    "username",
		passwordParam:    "password",
		rememberParam:    "remember-me",
		rememberCookieValidity: time.Hour,

		mfaUrl:        "/login/mfa",
		mfaVerifyUrl:  "/login/mfa",
		mfaRefreshUrl: "/login/mfa/refresh",
		mfaErrorUrl:   "/login/mfa?error=true",
		otpParam:      "otp",
	}
}
