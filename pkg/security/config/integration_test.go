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

package config

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/extsamlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/authorize"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/clientauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/token"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	samlidp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/idp"
	samlsp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/sp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/embedded"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/crewjam/saml"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// go:embed testdata/template/*.tmpl
//var TestHTMLTemplates embed.FS

/*************************
	Setup
 *************************/

const (
	ExpectedAuthorizeCallback = `http://localhost/test/callback`
	TestClientID              = "test-client"
	TestTenantedClientID1     = "tenant-client-1"
	TestTenantedClientID2     = "tenant-client-2"
	TestTenantedClientID3     = "tenant-client-3"
	TestClientSecret          = "test-secret"
	TestOAuth2CallbackURL     = "http://localhost/oauth/callback"
)

const (
	PermissionSwitchTenant                = "SWITCH_TENANT"
	PermissionAccessAllTenants            = "ACCESS_ALL_TENANTS"
	PermissionViewOperatorLoginAsCustomer = "VIEW_OPERATOR_LOGIN_AS_CUSTOMER"
)

// TestMain is the only place we should kick off embedded redis
func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		embedded.Redis(),
	)
}

type IntegrationTestDI struct {
	fx.In
	AppCtx  *bootstrap.ApplicationContext
	SecReg  security.Registrar
	WebReg  *web.Registrar
	Mocking sectest.MockingProperties
}

type IntegrationTestOut struct {
	fx.Out
	DiscoveryCustomizers *discovery.Customizers
	IdpManager           idp.IdentityProviderManager
	AccountStore         security.AccountStore
	PasswordEncoder      passwd.PasswordEncoder
	FedAccountStore      security.FederatedAccountStore
	SamlClientStore      samlctx.SamlClientStore
	TenancyAccessor      tenancy.Accessor
}

type IntegrationTestOption func(di IntegrationTestDI, out *IntegrationTestOut)

func IntegrationTestMocksProvider(opts ...IntegrationTestOption) func(IntegrationTestDI) IntegrationTestOut {
	return func(di IntegrationTestDI) IntegrationTestOut {
		mockTenantAccessor := mocks.NewMockTenancyAccessorUsingStrIds([]mocks.TenancyRelationWithStrId{
			{ParentId: "id-tenant-root", ChildId: "id-tenant-1"},
			{ParentId: "id-tenant-root", ChildId: "id-tenant-2"},
			{ParentId: "id-tenant-root", ChildId: "id-tenant-3"},
		}, "id-tenant-root")

		integrationTestOut := IntegrationTestOut{
			DiscoveryCustomizers: &discovery.Customizers{},
			IdpManager:           testdata.NewMockedIDPManager(),
			AccountStore:         sectest.NewMockedAccountStore(di.Mocking.Accounts.Values()),
			PasswordEncoder:      passwd.NewNoopPasswordEncoder(),
			FedAccountStore:      sectest.NewMockedFederatedAccountStore(di.Mocking.FederatedUsers.Values()...),
			SamlClientStore:      samltest.NewMockedClientStore(samltest.ClientsWithPropertiesPrefix(di.AppCtx.Config(), "mocking.clients")),
			TenancyAccessor:      mockTenantAccessor,
		}
		for _, opt := range opts {
			opt(di, &integrationTestOut)
		}

		return integrationTestOut
	}
}

/*************************
	Test
 *************************/

type intDI struct {
	fx.In
	FedAccountStore security.FederatedAccountStore
	Mocking         sectest.MockingProperties
	TokenReader     oauth2.TokenStoreReader
}

func TestWithMockedServer(t *testing.T) {
	di := &intDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(2*time.Minute),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession(), sectest.MWForcePreOAuth2AuthValidation()),
		apptest.WithModules(
			authserver.Module, resserver.Module,
			passwdidp.Module, extsamlidp.Module, authorize.Module, samlidp.Module,
			passwd.Module, formlogin.Module, logout.Module,
			samlctx.Module, samlsp.Module,
			basicauth.Module, clientauth.Module,
			token.Module, access.Module, errorhandling.Module,
			request_cache.Module, csrf.Module, session.Module,
			redis.Module, tenancy.Module,
		),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(
				IntegrationTestMocksProvider(func(di IntegrationTestDI, out *IntegrationTestOut) {
					out.AccountStore = sectest.NewMockedAccountStoreWithFinalize(
						di.Mocking.Accounts.Values(),
						di.Mocking.Tenants.Values(),
					)
				}),
				sectest.BindMockingProperties,
				testdata.NewAuthServerConfigurer, //This configurer will set up mocked client store, mocked tenant store etc.
				testdata.NewResServerConfigurer,
			),
		),
		test.GomegaSubTest(SubTestOAuth2AuthorizeWithPasswdIDP(di), "TestOAuth2AuthorizeWithPasswdIDP"),
		test.GomegaSubTest(SubTestOAuth2AuthorizeWithSamlSSO(di), "TestOAuth2AuthorizeWithSamlSSO"),
		test.GomegaSubTest(SubTestSamlSSOAuthorizeWithPasswdIDP(di), "TestSamlSSOAuthorizeWithPasswdIDP"),
		//token tests
		test.GomegaSubTest(SubTestOAuth2AuthCode(di), "TestOAuth2AuthCode"),
		test.GomegaSubTest(SubTestOAuth2AuthCodeWithoutTenant(di), "TestOAuth2AuthCodeWithoutTenant"),
		test.GomegaSubTest(SubTestOAuth2AuthCodeWithTenantClient(di), "TestOAuth2AuthCodeWithTenantClient"),
		test.GomegaSubTest(SubTestOAuth2PasswordGrant(di), "TestOAuth2PasswordGrant"),
		test.GomegaSubTest(SubTestTenantClientCredential(di), "TestTenantClientCredential"),

		//switch tenants
		test.GomegaSubTest(SubTestOauth2SwitchTenantWithPerTenantPermission(di), "TestOauth2SwitchTenantWithPerTenantPermission"),
		test.GomegaSubTest(SubTestOauth2AccessCodeSwitchTenant(di), "TestOauth2AccessCodeSwitchTenant"),
	)
}

func TestWithMockedServerWithoutFinalizer(t *testing.T) {
	di := &intDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(2*time.Minute),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(
			authserver.Module, resserver.Module,
			passwdidp.Module, extsamlidp.Module, authorize.Module, samlidp.Module,
			passwd.Module, formlogin.Module, logout.Module,
			samlctx.Module, samlsp.Module,
			basicauth.Module, clientauth.Module,
			token.Module, access.Module, errorhandling.Module,
			request_cache.Module, csrf.Module, session.Module,
			redis.Module,
		),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(
				IntegrationTestMocksProvider(),
				sectest.BindMockingProperties,
				testdata.NewAuthServerConfigurer,
				testdata.NewResServerConfigurer,
			),
		),
		// a user has access to two tenants, switch from one to the other
		// the permission is not per tenant, so user permission doesn't change
		test.GomegaSubTest(SubTestOAuth2SwitchTenantNoFinalizer(di), "TestOauth2SwitchTenant"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestOAuth2AuthorizeWithPasswdIDP(_ *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		uri := fmt.Sprintf("http://%s/test/v2/authorize", testdata.IdpDomainPasswd)
		req = webtest.NewRequest(ctx, http.MethodGet, uri, nil,
			withDefaultAuthCode())
		resp = webtest.MustExec(ctx, req).Response
		fmt.Printf("%v\n", resp)
		assertRedirectResponse(t, g, resp, "/test/login")
	}
}

func SubTestOAuth2AuthorizeWithSamlSSO(_ *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		uri := fmt.Sprintf("http://%s/test/v2/authorize", testdata.IdpDomainExtSAML)
		req = webtest.NewRequest(ctx, http.MethodGet, uri, nil,
			withDefaultAuthCode())
		resp = webtest.MustExec(ctx, req).Response
		fmt.Printf("%v\n", resp)
		assertRedirectResponse(t, g, resp, testdata.ExtSamlIdpSSOUrl)
	}
}

func SubTestSamlSSOAuthorizeWithPasswdIDP(_ *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		uri := fmt.Sprintf("http://%s/test/v2/authorize", testdata.IdpDomainPasswd)
		req = webtest.NewRequest(ctx, http.MethodPost, uri, nil, withDefaultSamlSSO())
		resp = webtest.MustExec(ctx, req).Response
		fmt.Printf("%v\n", resp)
		assertRedirectResponse(t, g, resp, "/test/login")
	}
}

func SubTestOAuth2AuthCode(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// mock authentication
		fedAccount := di.Mocking.FederatedUsers.MapValues()["fed1"]
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		// authorize
		req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, authorizeReqOptions("test-client"))
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusFound), "response should have correct status code")
		assertAuthorizeResponse(t, g, resp.Response, false)

		// token
		code := extractAuthCode(resp.Response)
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code, TestClientID, ""), tokenReqOptions())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
		a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)

		auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
		userDetail, ok := auth.Details().(security.UserDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(userDetail.UserId()).To(Equal(fedAccount.UserId))
		tenantDetail, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(tenantDetail.TenantId()).To(Equal(fedAccount.DefaultTenant))
		providerDetail, ok := auth.Details().(security.ProviderDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(providerDetail.ProviderId()).To(Not(BeEmpty()))
	}
}

func SubTestOAuth2AuthCodeWithoutTenant(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// mock authentication
		fedAccount := di.Mocking.FederatedUsers.MapValues()["fed2"]
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		// authorize
		req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, authorizeReqOptions("test-client"))
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusFound), "response should have correct status code")
		assertAuthorizeResponse(t, g, resp.Response, false)

		// token
		code := extractAuthCode(resp.Response)
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code, TestClientID, ""), tokenReqOptions())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
		a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)

		auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
		userDetail, ok := auth.Details().(security.UserDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(userDetail.UserId()).To(Equal(fedAccount.UserId))
		_, ok = auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeFalse())
		_, ok = auth.Details().(security.ProviderDetails)
		g.Expect(ok).To(BeFalse())
	}
}

type AuthCodeWithTenantClientTestStruct struct {
	name                               string
	clientId                           string
	expectAuthorizeErr                 bool
	expectedAuthTenantId               string
	expectedEffectiveAssignedTenantIds []string
	expectedAuthProviderId             string
}

func SubTestOAuth2AuthCodeWithTenantClient(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		fedAccount := di.Mocking.FederatedUsers.MapValues()["fed1"] //this user has access to 2 tenants
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		tests := []AuthCodeWithTenantClientTestStruct{
			{
				name:                               "User tenants and client tenants has no overlap",
				clientId:                           TestTenantedClientID3,
				expectAuthorizeErr:                 false,
				expectedAuthTenantId:               "",
				expectedEffectiveAssignedTenantIds: nil,
				expectedAuthProviderId:             "",
			},
			{
				name:                               "User with client that has access to all tenants",
				clientId:                           TestClientID,
				expectAuthorizeErr:                 false,
				expectedAuthTenantId:               fedAccount.DefaultTenant,
				expectedEffectiveAssignedTenantIds: fedAccount.Tenants,
				expectedAuthProviderId:             "test-provider",
			},
			{
				name:                               "User tenants overlap with client tenants",
				clientId:                           TestTenantedClientID2,
				expectAuthorizeErr:                 false,
				expectedAuthTenantId:               "",                      //this is empty because the user's default tenant is not in client's tenants.
				expectedEffectiveAssignedTenantIds: []string{"id-tenant-2"}, //this is the intersection of the user's tenants and the client's tenants
				expectedAuthProviderId:             "",                      //because there isn't a default tenant, the provider is empty. (Provider is derived from the selected tenant).
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				// authorize
				req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, authorizeReqOptions(test.clientId)) //this client has 1 tenants, no intersection with user
				resp := webtest.MustExec(ctx, req)
				g.Expect(resp).ToNot(BeNil(), "response should not be nil")
				assertAuthorizeResponse(t, g, resp.Response, test.expectAuthorizeErr)

				if !test.expectAuthorizeErr {
					code := extractAuthCode(resp.Response)
					req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code, test.clientId, ""), tokenReqOptions())
					resp = webtest.MustExec(ctx, req)
					g.Expect(resp).ToNot(BeNil(), "response should not be nil")
					g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
					a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)
					auth, err := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
					g.Expect(err).To(Not(HaveOccurred()))
					assertUserAuth(t, g, auth, fedAccount.UserId, test.expectedAuthTenantId, utils.NewStringSet(test.expectedEffectiveAssignedTenantIds...), test.expectedAuthProviderId)
				}
			})
		}
	}
}

type PasswordGrantTestStruct struct {
	name                               string
	clientId                           string
	expectedTokenRespStatus            int
	expectedAuthTenantId               string
	expectedEffectiveAssignedTenantIds []string
	expectedAuthProviderId             string
}

func SubTestOAuth2PasswordGrant(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tests := []PasswordGrantTestStruct{
			{
				name:                               "User tenants and client tenant has no overlap",
				clientId:                           TestTenantedClientID3,
				expectedTokenRespStatus:            http.StatusOK,
				expectedAuthTenantId:               "",
				expectedEffectiveAssignedTenantIds: nil,
				expectedAuthProviderId:             "",
			},
			{
				name:                               "User with client that has access to all tenants",
				clientId:                           TestClientID,
				expectedTokenRespStatus:            http.StatusOK,
				expectedAuthTenantId:               "id-tenant-1",
				expectedEffectiveAssignedTenantIds: []string{"id-tenant-1", "id-tenant-2"},
				expectedAuthProviderId:             "test-provider",
			},
			{
				name:                               "User tenants overlaps client tenants",
				clientId:                           TestTenantedClientID2,
				expectedTokenRespStatus:            http.StatusOK,
				expectedAuthTenantId:               "",                      // because default tenantId doesn't overlap with client tenant
				expectedEffectiveAssignedTenantIds: []string{"id-tenant-2"}, // the user tenants that overlaps with client tenants
				expectedAuthProviderId:             "",                      // because providerId is derived from selected tenant
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req := webtest.NewRequest(ctx, http.MethodPost, "/v2/token", passwordGrantReqBody("", "regular", "regular"), tokenReqOptions(), withClientAuth(test.clientId, TestClientSecret))
				resp := webtest.MustExec(ctx, req)
				g.Expect(resp).ToNot(BeNil(), "response should not be nil")
				g.Expect(resp.Response.StatusCode).To(Equal(test.expectedTokenRespStatus), "response should have correct status code")

				if test.expectedTokenRespStatus == http.StatusOK {
					a := assertTokenResponse(t, g, resp.Response, "regular", false)
					auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
					g.Expect(e).ToNot(HaveOccurred())
					assertUserAuth(t, g, auth, "id-regular", test.expectedAuthTenantId, utils.NewStringSet(test.expectedEffectiveAssignedTenantIds...), test.expectedAuthProviderId)
				}
			})
		}
	}
}

type ClientCredentialTestStruct struct {
	name                    string
	clientId                string
	selectTenantId          string
	expectedTenantId        string
	expectedTokenRespStatus int
	clientTenants           []string
	clientScopes            []string
}

func SubTestTenantClientCredential(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tests := []ClientCredentialTestStruct{
			{
				name:                    "client with access to root tenant, but does not select tenant",
				clientId:                TestClientID,
				selectTenantId:          "",
				expectedTenantId:        "id-tenant-root",
				expectedTokenRespStatus: http.StatusOK,
				clientTenants:           []string{"id-tenant-root"},
				clientScopes:            []string{"scope_a"},
			},
			{
				name:                    "client with access to all tenants, select tenant",
				clientId:                TestClientID,
				selectTenantId:          "id-tenant-1",
				expectedTenantId:        "id-tenant-1",
				expectedTokenRespStatus: http.StatusOK,
				clientTenants:           []string{"id-tenant-root"},
				clientScopes:            []string{"scope_a"},
			},
			{
				name:                    "client with access to multiple tenants without tenant selection",
				clientId:                TestTenantedClientID1,
				selectTenantId:          "",
				expectedTenantId:        "",
				expectedTokenRespStatus: http.StatusOK,
				clientTenants:           []string{"id-tenant-1", "id-tenant-2"},
				clientScopes:            []string{"scope_a", "scope_b"},
			},
			{
				name:                    "client to multiple tenant with tenant selection",
				clientId:                TestTenantedClientID1,
				selectTenantId:          "id-tenant-1",
				expectedTenantId:        "id-tenant-1",
				expectedTokenRespStatus: http.StatusOK,
				clientTenants:           []string{"id-tenant-1", "id-tenant-2"},
				clientScopes:            []string{"scope_a", "scope_b"},
			},
			{
				name:                    "client to multiple tenant with tenant selection that is not accessible by client",
				clientId:                TestTenantedClientID1,
				selectTenantId:          "id-tenant-3",
				expectedTokenRespStatus: http.StatusBadRequest,
			},
			{
				name:                    "client to multiple tenant with tenant selection for non-exist tenant",
				clientId:                TestTenantedClientID1,
				selectTenantId:          "id-tenant-not-exist",
				expectedTokenRespStatus: http.StatusBadRequest,
			},
			{
				name:                    "client to single tenant without selection",
				clientId:                TestTenantedClientID2,
				selectTenantId:          "",
				expectedTenantId:        "id-tenant-2", //because this client only has one tenant, so we defaulted to it
				expectedTokenRespStatus: http.StatusOK,
				clientTenants:           []string{"id-tenant-2"},
				clientScopes:            []string{"scope_a", "scope_b"},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req := webtest.NewRequest(ctx, http.MethodPost, "/v2/token", clientCredentialReqBody(test.selectTenantId), tokenReqOptions(), withClientAuth(test.clientId, TestClientSecret))
				resp := webtest.MustExec(ctx, req)
				g.Expect(resp).ToNot(BeNil(), "response should not be nil")
				g.Expect(resp.Response.StatusCode).To(Equal(test.expectedTokenRespStatus), "response should have correct status code")

				if test.expectedTokenRespStatus == http.StatusOK {
					a := assertTokenResponse(t, g, resp.Response, "", false)
					auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
					g.Expect(e).ToNot(HaveOccurred())
					assertClientCredentialAuth(t, g, auth, test.clientId, test.expectedTenantId, utils.NewStringSet(test.clientTenants...), utils.NewStringSet(test.clientScopes...))
				}
			})
		}
	}
}

func SubTestOauth2AccessCodeSwitchTenant(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// mock authentication
		fedAccount := di.Mocking.FederatedUsers.MapValues()["fed3"]
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		// authorize
		req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, func(req *http.Request) {
			req.Host = fmt.Sprintf("http://%s", testdata.IdpDomainExtSAML)
			req.URL.Host = fmt.Sprintf("http://%s", testdata.IdpDomainExtSAML)
			values := url.Values{}
			values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode)
			values.Set(oauth2.ParameterResponseType, "code")
			values.Set(oauth2.ParameterClientId, "test-client")
			values.Set(oauth2.ParameterRedirectUri, "http://localhost/test/callback")
			values.Set(oauth2.ParameterTenantId, "id-tenant-3")
			req.URL.RawQuery = values.Encode()
		})
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusFound), "response should have correct status code")
		assertAuthorizeResponse(t, g, resp.Response, false)

		// token
		code := extractAuthCode(resp.Response)
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code, TestClientID, ""), tokenReqOptions())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
		a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)

		auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
		userDetail, ok := auth.Details().(security.UserDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(userDetail.UserId()).To(Equal(fedAccount.UserId))
		tenantDetail, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(tenantDetail.TenantId()).To(Equal("id-tenant-3"))
		userDetails, ok := auth.Details().(security.AuthenticationDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(userDetails.Permissions()).To(Equal(utils.NewStringSet(
			PermissionSwitchTenant,
			PermissionAccessAllTenants,
			PermissionViewOperatorLoginAsCustomer,
		)))
		providerDetail, ok := auth.Details().(security.ProviderDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(providerDetail.ProviderId()).To(Equal("test-provider"))
	}
}

type SwitchTenantTestStruct struct {
	name            string
	tenantID        string
	username        string
	wantPermissions utils.StringSet
}

// SubTestOauth2SwitchTenantWithPerTenantPermission will test that a user can start off with no
// tenant, and then switch to different tenants and check that the permissions changes each time
// it switches
func SubTestOauth2SwitchTenantWithPerTenantPermission(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		fedAccount := di.Mocking.FederatedUsers.MapValues()["fed3"]
		tests := []SwitchTenantTestStruct{
			{
				name:     "tenant-1",
				tenantID: "id-tenant-1",
				username: fedAccount.Username,
				wantPermissions: utils.NewStringSet(
					PermissionSwitchTenant,
				),
			},
			{
				name:     "tenant-2",
				tenantID: "id-tenant-2",
				username: fedAccount.Username,
				wantPermissions: utils.NewStringSet(
					PermissionSwitchTenant,
					PermissionAccessAllTenants,
				),
			},
			{
				name:     "tenant-3",
				tenantID: "id-tenant-3",
				username: fedAccount.Username,
				wantPermissions: utils.NewStringSet(
					PermissionSwitchTenant,
					PermissionAccessAllTenants,
					PermissionViewOperatorLoginAsCustomer,
				),
			},
		}
		SubTestOauth2SwitchTenant(di, tests, fedAccount)(ctx, t, g)
	}
}

// SubTestOAuth2SwitchTenantNoFinalizer will expect permissions values consistent with the
// account store having no per-tenant finalizer. This means that switching tenants should
// yield no change in permissions
func SubTestOAuth2SwitchTenantNoFinalizer(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fedAccount := di.Mocking.FederatedUsers.MapValues()["fed3"]
		tests := []SwitchTenantTestStruct{
			{
				name:     "tenant-1",
				tenantID: "id-tenant-1",
				username: fedAccount.Username,
				wantPermissions: utils.NewStringSet(
					PermissionSwitchTenant,
				),
			},
			{
				name:     "tenant-2",
				tenantID: "id-tenant-2",
				username: fedAccount.Username,
				wantPermissions: utils.NewStringSet(
					PermissionSwitchTenant,
				),
			},
			{
				name:     "tenant-3",
				tenantID: "id-tenant-3",
				username: fedAccount.Username,
				wantPermissions: utils.NewStringSet(
					PermissionSwitchTenant,
				),
			},
		}
		SubTestOauth2SwitchTenant(di, tests, fedAccount)(ctx, t, g)
	}
}

// SubTestOauth2SwitchTenant is used as an intermediate test. Subtests may define the testStruct
// to customize what values are supplied and expected
func SubTestOauth2SwitchTenant(
	di *intDI,
	testStruct []SwitchTenantTestStruct,
	fedAccount *sectest.MockedFederatedUserProperties,
) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		// authorize
		req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, authorizeReqOptions("test-client"))
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusFound), "response should have correct status code")
		assertAuthorizeResponse(t, g, resp.Response, false)

		// token
		code := extractAuthCode(resp.Response)
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code, TestClientID, ""), tokenReqOptions())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
		a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)

		// verify token
		auth, err := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
		g.Expect(err).To(BeNil(), "unable to read auth: %v", err)
		tenantDetails, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeFalse()) // because no tenant selection
		userDetails, ok := auth.Details().(security.AuthenticationDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(userDetails.Permissions()).To(Equal(utils.NewStringSet(PermissionSwitchTenant)))

		oldValues := struct {
			tenantID        string
			username        string
			wantPermissions utils.StringSet
		}{
			tenantID:        "",
			username:        fedAccount.Username,
			wantPermissions: utils.NewStringSet(PermissionSwitchTenant),
		}

		oldRefreshTokenValue := a.RefreshToken().Value()
		for _, tt := range testStruct {
			t.Run(fmt.Sprintf("With Tenant: %v", tt.name), func(t *testing.T) {
				g = gomega.NewGomegaWithT(t)
				//switch tenant
				req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token",
					switchTenantBody(a.Value(), tt.tenantID), tokenReqOptions(), withDefaultClientAuth(),
				)
				resp = webtest.MustExec(ctx, req)
				g.Expect(resp).ToNot(BeNil())
				g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK))
				a = assertTokenResponse(t, g, resp.Response, tt.username, true)
				auth, err = di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)

				// verify token
				tenantDetails, ok = auth.Details().(security.TenantDetails)
				g.Expect(ok).To(BeTrue())
				g.Expect(tenantDetails.TenantId()).To(Equal(tt.tenantID))
				g.Expect(err).To(BeNil(), "unable to read auth: %v", err)
				userDetails, ok = auth.Details().(security.AuthenticationDetails)
				g.Expect(ok).To(BeTrue())
				g.Expect(userDetails.Permissions()).To(Equal(tt.wantPermissions))

				// verify new refresh token
				refreshToken, err := di.TokenReader.ReadRefreshToken(ctx, a.RefreshToken().Value())
				g.Expect(err).To(BeNil(), "unable to read auth: %v", err)
				req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token",
					requestNewAccessToken(refreshToken.Value()), tokenReqOptions(), withDefaultClientAuth(),
				)
				resp = webtest.MustExec(ctx, req)
				g.Expect(resp).ToNot(BeNil())
				g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK))
				a = assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)
				auth, err = di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)

				tenantDetails, ok = auth.Details().(security.TenantDetails)
				g.Expect(ok).To(BeTrue())
				g.Expect(tenantDetails.TenantId()).To(Equal(tt.tenantID))
				g.Expect(err).To(BeNil(), "unable to read auth: %v", err)
				userDetails, ok = auth.Details().(security.AuthenticationDetails)
				g.Expect(ok).To(BeTrue())
				g.Expect(userDetails.Permissions()).To(Equal(tt.wantPermissions))

				// verify old refresh token
				oldRefreshToken, err := di.TokenReader.ReadRefreshToken(ctx, oldRefreshTokenValue)
				g.Expect(err).To(BeNil(), "unable to read auth: %v", err)
				req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token",
					requestNewAccessToken(oldRefreshToken.Value()), tokenReqOptions(), withDefaultClientAuth(),
				)
				resp = webtest.MustExec(ctx, req)
				g.Expect(resp).ToNot(BeNil())
				g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK))
				a = assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)
				auth, err = di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)

				tenantDetails, ok = auth.Details().(security.TenantDetails)
				if oldValues.tenantID != "" {
					g.Expect(ok).To(BeTrue())
					g.Expect(tenantDetails.TenantId()).To(Equal(oldValues.tenantID))
				} else {
					g.Expect(ok).To(BeFalse())
				}

				g.Expect(err).To(BeNil(), "unable to read auth: %v", err)
				userDetails, ok = auth.Details().(security.AuthenticationDetails)
				g.Expect(ok).To(BeTrue())
				g.Expect(userDetails.Permissions()).To(Equal(oldValues.wantPermissions))

				// update old token
				oldRefreshTokenValue = refreshToken.Value()
				oldValues = struct {
					tenantID        string
					username        string
					wantPermissions utils.StringSet
				}{
					tenantID:        tt.tenantID,
					username:        tt.username,
					wantPermissions: tt.wantPermissions,
				}

			})
		}
	}
}

/*************************
	Helpers
 *************************/

func withClientAuth(clientId, secret string) webtest.RequestOptions {
	v := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientId, secret)))
	return webtest.Headers("Authorization", fmt.Sprintf("Basic %s", v))
}

func withDefaultClientAuth() webtest.RequestOptions {
	return withClientAuth(TestClientID, TestClientSecret)
}

func withDefaultAuthCode() webtest.RequestOptions {
	return webtest.Queries(
		oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode,
		oauth2.ParameterClientId, TestClientID,
		oauth2.ParameterRedirectUri, TestOAuth2CallbackURL,
	)
}

func withDefaultSamlSSO() webtest.RequestOptions {
	return func(req *http.Request) {
		webtest.Queries(
			oauth2.ParameterGrantType, oauth2.GrantTypeSamlSSO,
			oauth2.ParameterClientId, TestClientID,
			oauth2.ParameterRedirectUri, TestOAuth2CallbackURL,
		)(req)
	}
}

func assertRedirectResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedUrl string) {
	g.Expect(resp.StatusCode).To(BeNumerically("~", 300, 399), "response should be redirect")
	g.Expect(resp.Header).To(HaveKey("Location"), "response should have redirect location")

	expected, e := url.Parse(expectedUrl)
	g.Expect(e).To(Succeed(), "expected URL should be valid")
	loc, e := url.Parse(resp.Header.Get("Location"))
	g.Expect(e).To(Succeed(), "response's redirect location should be valid")
	if expected.IsAbs() {
		g.Expect(loc.String()).To(HavePrefix(expectedUrl), "response's redirect location should have correct host, port and path")
	} else {
		g.Expect(loc.Path).To(Equal(expectedUrl), "response's redirect location should have correct path")
	}
}

func contextWithSamlAuth(ctx context.Context, fedAcctStore security.FederatedAccountStore, mock *sectest.MockedFederatedUserProperties) (context.Context, error) {
	assertion := mockAssertion(mock)
	acct, e := fedAcctStore.LoadAccountByExternalId(ctx,
		mock.ExtIdName,
		mock.ExtIdValue,
		mock.ExtIdpName,
		MockAutoCreateUserDetails{},
		assertion)
	if e != nil {
		return nil, e
	}

	return sectest.ContextWithSecurity(ctx, sectest.Authentication(
		&samltest.MockedSamlAssertionAuthentication{
			Account:       acct,
			DetailsMap:    map[string]interface{}{},
			SamlAssertion: assertion,
		}),
	), nil
}

func mockAssertion(mock *sectest.MockedFederatedUserProperties) *saml.Assertion {
	return samltest.MockAssertion(func(opt *samltest.AssertionOption) {
		opt.NameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:email"
		opt.NameID = mock.ExtIdValue
		opt.RequestID = uuid.New().String()
		opt.Issuer = "http://some-entity-id"
		opt.Recipient = "http://some-sp/sso"
		opt.Audience = "http://some-sp"
	})
}

type MockAutoCreateUserDetails struct{}

func (m MockAutoCreateUserDetails) IsEnabled() bool {
	return true
}

func (m MockAutoCreateUserDetails) GetEmailWhiteList() []string {
	return []string{}
}

func (m MockAutoCreateUserDetails) GetAttributeMapping() map[string]string {
	return map[string]string{
		"firstName": "FirstName",
		"lastName":  "LastName",
		"email":     "Email",
	}
}

func (m MockAutoCreateUserDetails) GetElevatedUserRoleNames() []string {
	return []string{}
}

func (m MockAutoCreateUserDetails) GetRegularUserRoleNames() []string {
	return []string{}
}

func authorizeReqOptions(clientId string) webtest.RequestOptions {
	return func(req *http.Request) {
		req.Host = testdata.IdpDomainExtSAML
		req.URL.Host = testdata.IdpDomainExtSAML
		values := url.Values{}
		values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode)
		values.Set(oauth2.ParameterResponseType, "code")
		values.Set(oauth2.ParameterClientId, clientId)
		values.Set(oauth2.ParameterRedirectUri, "http://localhost/test/callback")
		req.URL.RawQuery = values.Encode()
	}
}

func extractAuthCode(resp *http.Response) string {
	loc := resp.Header.Get("Location")
	locUrl, _ := url.Parse(loc)
	return locUrl.Query().Get("code")
}

func authCodeReqBody(code string, clientId string, tenantId string) io.Reader {
	values := url.Values{}
	values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode)
	values.Set(oauth2.ParameterClientId, clientId)
	values.Set(oauth2.ParameterClientSecret, TestClientSecret)
	values.Set(oauth2.ParameterRedirectUri, "http://localhost/test/callback")
	values.Set(oauth2.ParameterAuthCode, code)
	if tenantId != "" {
		values.Set(oauth2.ParameterTenantId, tenantId)
	}
	return strings.NewReader(values.Encode())
}

func clientCredentialReqBody(tenantId string) io.Reader {
	values := url.Values{}
	values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeClientCredentials)
	if tenantId != "" {
		values.Set(oauth2.ParameterTenantId, tenantId)
	}
	return strings.NewReader(values.Encode())
}

func passwordGrantReqBody(tenantId string, username string, password string) io.Reader {
	values := url.Values{}
	values.Set(oauth2.ParameterGrantType, oauth2.GrantTypePassword)
	values.Set(oauth2.ParameterUsername, username)
	values.Set(oauth2.ParameterPassword, password)
	if tenantId != "" {
		values.Set(oauth2.ParameterTenantId, tenantId)
	}
	return strings.NewReader(values.Encode())
}

func tokenReqOptions() webtest.RequestOptions {
	return func(req *http.Request) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
	}
}

func switchTenantBody(accessToken string, tenantId string) io.Reader {
	values := url.Values{}
	values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeSwitchTenant)
	values.Set(oauth2.ParameterAccessToken, accessToken)
	values.Set(oauth2.ParameterTenantId, tenantId)
	return strings.NewReader(values.Encode())
}

func requestNewAccessToken(refreshToken string) io.Reader {
	values := url.Values{}
	values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeRefresh)
	values.Set(oauth2.ParameterRefreshToken, refreshToken)
	return strings.NewReader(values.Encode())
}

func assertTokenResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedUsername string, expectRefreshToken bool) oauth2.AccessToken {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `token response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.access_token"), "token response should have access_token")
	g.Expect(body).To(HaveJsonPath("$.expires_in"), "token response should have expires_in")
	g.Expect(body).To(HaveJsonPath("$.scope"), "token response should have scope")
	g.Expect(body).To(HaveJsonPathWithValue("$.token_type", ContainElement("bearer")), "token response should have token_type")
	g.Expect(body).To(HaveJsonPathWithValue("$.username", expectedUsername), "token response should have correct username")

	if expectRefreshToken {
		g.Expect(body).To(HaveJsonPath("$.refresh_token"), "token response should have refresh_token")
	} else {
		g.Expect(body).NotTo(HaveJsonPath("$..refresh_token"), "token response should not have refresh_token")
	}

	accessToken := oauth2.NewDefaultAccessToken("")
	e = json.Unmarshal(body, accessToken)
	g.Expect(e).ToNot(HaveOccurred())
	return accessToken
}

func assertAuthorizeResponse(t *testing.T, g *gomega.WithT, resp *http.Response, expectErr bool) {
	g.Expect(resp.Header.Get("Set-Cookie")).To(Not(BeEmpty()), "authorize response should set cookie")
	expected, _ := url.Parse(ExpectedAuthorizeCallback)
	loc := resp.Header.Get("Location")
	locUrl, e := url.Parse(loc)
	g.Expect(e).To(Succeed(), "authorize redirect location should be a valid URL")
	g.Expect(locUrl.Scheme).To(Equal(expected.Scheme), "authorize redirect should have correct scheme")
	g.Expect(locUrl.Host).To(Equal(expected.Host), "authorize redirect should have correct host")
	g.Expect(locUrl.Path).To(Equal(expected.Path), "authorize redirect should have correct path")
	q := locUrl.Query()
	switch {
	case expectErr:
		g.Expect(q.Get("error")).To(Not(BeEmpty()), "authorize redirect queries for error should have error")
		g.Expect(q.Get("code")).To(BeEmpty(), "authorize redirect queries for error should not have code")
	default:
		g.Expect(q.Get("code")).To(Not(BeEmpty()), "authorize redirect queries should have code")
	}
}

// This interface is not exposed, we declared it in test to check the internal implementation
type ClientDetails interface {
	ClientId() string
	AssignedTenantIds() utils.StringSet
	Scopes() utils.StringSet
}

func assertClientCredentialAuth(_ *testing.T, g *gomega.WithT, auth oauth2.Authentication, expectedClientId string, expectedTenantId string, expectedTenants utils.StringSet, expectedScopes utils.StringSet) {
	g.Expect(auth.Principal()).To(Equal(expectedClientId))
	g.Expect(auth.Permissions()).To(HaveLen(0))
	g.Expect(auth.UserAuthentication()).To(BeNil())
	g.Expect(auth.State()).To(Equal(security.StateAuthenticated))
	g.Expect(auth.UserAuthentication()).To(BeNil())
	td, ok := auth.Details().(security.TenantDetails)
	if expectedTenantId != "" {
		g.Expect(ok).To(BeTrue())
		g.Expect(td.TenantId()).To(Equal(expectedTenantId))
	} else {
		g.Expect(ok).To(BeFalse())
	}
	cd, ok := auth.Details().(ClientDetails)
	g.Expect(ok).To(BeTrue())
	g.Expect(cd.ClientId()).To(Equal(expectedClientId))
	g.Expect(cd.Scopes().HasAll(expectedScopes.Values()...)).To(BeTrue(), fmt.Sprintf("expected scopes %s doesn't match actual scopes %s", expectedScopes, cd.Scopes()))
	g.Expect(expectedScopes.HasAll(cd.Scopes().Values()...)).To(BeTrue(), fmt.Sprintf("expected scopes %s doesn't match actual scopes %s", expectedScopes, cd.Scopes()))
	g.Expect(cd.AssignedTenantIds().HasAll(expectedTenants.Values()...)).To(BeTrue())
	g.Expect(expectedTenants.HasAll(cd.AssignedTenantIds().Values()...)).To(BeTrue())
}

// This interface is not exposed. We define it in test in order to check the internal implementation.
type TenantAccessDetails interface {
	EffectiveAssignedTenantIds() utils.StringSet
}

func assertUserAuth(_ *testing.T, g *gomega.WithT, auth oauth2.Authentication, expectedUserId string, expectedTenantId string, expectedAssignedTenants utils.StringSet, expectedProviderId string) {
	userDetail, ok := auth.Details().(security.UserDetails)
	g.Expect(ok).To(BeTrue())
	g.Expect(userDetail.UserId()).To(Equal(expectedUserId))

	tenantAccessDetails, ok := auth.Details().(TenantAccessDetails)
	g.Expect(tenantAccessDetails.EffectiveAssignedTenantIds().HasAll(expectedAssignedTenants.Values()...)).To(BeTrue(), fmt.Sprintf("expected tenants %s, actual tenants %s", expectedAssignedTenants, tenantAccessDetails.EffectiveAssignedTenantIds()))
	g.Expect(expectedAssignedTenants.HasAll(tenantAccessDetails.EffectiveAssignedTenantIds().Values()...)).To(BeTrue(), fmt.Sprintf("expected tenants %s, actual tenants %s", expectedAssignedTenants, tenantAccessDetails.EffectiveAssignedTenantIds()))

	tenantDetail, ok := auth.Details().(security.TenantDetails)
	if expectedTenantId != "" {
		g.Expect(ok).To(BeTrue())
		g.Expect(tenantDetail.TenantId()).To(Equal(expectedTenantId))
	} else {
		g.Expect(ok).To(BeFalse())
	}

	if expectedProviderId != "" {
		providerDetail, ok := auth.Details().(security.ProviderDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(providerDetail.ProviderId()).To(Equal(expectedProviderId))
	} else {
		g.Expect(ok).To(BeFalse())
	}
}
