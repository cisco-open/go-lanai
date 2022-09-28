package saml_auth

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	lanaisaml "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/testdata"
	samlutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"embed"
	"encoding/base64"
	"encoding/xml"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"net/url"
	"testing"
)

/*************************
	Setup
 *************************/

const (
	TestContextPath      = webtest.DefaultContextPath
	TestLogoutErrorURL   = "/error"
	TestLogoutSuccessURL = "/logout/success"
	TestRelayState       = "MjJkNjBhNWYtMzAzMS00NmZkLWE2NjktMjRlZTFjNTZiZDBj"


)

//go:embed testdata/template/*.tmpl
var TestHTMLTemplates embed.FS

var TestSP *saml.ServiceProvider

type TestLogoutSecConfigurer struct{}

func (c *TestLogoutSecConfigurer) Configure(ws security.WebSecurity) {
	ssoUrl, _ := url.Parse("/does/not/matter")
	ws.Route(matcher.AnyRoute()).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New()).
		With(request_cache.New()).
		With(logout.New().
			LogoutUrl(testdata.TestIdpSloPath).
			AddErrorHandler(redirect.NewRedirectWithURL(TestLogoutErrorURL)).
			AddSuccessHandler(redirect.NewRedirectWithURL(TestLogoutSuccessURL)),
		).
		With(NewLogout().
			Issuer(testdata.TestIssuer).
			EnableSLO(testdata.TestIdpSloPath).
			SsoCondition(matcher.AnyRequest()).
			SsoLocation(ssoUrl).
			MetadataPath("does/not/matter"),
		)
}

type LogoutTestOut struct {
	fx.Out
	SecConfigurer   security.Configurer
	TestSP          *saml.ServiceProvider
	SamlClientStore saml_auth_ctx.SamlClientStore
}

func LogoutTestSecurityConfigProvider(registrar security.Registrar, webReg *web.Registrar) LogoutTestOut {
	webReg.MustRegister(TestHTMLTemplates)
	cfg := TestLogoutSecConfigurer{}
	registrar.Register(&cfg)
	if TestSP == nil {
		TestSP = testdata.NewTestSP()
	}
	return LogoutTestOut{
		SecConfigurer: &cfg,
		TestSP: TestSP,
		SamlClientStore: testdata.NewTestSamlClientStore(TestSP),
	}
}

/*************************
	Test
 *************************/

type sloTestDI struct {
	fx.In
	Properties   lanaisaml.SamlProperties
	MockedSigner *saml.ServiceProvider
}

func TestWithMockedServer(t *testing.T) {
	di := &sloTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(Module, logout.Module, request_cache.Module, lanaisaml.Module, access.AccessControlModule, errorhandling.ErrorHandlingModule),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(LogoutTestSecurityConfigProvider),
		),
		test.GomegaSubTest(SubTestSLORedirectBinding(di), "TestSLORedirectBinding"),
		test.GomegaSubTest(SubTestSLOPostBinding(di), "TestSLOPostBinding"),
		test.GomegaSubTest(SubTestSLORequesterError(di), "TestSLORequesterError"),
		test.GomegaSubTest(SubTestSLOResponderError(di), "TestSLOResponderError"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestSLORedirectBinding(_ *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.WithMockedSecurity(ctx)
		req, e := newLogoutRequest(ctx, saml.HTTPRedirectBinding, TestSP)
		g.Expect(e).To(Succeed(), "creating redirect SAML request should succeed")
		resp = webtest.MustExec(ctx, req).Response
		assertLogoutSuccessResponse(t, g, resp)
	}
}

func SubTestSLOPostBinding(_ *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.WithMockedSecurity(ctx)
		req, e := newLogoutRequest(ctx, saml.HTTPPostBinding, TestSP)
		g.Expect(e).To(Succeed(), "creating post SAML request should succeed")
		resp = webtest.MustExec(ctx, req).Response
		assertLogoutSuccessResponse(t, g, resp)
	}
}

func SubTestSLORequesterError(_ *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.WithMockedSecurity(ctx)
		sp := *TestSP
		sp.EntityID = "http://unregistered/sp"
		req, e := newLogoutRequest(ctx, saml.HTTPPostBinding, &sp)

		g.Expect(e).To(Succeed(), "creating redirect SAML request should succeed")
		resp = webtest.MustExec(ctx, req).Response
		assertLogoutRequesterErrorResponse(t, g, resp)
	}
}

func SubTestSLOResponderError(_ *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.WithMockedSecurity(ctx)
		req, e := newLogoutRequest(ctx, saml.HTTPRedirectBinding, TestSP, func(httpReq *http.Request) {
			q := httpReq.URL.Query()
			q.Set("Signature", "YXNkZmFzZGZzZGZzYWRm")
			httpReq.URL.RawQuery = q.Encode()
		})

		g.Expect(e).To(Succeed(), "creating redirect SAML request should succeed")
		resp = webtest.MustExec(ctx, req).Response
		assertLogoutResponderErrorResponse(t, g, resp)
	}
}

/*************************
	Helpers
 *************************/

type logoutSamlOptions func(samlReq *saml.LogoutRequest)
type logoutHttpOptions func(httpReq *http.Request)
func newLogoutRequest(ctx context.Context, binding string, sp *saml.ServiceProvider, opts...any) (*http.Request, error) {
	sr, e := samlutils.NewFixedLogoutRequest(sp, testdata.TestIdpSloURL.String(), "any_name_id")
	if e != nil {
		return nil, e
	}
	for _, v := range opts {
		switch fn := v.(type) {
		case func(samlReq *saml.LogoutRequest):
			fn(&sr.LogoutRequest)
		case logoutSamlOptions:
			fn(&sr.LogoutRequest)
		}
	}

	var req *http.Request
	switch binding {
	case saml.HTTPRedirectBinding:
		sloUrl, e := sr.Redirect(TestRelayState, sp)
		if e != nil {
			return nil, e
		}
		req = webtest.NewRequest(ctx, http.MethodGet, sloUrl.String(), nil)
	default:
		doc := etree.NewDocument()
		doc.SetRoot(sr.Element())
		srBuf, e := doc.WriteToBytes()
		if e != nil {
			return nil, e
		}
		encoded := base64.StdEncoding.EncodeToString(srBuf)
		values := url.Values{
			"SAMLRequest": []string{encoded},
			"RelayState": []string{TestRelayState},
		}
		req = webtest.NewRequest(ctx, http.MethodPost, testdata.TestIdpSloURL.String(), bytes.NewReader([]byte(values.Encode())))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for _, v := range opts {
		switch fn := v.(type) {
		case func(httpReq *http.Request):
			fn(req)
		case logoutHttpOptions:
			fn(req)
		}
	}
	return req, nil
}

func assertLogoutSuccessResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	sloResp := assertLogoutResponse(t, g, resp)
	g.Expect(sloResp.Status.StatusCode).To(Not(BeNil()), "SAML response should have status code")
	g.Expect(sloResp.Status.StatusCode.Value).To(Equal(saml.StatusSuccess), "SAML response should have success status code")
}

func assertLogoutResponderErrorResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	sloResp := assertLogoutResponse(t, g, resp)
	g.Expect(sloResp.Status.StatusCode).To(Not(BeNil()), "SAML response should have status code")
	g.Expect(sloResp.Status.StatusCode.Value).To(Equal(saml.StatusAuthnFailed), "SAML response should have success status code")
	g.Expect(sloResp.Status.StatusMessage).To(Not(BeNil()), "SAML response should have status message")
	g.Expect(sloResp.Status.StatusMessage.Value).To(Not(BeEmpty()), "SAML response should have non-empty status message")
}

func assertLogoutRequesterErrorResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	g.Expect(resp.StatusCode).To(Not(Equal(http.StatusOK)), "response should not be 200")
	g.Expect(resp.Header.Get("Content-Type")).To(HavePrefix("text/html"), "response should be HTML")
	doc := etree.NewDocument()
	_, e := doc.ReadFrom(resp.Body)
	g.Expect(e).To(Succeed(), "response body should be a valid HTML")

	el := doc.FindElement("//p[@id='error-msg']")
	g.Expect(el).To(Not(BeNil()), "response should have error message tag")
	g.Expect(el.Text()).To(Not(BeEmpty()), "response should have non-empty error message")
}

func assertLogoutResponse(t *testing.T, g *gomega.WithT, resp *http.Response) *saml.LogoutResponse {
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "response should be 200")
	values := extractHTMLFormData(t, g, resp)
	g.Expect(values).To(HaveKey(samlutils.HttpParamSAMLResponse), "response should submit %s", samlutils.HttpParamSAMLResponse)
	g.Expect(values).To(HaveKey(samlutils.HttpParamRelayState), "response should submit %s", samlutils.HttpParamRelayState)

	encoded := values[samlutils.HttpParamSAMLResponse]
	decoded, e := base64.StdEncoding.DecodeString(encoded)
	g.Expect(e).To(Succeed(), "%s should be valid base64", samlutils.HttpParamSAMLResponse)
	var sloResp saml.LogoutResponse
	e = xml.Unmarshal(decoded, &sloResp)
	g.Expect(e).To(Succeed(), "%s should be valid XML of LogoutResponse", samlutils.HttpParamSAMLResponse)

	//g.Expect(values[samlutils.HttpParamRelayState]).To(Equal(TestRelayState), "RelayState should match")
	return &sloResp
}

func extractHTMLFormData(_ *testing.T, g *gomega.WithT, resp *http.Response) map[string]string {
	g.Expect(resp.Header.Get("Content-Type")).To(HavePrefix("text/html"), "response should be HTML")
	doc := etree.NewDocument()
	_, e := doc.ReadFrom(resp.Body)
	g.Expect(e).To(Succeed(), "response body should be a valid HTML")

	values := map[string]string{}
	elems := doc.FindElements("//input")
	for _, el := range elems {
		name := el.SelectAttrValue("name", "unknown")
		value := el.SelectAttrValue("value", "")
		values[name] = value
	}
	return values
}