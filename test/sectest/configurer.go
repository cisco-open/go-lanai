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

package sectest

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"github.com/cisco-open/go-lanai/pkg/web/middleware"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"go.uber.org/fx"
	"net/http"
)

/**************************
	Context
 **************************/

// MWMockContext value carrier for mocking authentication in MW
type MWMockContext struct {
	Request *http.Request
}

// MWMocker interface that mocked authentication middleware uses to mock authentication at runtime
type MWMocker interface {
	Mock(MWMockContext) security.Authentication
}

/**************************
	Test Options
 **************************/

type MWMockOptions func(opt *MWMockOption)

type MWMockOption struct {
	Route         web.RouteMatcher
	Condition     web.RequestMatcher
	MWMocker      MWMocker
	MWOrder       int
	Configurer    security.Configurer
	Session       bool
	ForceOverride bool
}

var defaultMWMockOption = MWMockOption{
	MWOrder:  security.MWOrderPreAuth + 5,
	MWMocker: DirectExtractionMWMocker{},
	Route:    matcher.AnyRoute(),
}

// WithMockedMiddleware is a test option that automatically install a middleware that populate/save
// security.Authentication into gin.Context.
//
// This test option works with webtest.WithMockedServer without any additional settings:
// - By default extract security.Authentication from request's context.
// Note: 	Since gin-gonic v1.8.0+, this test option is not required anymore for webtest.WithMockedServer. Values in
//
//	request's context is automatically linked with gin.Context.
//
// When using with webtest.WithRealServer, a custom MWMocker is required. The MWMocker can be provided by:
//   - Using MWCustomMocker option
//   - Providing a MWMocker using uber/fx
//   - Providing a security.Configurer with NewMockedMW:
//     <code>
//     func realServerSecConfigurer(ws security.WebSecurity) {
//     ws.Route(matcher.AnyRoute()).
//     With(NewMockedMW().
//     Mocker(MWMockFunc(realServerMockFunc)),
//     )
//     }
//     </code>
//
// See examples package for more details.
func WithMockedMiddleware(opts ...MWMockOptions) test.Options {
	opt := defaultMWMockOption
	for _, fn := range opts {
		fn(&opt)
	}
	testOpts := []test.Options{
		apptest.WithModules(security.Module),
		apptest.WithFxOptions(
			fx.Invoke(registerSecTest),
		),
	}
	if opt.MWMocker != nil {
		testOpts = append(testOpts, apptest.WithFxOptions(fx.Provide(func() MWMocker { return opt.MWMocker })))
	}
	if opt.Configurer != nil {
		testOpts = append(testOpts, apptest.WithFxOptions(fx.Invoke(func(reg security.Registrar) {
			reg.Register(opt.Configurer)
		})))
	} else {
		testOpts = append(testOpts, apptest.WithFxOptions(fx.Invoke(RegisterTestConfigurer(opts...))))
	}
	if opt.Session {
		testOpts = append(testOpts,
			apptest.WithModules(session.Module),
			apptest.WithFxOptions(fx.Decorate(MockedSessionStoreDecorator)),
		)
	}
	return test.WithOptions(testOpts...)
}

// MWRoute returns option for WithMockedMiddleware.
// This route is applied to the default test security.Configurer
func MWRoute(matchers ...web.RouteMatcher) MWMockOptions {
	return func(opt *MWMockOption) {
		for i, m := range matchers {
			if i == 0 {
				opt.Route = m
			} else {
				opt.Route = opt.Route.Or(m)
			}
		}
	}
}

// MWCondition returns option for WithMockedMiddleware.
// This condition is applied to the default test security.Configurer
func MWCondition(matchers ...web.RequestMatcher) MWMockOptions {
	return func(opt *MWMockOption) {
		for i, m := range matchers {
			if i == 0 {
				opt.Condition = m
			} else {
				opt.Condition = opt.Route.Or(m)
			}
		}
	}
}

// MWEnableSession returns option for WithMockedMiddleware.
// Enabling in-memory session
func MWEnableSession() MWMockOptions {
	return func(opt *MWMockOption) {
		opt.Session = true
	}
}

// MWForcePreOAuth2AuthValidation returns option for WithMockedMiddleware.
// Decrease the order of mocking middleware such that it runs before OAuth2 authorize validation.
func MWForcePreOAuth2AuthValidation() MWMockOptions {
	return func(opt *MWMockOption) {
		opt.MWOrder = security.MWOrderOAuth2AuthValidation - 5
	}
}

// MWForceOverride returns option for WithMockedMiddleware.
// Add a middleware after the last auth middleware (before access control) that override any other installed authenticators.
func MWForceOverride() MWMockOptions {
	return func(opt *MWMockOption) {
		opt.ForceOverride = true
	}
}

// MWCustomConfigurer returns option for WithMockedMiddleware.
// If set to nil, MWMockOption.Route and MWMockOption.Condition are used to generate a default configurer
// If set to non-nil, MWMockOption.Route and MWMockOption.Condition are ignored
func MWCustomConfigurer(configurer security.Configurer) MWMockOptions {
	return func(opt *MWMockOption) {
		opt.Configurer = configurer
	}
}

// MWCustomMocker returns option for WithMockedMiddleware.
// If set to nil, fx provided MWMocker will be used
func MWCustomMocker(mocker MWMocker) MWMockOptions {
	return func(opt *MWMockOption) {
		opt.MWMocker = mocker
	}
}

/**************************
	Mockers
 **************************/

// MWMockFunc wrap a function to MWMocker interface
type MWMockFunc func(MWMockContext) security.Authentication

func (f MWMockFunc) Mock(mc MWMockContext) security.Authentication {
	return f(mc)
}

// DirectExtractionMWMocker is an MWMocker that extracts authentication from context.
// This is the implementation is works together with webtest.WithMockedServer and WithMockedSecurity,
// where a context is injected with security.Authentication and directly passed into http.Request
type DirectExtractionMWMocker struct{}

func (m DirectExtractionMWMocker) Mock(mc MWMockContext) security.Authentication {
	return security.Get(mc.Request.Context())
}

/**************************
	Feature
 **************************/

var (
	FeatureId = security.FeatureId("SecTest", security.FeatureOrderAuthenticator)
)

type regDI struct {
	fx.In
	SecRegistrar security.Registrar `optional:"true"`
}

func registerSecTest(di regDI) {
	if di.SecRegistrar != nil {
		configurer := newFeatureConfigurer()
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}

type Feature struct {
	MWOrder  int
	MWMocker MWMocker
	Override bool
}

// NewMockedMW Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func NewMockedMW() *Feature {
	return &Feature{
		MWOrder:  defaultMWMockOption.MWOrder,
		MWMocker: defaultMWMockOption.MWMocker,
	}
}

func (f *Feature) Order(mwOrder int) *Feature {
	f.MWOrder = mwOrder
	return f
}

func (f *Feature) Mocker(mocker MWMocker) *Feature {
	f.MWMocker = mocker
	return f
}

func (f *Feature) ForceOverride(override bool) *Feature {
	f.Override = override
	return f
}

func (f *Feature) MWMockFunc(mocker MWMockFunc) *Feature {
	f.MWMocker = mocker
	return f
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *Feature {
	feature := NewMockedMW()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

type FeatureConfigurer struct {
}

func newFeatureConfigurer() *FeatureConfigurer {
	return &FeatureConfigurer{}
}

func (c *FeatureConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)
	mock := &MockAuthenticationMiddleware{
		MWMocker: f.MWMocker,
	}
	mw := middleware.NewBuilder("mocked-auth-mw").
		Order(f.MWOrder).
		Use(mock.AuthenticationHandlerFunc())
	ws.Add(mw)

	if f.Override {
		overrideMW := middleware.NewBuilder("mocked-auth-override-mw").
			Order(security.MWOrderAccessControl - 5).
			Use(mock.ForceOverrideHandlerFunc())
		ws.Add(overrideMW)
	}

	return nil
}

/**************************
	Security Configurer
 **************************/

type mwDI struct {
	fx.In
	Registrar security.Registrar `optional:"true"`
	Mocker    MWMocker           `optional:"true"`
}

func RegisterTestConfigurer(opts ...MWMockOptions) func(di mwDI) {
	opt := defaultMWMockOption
	for _, fn := range opts {
		fn(&opt)
	}
	return func(di mwDI) {
		if opt.MWMocker == nil {
			opt.MWMocker = di.Mocker
		}
		configurer := security.ConfigurerFunc(newTestSecurityConfigurer(&opt))
		di.Registrar.Register(configurer)
	}
}

func newTestSecurityConfigurer(opt *MWMockOption) func(ws security.WebSecurity) {
	return func(ws security.WebSecurity) {
		ws = ws.Route(opt.Route).With(NewMockedMW().
			Order(opt.MWOrder).
			Mocker(opt.MWMocker).
			ForceOverride(opt.ForceOverride),
		)
		if opt.Condition != nil {
			ws.Condition(opt.Condition)
		}
	}
}
