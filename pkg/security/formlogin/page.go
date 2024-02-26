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
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/security/redirect"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/template"
	"strings"
)

const (
	LoginModelKeyUsernameParam      = "usernameParam"
	LoginModelKeyPasswordParam      = "passwordParam"
	LoginModelKeyLoginProcessUrl    = "loginProcessUrl"
	LoginModelKeyRememberedUsername = "rememberedUsername"
	LoginModelKeyOtpParam           = "otpParam"
	LoginModelKeyMfaVerifyUrl       = "mfaVerifyUrl"
	LoginModelKeyMfaRefreshUrl      = "mfaRefreshUrl"
	LoginModelKeyMsxVersion         = "MSXVersion"
)

type DefaultFormLoginController struct {
	buildInfoResolver bootstrap.BuildInfoResolver
	loginTemplate     string
	loginProcessUrl   string
	usernameParam     string
	passwordParam     string

	mfaTemplate   string
	mfaVerifyUrl  string
	mfaRefreshUrl string
	otpParam      string
}

type PageOptionsFunc func(*DefaultFormLoginPageOptions)

type DefaultFormLoginPageOptions struct {
	BuildInfoResolver bootstrap.BuildInfoResolver
	LoginTemplate     string
	UsernameParam     string
	PasswordParam     string
	LoginProcessUrl   string

	MfaTemplate   string
	OtpParam      string
	MfaVerifyUrl  string
	MfaRefreshUrl string
}

func NewDefaultLoginFormController(options ...PageOptionsFunc) *DefaultFormLoginController {
	opts := DefaultFormLoginPageOptions{}
	for _, f := range options {
		f(&opts)
	}

	return &DefaultFormLoginController{
		buildInfoResolver: opts.BuildInfoResolver,
		loginTemplate:     opts.LoginTemplate,
		loginProcessUrl:   opts.LoginProcessUrl,
		usernameParam:     opts.UsernameParam,
		passwordParam:     opts.PasswordParam,

		mfaTemplate:   opts.MfaTemplate,
		mfaVerifyUrl:  opts.MfaVerifyUrl,
		mfaRefreshUrl: opts.MfaRefreshUrl,
		otpParam:      opts.OtpParam,
	}
}

type LoginRequest struct {
	Error bool `form:"error"`
}

type OTPVerificationRequest struct {
	Error bool `form:"error"`
}

func (c *DefaultFormLoginController) Mappings() []web.Mapping {
	return []web.Mapping{
		template.New().Get("/login").HandlerFunc(c.LoginForm).Build(),
		template.New().Get("/login/mfa").HandlerFunc(c.OtpVerificationForm).Build(),
	}
}

func (c *DefaultFormLoginController) LoginForm(ctx context.Context, r *LoginRequest) (*template.ModelView, error) {
	model := template.Model{
		LoginModelKeyUsernameParam:   c.usernameParam,
		LoginModelKeyPasswordParam:   c.passwordParam,
		LoginModelKeyLoginProcessUrl: c.loginProcessUrl,
		LoginModelKeyMsxVersion:      c.msxVersion(),
	}

	s := session.Get(ctx)
	if s != nil {
		if err, errOk := s.Flash(redirect.FlashKeyPreviousError).(error); errOk && r.Error {
			model[template.ModelKeyError] = err
		}

		if username, usernameOk := s.Flash(c.usernameParam).(string); usernameOk {
			model[c.usernameParam] = username
		}
	}

	if gc := web.GinContext(ctx); gc != nil {
		if remembered, e := gc.Cookie(CookieKeyRememberedUsername); e == nil && remembered != "" {
			model[LoginModelKeyRememberedUsername] = remembered
		}
	}

	return &template.ModelView{
		View:  c.loginTemplate,
		Model: model,
	}, nil
}

func (c *DefaultFormLoginController) OtpVerificationForm(ctx context.Context, r *OTPVerificationRequest) (*template.ModelView, error) {
	model := template.Model{
		LoginModelKeyOtpParam:      c.otpParam,
		LoginModelKeyMfaVerifyUrl:  c.mfaVerifyUrl,
		LoginModelKeyMfaRefreshUrl: c.mfaRefreshUrl,
		LoginModelKeyMsxVersion:    c.msxVersion(),
	}

	s := session.Get(ctx)
	if s != nil {
		if err, errOk := s.Flash(redirect.FlashKeyPreviousError).(error); errOk && r.Error {
			model[template.ModelKeyError] = err
		}
	}

	return &template.ModelView{
		View:  c.mfaTemplate,
		Model: model,
	}, nil
}

func (c *DefaultFormLoginController) msxVersion() string {
	if c.buildInfoResolver != nil {
		return c.buildInfoResolver.Resolve().Version
	}

	if strings.ToLower(bootstrap.BuildVersion) == "unknown" {
		return ""
	}
	return bootstrap.BuildVersion
}
