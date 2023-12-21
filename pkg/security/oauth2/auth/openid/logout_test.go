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

package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
	"time"
)

const (
	TestRPUrl                   = "http://some.domain"
	TestOidcSloPath             = "/logout"
	TestLogoutErrorURL          = "/error"
	TestNonOIDCLogoutSuccessURL = "/logout/success"
	TestDefaultKid              = "default"
	TestAuthServerProtocol      = "http"
	TestAuthServerDomain        = "localhost"
	TestAuthServerPort          = 8900
	TestUsername                = "my-name"
	TestClientID                = "test-client"
	TestClientSecret            = "test-client-secret"
)

var (
	jwtStore = jwt.NewSingleJwkStore(TestDefaultKid)
	jwtDec   = jwt.NewRS256JwtDecoder(jwtStore, TestDefaultKid)
	jwtEnc   = jwt.NewRS256JwtEncoder(jwtStore, TestDefaultKid)
	issuer   = security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
		*opt = security.DefaultIssuerDetails{
			Protocol:    TestAuthServerProtocol,
			Domain:      TestAuthServerDomain,
			Port:        TestAuthServerPort,
			ContextPath: webtest.DefaultContextPath,
			IncludePort: true,
		}
	})
)

type TestLogoutSecConfigurer struct{}

func (c *TestLogoutSecConfigurer) Configure(ws security.WebSecurity) {
	clientStore := sectest.NewMockedClientStore(&sectest.MockedClientProperties{
		ClientID:     TestClientID,
		Secret:       TestClientSecret,
		GrantTypes:   utils.CommaSeparatedSlice{"authorization_code"},
		Scopes:       utils.CommaSeparatedSlice{"openid", "profile", "email", "address", "phone", "read", "write"},
		RedirectUris: utils.CommaSeparatedSlice{TestRPUrl},
		ATValidity:   utils.Duration(time.Hour),
		RTValidity:   utils.Duration(time.Hour),
	})

	oidcLogoutHandler := NewOidcLogoutHandler(func(opt *HandlerOption) {
		opt.Dec = jwtDec
		opt.Issuer = issuer
		opt.ClientStore = clientStore
	})
	oidcLogoutSuccessHandler := NewOidcSuccessHandler(func(opt *SuccessOption) {
		opt.ClientStore = clientStore
		opt.WhitelabelErrorPath = TestLogoutErrorURL
	})
	oidcEntryPoint := NewOidcEntryPoint(func(opt *EpOption) {
		opt.WhitelabelErrorPath = TestLogoutErrorURL
	})

	ws.Route(matcher.AnyRoute()).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New()).
		With(request_cache.New()).
		With(logout.New().
			LogoutUrl(TestOidcSloPath).
			//oidc handler and success handler
			AddLogoutHandler(oidcLogoutHandler).
			AddSuccessHandler(oidcLogoutSuccessHandler).
			AddEntryPoint(oidcEntryPoint).
			// in real configuration, this would be the success handler for other logout protocols
			// we expect them to not get called when the logout is for OIDC
			AddSuccessHandler(redirect.NewRedirectWithURL(TestNonOIDCLogoutSuccessURL)).
			AddErrorHandler(UselessHandler{}).
			AddSuccessHandler(UselessHandler{}),
		)
}

type initDI struct {
	fx.In
	SecurityRegistrar security.Registrar
}

func ConfigureLogout(di initDI) {
	lc := &TestLogoutSecConfigurer{}
	di.SecurityRegistrar.Register(lc)
}

func TestWithMockedServer(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(logout.Module, request_cache.Module, access.Module, errorhandling.Module),
		apptest.WithFxOptions(
			fx.Invoke(ConfigureLogout),
		),
		test.GomegaSubTest(SubTestLogoutSuccess(), "TestLogoutSuccess"),
		test.GomegaSubTest(SubTestLogoutWithWrongIdTokenHint(), "TestLogoutWithWrongIdTokenHint"),
		test.GomegaSubTest(SubTestLogoutWithInvalidIdTokenHint(), "TestLogoutWithInvalidIdTokenHint"),
		test.GomegaSubTest(SubTestLogoutWithInvalidRedirect(), "TestLogoutWithInvalidRedirect"),
		test.GomegaSubTest(SubTestLogoutWithoutRedirect(), "TestLogoutWithoutRedirect"))
}

func SubTestLogoutSuccess() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		claims := oauth2.MapClaims{
			"aud": TestClientID,
			"iss": issuer.Identifier(),
			"sub": TestUsername,
		}

		idToken, err := jwtEnc.Encode(ctx, claims)
		g.Expect(err).ToNot(HaveOccurred())

		state := "some-state"

		ctx = sectest.ContextWithSecurity(ctx, mockedAuthentication())
		req := webtest.NewRequest(ctx, http.MethodGet, "http://localhost:8900/test/logout", nil,
			webtest.Queries("post_logout_redirect_uri", "http://some.domain",
				"id_token_hint", idToken,
				"state", state))

		resp := webtest.MustExec(ctx, req).Response

		g.Expect(resp.StatusCode).To(Equal(http.StatusFound))
		location, err := resp.Location()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(location.String()).To(Equal(fmt.Sprintf("http://some.domain?state=%s", state)))
	}
}

func SubTestLogoutWithWrongIdTokenHint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		claims := oauth2.MapClaims{
			"aud": TestClientID,
			"iss": issuer.Identifier(),
			"sub": "some-other-user",
		}

		idToken, err := jwtEnc.Encode(ctx, claims)
		g.Expect(err).ToNot(HaveOccurred())

		state := "some-state"

		ctx = sectest.ContextWithSecurity(ctx, mockedAuthentication())
		req := webtest.NewRequest(ctx, http.MethodGet, "http://localhost:8900/test/logout", nil,
			webtest.Queries("post_logout_redirect_uri", "http://some.domain",
				"id_token_hint", idToken,
				"state", state))

		resp := webtest.MustExec(ctx, req).Response

		g.Expect(resp.StatusCode).To(Equal(http.StatusFound))
		location, err := resp.Location()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(location.String()).To(Equal("/test/error"))
	}
}

func SubTestLogoutWithInvalidIdTokenHint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		anotherJwtStore := jwt.NewSingleJwkStore(TestDefaultKid)
		anotherEnc := jwt.NewRS256JwtEncoder(anotherJwtStore, TestDefaultKid)

		claims := oauth2.MapClaims{
			"aud": TestClientID,
			"iss": issuer.Identifier(),
			"sub": TestUsername,
		}
		idToken, err := anotherEnc.Encode(ctx, claims)
		g.Expect(err).ToNot(HaveOccurred())

		state := "some-state"

		ctx = sectest.ContextWithSecurity(ctx, mockedAuthentication())
		req := webtest.NewRequest(ctx, http.MethodGet, "http://localhost:8900/test/logout", nil,
			webtest.Queries("post_logout_redirect_uri", "http://some.domain",
				"id_token_hint", idToken,
				"state", state))

		resp := webtest.MustExec(ctx, req).Response

		g.Expect(resp.StatusCode).To(Equal(http.StatusFound))
		location, err := resp.Location()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(location.String()).To(Equal("/test/error"))
	}
}

func SubTestLogoutWithInvalidRedirect() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		claims := oauth2.MapClaims{
			"aud": TestClientID,
			"iss": issuer.Identifier(),
			"sub": TestUsername,
		}

		idToken, err := jwtEnc.Encode(ctx, claims)
		g.Expect(err).ToNot(HaveOccurred())

		state := "some-state"

		ctx = sectest.ContextWithSecurity(ctx, mockedAuthentication())
		req := webtest.NewRequest(ctx, http.MethodGet, "http://localhost:8900/test/logout", nil,
			webtest.Queries("post_logout_redirect_uri", "http://unregistered.domain",
				"id_token_hint", idToken,
				"state", state))

		resp := webtest.MustExec(ctx, req).Response

		g.Expect(resp.StatusCode).To(Equal(http.StatusFound))
		location, err := resp.Location()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(location.String()).To(Equal("/test/error"))
	}
}

// we expect to go to our default logout success endpoint
func SubTestLogoutWithoutRedirect() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.ContextWithSecurity(ctx, mockedAuthentication())
		req := webtest.NewRequest(ctx, http.MethodGet, "http://localhost:8900/test/logout", nil)

		resp := webtest.MustExec(ctx, req).Response

		g.Expect(resp.StatusCode).To(Equal(http.StatusFound))
		location, err := resp.Location()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(location.String()).To(Equal(fmt.Sprintf("/test/logout/success")))
	}
}

type UselessHandler struct{}

func (h UselessHandler) HandleAuthenticationSuccess(ctx context.Context, _ *http.Request, rw http.ResponseWriter, _, _ security.Authentication) {
	h.doHandle(ctx, rw)
}

func (h UselessHandler) HandleAuthenticationError(ctx context.Context, _ *http.Request, rw http.ResponseWriter, _ error) {
	h.doHandle(ctx, rw)
}

func (h UselessHandler) doHandle(_ context.Context, rw http.ResponseWriter) {
	if grw, ok := rw.(gin.ResponseWriter); ok && grw.Written() {
		return
	}
	rw.WriteHeader(http.StatusUnauthorized)
	_, _ = rw.Write([]byte("this should not happen"))
}

func mockedAuthentication(opts ...sectest.SecurityMockOptions) sectest.SecurityContextOptions {
	opts = append([]sectest.SecurityMockOptions{
		func(d *sectest.SecurityDetailsMock) {
			d.Username = TestUsername
		},
	}, opts...)
	return func(opt *sectest.SecurityContextOption) {
		mock := sectest.SecurityDetailsMock{}
		for _, fn := range opts {
			fn(&mock)
		}
		opt.Authentication = &sectest.MockedAccountAuthentication{
			Account: sectest.MockedAccount{
				MockedAccountDetails: sectest.MockedAccountDetails{
					UserId:          mock.UserId,
					Username:        mock.Username,
					TenantId:        mock.TenantId,
					DefaultTenant:   mock.TenantId,
					AssignedTenants: mock.Tenants,
					Permissions:     mock.Permissions,
				},
			},
			AuthState: security.StateAuthenticated,
		}
	}
}
