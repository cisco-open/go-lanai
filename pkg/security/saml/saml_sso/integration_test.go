package saml_auth

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samlssotest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"embed"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)
import 	. "github.com/onsi/gomega"

//go:embed testdata/*
var whiteLabelContent embed.FS

const (
	TEST_SAML_SP_CERT = "testdata/saml_test_sp.cert"
	TEST_SAML_SP_KEY = "testdata/saml_test_sp.key"

	TEST_SAML_SP_1_URL = "http://localhost:8000"
	TEST_SAML_SP_2_URL = "http://localhost:8001"
)
var testRootTenantId = uuid.New()
var testTenantId1 = uuid.New()
var testTenantId2 = uuid.New()
var testTenantId3 = uuid.New()

var testUser1 = &sectest.MockedAccountProperties{
	Username: "testuser1",
	Tenants: []string{testTenantId1.String()},
}

var testUser2 = &sectest.MockedAccountProperties{
	Username: "testuser2",
	Tenants: []string{testTenantId1.String(), testTenantId2.String()},
}

var testUser3 = &sectest.MockedAccountProperties{
	Username: "testuser3",
	Tenants: []string{testTenantId3.String()},
}

var testSp1 = samlssotest.NewSamlSp(TEST_SAML_SP_1_URL, TEST_SAML_SP_CERT, TEST_SAML_SP_KEY)
var testSp2 = samlssotest.NewSamlSp(TEST_SAML_SP_2_URL, TEST_SAML_SP_CERT, TEST_SAML_SP_KEY)


type DIForTest struct {
	fx.In
	Register *web.Registrar
	Server *web.Engine
	MockAuthMw *sectest.MockAuthenticationMiddleware
}

func Test_Saml_Sso (t *testing.T) {
	di := &DIForTest{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(webinit.Module, security.Module, errorhandling.ErrorHandlingModule, tenancy.Module, samlctx.Module, Module),
		apptest.WithDI(di), // tell test framework to do dependencies injection
		apptest.WithTimeout(300*time.Second),
		apptest.WithProperties("server.context-path: /auth",
			"security.auth.saml.certificate-file: testdata/saml_test.cert",
			"security.auth.saml.key-file: testdata/saml_test.key"),
		apptest.WithFxOptions(
			fx.Provide(provideMockSamlClient, provideMockAccountStore, provideMockAuthMw, provideMockedTenancyAccessor)),
		apptest.WithFxOptions(fx.Invoke(configureAuthorizationServer)),
		test.GomegaSubTest(SubTestTenantRestrictionAny(di), "SubTestTenantRestrictionAny"),
		test.GomegaSubTest(SubTestTenantRestrictionAll(di), "SubTestTenantRestrictionAll"))
}

func SubTestTenantRestrictionAny(di *DIForTest) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		//testSp1 is configured to use tenant restriction "any"

		//test user 1 has access to one of the tenant in testSp1's tenant restriction, so expect success
		di.MockAuthMw.MockedAuthentication = sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption){
			opt.Principal = testUser1.Username
			opt.State = security.StateAuthenticated
		})

		port := di.Register.ServerPort()
		resp, _ := http.DefaultClient.Post(fmt.Sprintf("http://localhost:%d/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", port),
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(samlssotest.MakeAuthnRequest(testSp1, "http://localhost/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer")))

		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
		samlResponseXml, err := samlssotest.ParseSamlResponse(resp.Body)
		if err != nil {
			t.Errorf("cannot parse response due to error %v", err)
		}
		status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Success']")
		g.Expect(status).ToNot(BeNil())

		//test user 2 has access to all of the tenant in testSp1's tenant restriction, so expect success
		di.MockAuthMw.MockedAuthentication = sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption){
			opt.Principal = testUser2.Username
			opt.State = security.StateAuthenticated
		})
		resp, _ = http.DefaultClient.Post(fmt.Sprintf("http://localhost:%d/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", port),
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(samlssotest.MakeAuthnRequest(testSp1, "http://localhost/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer")))

		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
		samlResponseXml, err = samlssotest.ParseSamlResponse(resp.Body)
		if err != nil {
			t.Errorf("cannot parse response due to error %v", err)
		}
		status = samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Success']")
		g.Expect(status).ToNot(BeNil())

		//test user 3 has no access to any of the tenant in testSp1's tenant restriction, so expect failure
		di.MockAuthMw.MockedAuthentication = sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption){
			opt.Principal = testUser3.Username
			opt.State = security.StateAuthenticated
		})
		resp, _ = http.DefaultClient.Post(fmt.Sprintf("http://localhost:%d/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", port),
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(samlssotest.MakeAuthnRequest(testSp1, "http://localhost/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer")))

		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusInternalServerError))
		b, _ := ioutil.ReadAll(resp.Body)
		htmlContent := string(b)
		g.Expect(strings.Contains(htmlContent, "client is restricted to tenants which the authenticated user does not have access to")).To(BeTrue())
	}
}

func SubTestTenantRestrictionAll(di *DIForTest) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		//testSp2 is configured to use tenant restriction "any"

		//test user 1 has access to one of the tenant in testSp1's tenant restriction, so expect it to be rejected
		di.MockAuthMw.MockedAuthentication = sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption){
			opt.Principal = testUser1.Username
			opt.State = security.StateAuthenticated
		})

		port := di.Register.ServerPort()
		resp, _ := http.DefaultClient.Post(fmt.Sprintf("http://localhost:%d/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", port),
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(samlssotest.MakeAuthnRequest(testSp2, "http://localhost/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer")))

		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusInternalServerError))
		b, _ := ioutil.ReadAll(resp.Body)
		htmlContent := string(b)
		g.Expect(strings.Contains(htmlContent, "client is restricted to tenants which the authenticated user does not have access to")).To(BeTrue())

		//test user 2 has access to all of the tenant in testSp1's tenant restriction, so expect success
		di.MockAuthMw.MockedAuthentication = sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption){
			opt.Principal = testUser2.Username
			opt.State = security.StateAuthenticated
		})
		resp, _ = http.DefaultClient.Post(fmt.Sprintf("http://localhost:%d/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", port),
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(samlssotest.MakeAuthnRequest(testSp2, "http://localhost/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer")))

		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
		samlResponseXml, err := samlssotest.ParseSamlResponse(resp.Body)
		if err != nil {
			t.Errorf("cannot parse response due to error %v", err)
		}
		status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Success']")
		g.Expect(status).ToNot(BeNil())

		//test user 3 has no access to any of the tenant in testSp1's tenant restriction, so expect failure
		di.MockAuthMw.MockedAuthentication = sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption){
			opt.Principal = testUser3.Username
			opt.State = security.StateAuthenticated
		})
		resp, _ = http.DefaultClient.Post(fmt.Sprintf("http://localhost:%d/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", port),
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(samlssotest.MakeAuthnRequest(testSp2, "http://localhost/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer")))

		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusInternalServerError))
		b, _ = ioutil.ReadAll(resp.Body)
		htmlContent = string(b)
		g.Expect(strings.Contains(htmlContent, "client is restricted to tenants which the authenticated user does not have access to")).To(BeTrue())
	}
}

type configureDI struct {
	fx.In
	SecurityRegistrar    security.Registrar
	WebRegister *web.Registrar
	Server *web.Engine
	MockAuthMw *sectest.MockAuthenticationMiddleware
}

// This method provides the configuration to setup saml sso feature.
// This is equivalent to authserver.OAuth2AuthorizeModule except this module only sets up the saml sso related feature
// and not the oauth auth server related feature
//
// In addition, this module sets up a mock auth middleware to provide a mocked user session.
func configureAuthorizationServer(di configureDI) {
	//This mocks the session based authentication middleware
	di.Server.Use(di.MockAuthMw.AuthenticationHandlerFunc())

	//This is so that we can render an html error page.
	//In a real server configuration, this is included by the passwdidp module
	di.WebRegister.MustRegister(web.OrderedFS(whiteLabelContent, 0))

	di.SecurityRegistrar.Register(&authorizeEndpointConfigurer{})
}

type authorizeEndpointConfigurer struct {
}

func(c *authorizeEndpointConfigurer) Configure(ws security.WebSecurity) {
	location := &url.URL{Path: "/v2/authorize"}
	ws.Route(matcher.RouteWithPattern(location.Path)).
		With(NewEndpoint().
			Issuer(security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
				*opt = security.DefaultIssuerDetails{
					Protocol:    "http",
					Domain:      "localhost",
					ContextPath: "/auth",
					IncludePort: false,
				}})).
			SsoCondition(matcher.RequestWithParam(oauth2.ParameterGrantType, samlctx.GrantTypeSamlSSO)).
			SsoLocation(&url.URL{Path: "/v2/authorize", RawQuery: fmt.Sprintf("%s=%s", oauth2.ParameterGrantType, samlctx.GrantTypeSamlSSO)}).
			MetadataPath("/metadata"))
}

func provideMockSamlClient() saml_auth_ctx.SamlClientStore {
	sp1Metadata, _ := xml.MarshalIndent(testSp1.Metadata(), "", "  ")
	sp2Metadata, _ := xml.MarshalIndent(testSp2.Metadata(), "", "  ")

	return samlssotest.NewMockedSamlClientStore(
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             testSp1.EntityID,
				MetadataSource:                       string(sp1Metadata),
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
			TenantRestrictions: utils.NewStringSet(testTenantId1.String(), testTenantId2.String()),
			TenantRestrictionType: TenantRestrictionTypeAny,
		},
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             testSp2.EntityID,
				MetadataSource:                       string(sp2Metadata),
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
			TenantRestrictions: utils.NewStringSet(testTenantId1.String(), testTenantId2.String()),
			TenantRestrictionType: TenantRestrictionTypeAll,
		})
}

func provideMockAccountStore() security.AccountStore {
	return sectest.NewMockedAccountStore(testUser1, testUser2, testUser3)
}

func provideMockAuthMw() *sectest.MockAuthenticationMiddleware {
	return sectest.NewMockAuthenticationMiddleware(nil)
}

func provideMockedTenancyAccessor() tenancy.Accessor {
	tenancyRelationship := []mocks.TenancyRelation{
		{Parent: testRootTenantId, Child: testTenantId1},
		{Parent: testRootTenantId, Child: testTenantId2},
		{Parent: testRootTenantId, Child: testTenantId3},
	}
	return mocks.NewMockTenancyAccessor(tenancyRelationship, uuid.New())
}