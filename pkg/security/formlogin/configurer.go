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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

var (
	FeatureId = security.FeatureId("FormLogin", security.FeatureOrderFormLogin)
)

//goland:noinspection GoNameStartsWithPackageName
type FormLoginConfigurer struct {
	serverProps web.ServerProperties
	configured  bool
}

func newFormLoginConfigurer(serverProps web.ServerProperties) *FormLoginConfigurer {
	return &FormLoginConfigurer{
		serverProps: serverProps,
	}
}

func (c *FormLoginConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := c.validate(feature.(*FormLoginFeature), ws); err != nil {
		return err
	}
	f := feature.(*FormLoginFeature)

	if err := c.configureErrorHandling(f, ws); err != nil {
		return err
	}

	if c.configured {
		logger.WithContext(ws.Context()).Warnf(`attempting to reconfigure login forms for WebSecurity [%v]. `+
			`Changes will not be applied. If this is expected, please ignore this warning`, ws)
		return nil
	}

	c.configured = true

	if err := c.configureLoginPage(f, ws); err != nil {
		return err
	}

	if err := c.configureMfaPage(f, ws); err != nil {
		return err
	}

	if err := c.configureLoginProcessing(f, ws); err != nil {
		return err
	}

	if err := c.configureMfaProcessing(f, ws); err != nil {
		return err
	}

	if err := c.configureCSRF(f, ws); err != nil {
		return err
	}

	return nil
}

func (c *FormLoginConfigurer) validate(f *FormLoginFeature, _ security.WebSecurity) error {
	if f.loginUrl == "" {
		return fmt.Errorf("loginUrl is missing for form login")
	}

	if f.successHandler == nil {
		f.successHandler = request_cache.NewSavedRequestAuthenticationSuccessHandler(
			redirect.NewRedirectWithRelativePath("/", true),
			security.IsBeingAuthenticated,
		)
	}

	if f.loginProcessUrl == "" {
		f.loginProcessUrl = f.loginUrl
	}

	if f.loginErrorUrl == "" && f.failureHandler == nil {
		f.loginErrorUrl = f.loginUrl + "?error=true"
	}

	if f.mfaEnabled && f.mfaUrl == "" {
		return fmt.Errorf("mfaUrl is missing for MFA")
	}

	if f.mfaEnabled && f.mfaVerifyUrl == "" {
		f.mfaVerifyUrl = f.mfaUrl
	}

	if f.mfaErrorUrl == "" && f.failureHandler == nil {
		f.mfaErrorUrl = f.mfaUrl + "?error=true"
	}

	return nil
}

func (c *FormLoginConfigurer) configureErrorHandling(f *FormLoginFeature, ws security.WebSecurity) error {
	errorRedirect := redirect.NewRedirectWithURL(f.loginErrorUrl)
	mfaErrorRedirect := redirect.NewRedirectWithURL(f.mfaErrorUrl)

	if f.failureHandler == nil {
		f.failureHandler = errorRedirect
	}

	var entryPoint security.AuthenticationEntryPoint = redirect.NewRedirectWithURL(f.loginUrl)
	if f.mfaEnabled {
		if _, ok := f.failureHandler.(*MfaAwareAuthenticationErrorHandler); !ok {
			f.failureHandler = &MfaAwareAuthenticationErrorHandler{
				delegate:           f.failureHandler,
				mfaPendingDelegate: mfaErrorRedirect,
			}
		}

		entryPoint = &MfaAwareAuthenticationEntryPoint{
			delegate:           entryPoint,
			mfaPendingDelegate: redirect.NewRedirectWithURL(f.mfaUrl),
		}
	}

	// override entry point and error handler
	errorhandling.Configure(ws).
		AuthenticationEntryPoint(request_cache.NewSaveRequestEntryPoint(entryPoint)).
		AuthenticationErrorHandler(f.failureHandler)

	// adding CSRF protection err handler, while keeping default
	csrf.Configure(ws).CsrfDeniedHandler(errorRedirect)

	return nil
}

func (c *FormLoginConfigurer) configureLoginPage(f *FormLoginFeature, ws security.WebSecurity) error {
	// let ws know to intercept additional url
	routeMatcher := matcher.RouteWithURL(f.loginUrl, http.MethodGet)
	requestMatcher := matcher.RequestWithURL(f.loginUrl, http.MethodGet)
	ws.Route(routeMatcher)

	// configure access
	access.Configure(ws).
		Request(requestMatcher).WithOrder(order.Highest).PermitAll()

	return nil
}

func (c *FormLoginConfigurer) configureMfaPage(f *FormLoginFeature, ws security.WebSecurity) error {
	// let ws know to intercept additional url
	routeMatcher := matcher.RouteWithURL(f.mfaUrl, http.MethodGet)
	requestMatcher := matcher.RequestWithURL(f.mfaUrl, http.MethodGet)
	ws.Route(routeMatcher)

	// configure access
	access.Configure(ws).
		Request(requestMatcher).WithOrder(order.Highest).
		HasPermissions(passwd.SpecialPermissionMFAPending, passwd.SpecialPermissionOtpId)

	return nil
}

func (c *FormLoginConfigurer) configureLoginProcessing(f *FormLoginFeature, ws security.WebSecurity) error {

	// let ws know to intercept additional url
	route := matcher.RouteWithURL(f.loginProcessUrl, http.MethodPost)
	ws.Route(route)

	// configure middlewares
	// Note: since this MW handles a new path, we add middleware as-is instead of a security.MiddlewareTemplate

	login := NewFormAuthenticationMiddleware(func(opts *FormAuthMWOptions) {
		opts.Authenticator = ws.Authenticator()
		opts.SuccessHandler = c.effectiveSuccessHandler(f, ws)
		opts.UsernameParam = f.usernameParam
		opts.PasswordParam = f.passwordParam
	})
	mw := middleware.NewBuilder("form login").
		ApplyTo(route).
		Order(security.MWOrderFormAuth).
		Use(login.LoginProcessHandlerFunc())

	ws.Add(mw)

	// configure additional endpoint mappings to trigger middleware
	ws.Add(mapping.Post(f.loginProcessUrl).
		HandlerFunc(security.NoopHandlerFunc()).
		Name("login process dummy"))

	return nil
}

func (c *FormLoginConfigurer) configureMfaProcessing(f *FormLoginFeature, ws security.WebSecurity) error {

	// let ws know to intercept additional url
	routeVerify := matcher.RouteWithURL(f.mfaVerifyUrl, http.MethodPost)
	routeRefresh := matcher.RouteWithURL(f.mfaRefreshUrl, http.MethodPost)
	requestMatcher := matcher.RequestWithURL(f.mfaVerifyUrl, http.MethodPost).
		Or(matcher.RequestWithURL(f.mfaRefreshUrl, http.MethodPost))
	ws.Route(routeVerify).Route(routeRefresh)

	// configure middlewares
	// Note: since this MW handles a new path, we add middleware as-is instead of a security.MiddlewareTemplate
	login := NewMfaAuthenticationMiddleware(func(opts *MfaMWOptions) {
		opts.Authenticator = ws.Authenticator()
		opts.SuccessHandler = c.effectiveSuccessHandler(f, ws)
		opts.OtpParam = f.otpParam
	})

	verifyMW := middleware.NewBuilder("otp verify").
		ApplyTo(routeVerify).
		Order(security.MWOrderFormAuth).
		Use(login.OtpVerifyHandlerFunc())

	refreshMW := middleware.NewBuilder("otp refresh").
		ApplyTo(routeRefresh).
		Order(security.MWOrderFormAuth).
		Use(login.OtpRefreshHandlerFunc())

	ws.Add(verifyMW, refreshMW)

	// configure additional endpoint mappings to trigger middleware
	ws.Add(mapping.Post(f.mfaVerifyUrl).
		HandlerFunc(security.NoopHandlerFunc()).
		Name("otp verify dummy"))
	ws.Add(mapping.Post(f.mfaRefreshUrl).
		HandlerFunc(security.NoopHandlerFunc()).
		Name("otp refresh dummy"))

	// configure access
	access.Configure(ws).
		Request(requestMatcher).WithOrder(order.Highest).
		HasPermissions(passwd.SpecialPermissionMFAPending, passwd.SpecialPermissionOtpId)

	return nil
}

func (c *FormLoginConfigurer) configureCSRF(f *FormLoginFeature, ws security.WebSecurity) error {
	csrfMatcher := matcher.RequestWithURL(f.loginProcessUrl, http.MethodPost).
		Or(matcher.RequestWithURL(f.mfaVerifyUrl, http.MethodPost)).
		Or(matcher.RequestWithURL(f.mfaRefreshUrl, http.MethodPost))
	csrf.Configure(ws).AddCsrfProtectionMatcher(csrfMatcher)
	return nil
}

func (c *FormLoginConfigurer) effectiveSuccessHandler(f *FormLoginFeature, ws security.WebSecurity) security.AuthenticationSuccessHandler {
	if _, ok := f.successHandler.(*MfaAwareSuccessHandler); f.mfaEnabled && !ok {
		f.successHandler = &MfaAwareSuccessHandler{
			delegate:           f.successHandler,
			mfaPendingDelegate: redirect.NewRedirectWithURL(f.mfaUrl),
		}
	}

	rememberUsernameHandler := newRememberUsernameSuccessHandler(func(h *RememberUsernameSuccessHandler) {
		h.contextPath = c.serverProps.ContextPath
		h.rememberParam = f.rememberParam
		h.cookieDomain = f.rememberCookieDomain
		h.cookieSecured = f.rememberCookieSecured
		h.cookieHttpOnly = !f.rememberCookieSecured
		h.cookieMaxAge = f.rememberCookieValidity
	})

	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler); ok {
		return security.NewAuthenticationSuccessHandler(globalHandler, rememberUsernameHandler, f.successHandler)
	} else {
		return security.NewAuthenticationSuccessHandler(rememberUsernameHandler, f.successHandler)
	}
}
