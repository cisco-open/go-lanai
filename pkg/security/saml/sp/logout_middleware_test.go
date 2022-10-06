package sp

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	lanaisaml "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/sp/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"errors"
	"fmt"
	"github.com/beevik/etree"
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
)

/*************************
	Setup
 *************************/

const (
	TestContextPath      = webtest.DefaultContextPath
	TestLogoutURL        = "/logout"
	TestSloURL           = "/saml/slo"
	TestLogoutSuccessURL = "/logout/success"
	TestLogoutErrorURL   = "/logout/error"
)

var (
	MockedLogoutAssertionOption = samltest.AssertionOption{
		Issuer:       "http://www.okta.com/exk668ha29xaI4in25d7",
		NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
		NameID:       "test-user",
		Recipient:    "http://saml.vms.com:8080/test/saml/SSO",
		Audience:     samltest.DefaultIssuer.Identifier() + "/saml/metadata",
		RequestID:    uuid.New().String(),
	}
	MockedLogoutResponseOption = samltest.LogoutResponseOption{
		Issuer:    "http://www.okta.com/exk668ha29xaI4in25d7",
		Recipient: "http://saml-alt.vms.com:8080/test/saml/slo",
		Audience:  samltest.DefaultIssuer.Identifier() + "/saml/metadata",
		RequestID: uuid.New().String(),
		Success:   true,
	}
)

type TestLogoutSecConfigurer struct{}

func (c *TestLogoutSecConfigurer) Configure(ws security.WebSecurity) {

	ws.Route(matcher.AnyRoute()).
		With(logout.New().
			LogoutUrl(TestLogoutURL).
			ErrorUrl(TestLogoutErrorURL).
		AddSuccessHandler(WarningsAwareSuccessHandler(TestLogoutSuccessURL)),
		).
		With(session.New()).
		With(csrf.New().IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(TestLogoutURL))).
		With(access.New().Request(matcher.AnyRequest()).Authenticated()).
		With(c.LogoutFeature())
}

func (c *TestLogoutSecConfigurer) LogoutFeature() *Feature {
	return NewLogout().
		Issuer(samltest.DefaultIssuer).
		ErrorPath(TestLogoutErrorURL)
}

type LogoutTestDI struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
	SecReg security.Registrar
	Props lanaisaml.SamlProperties
}

type LogoutTestOut struct {
	fx.Out
	SecConfigurer security.Configurer
	IdpManager    idp.IdentityProviderManager
	AccountStore  security.FederatedAccountStore
	MockedSigner  *saml.ServiceProvider
}

func LogoutTestSecurityConfigProvider(di LogoutTestDI) LogoutTestOut {
	cfg := TestLogoutSecConfigurer{}
	di.SecReg.Register(&cfg)
	idpManager := samltest.NewMockedIdpManager(samltest.IDPsWithPropertiesPrefix(di.AppCtx.Config(), "mocking.idp"))
	return LogoutTestOut{
		SecConfigurer: &cfg,
		IdpManager:    idpManager,
		AccountStore:  sectest.NewMockedFederatedAccountStore(),
		MockedSigner:  NewMockedSigner(di.Props, idpManager, cfg.LogoutFeature()),
	}
}

func NewMockedSigner(props lanaisaml.SamlProperties, idpManager idp.IdentityProviderManager, f *Feature) *saml.ServiceProvider {
	// make a copy
	c := newSamlConfigurer(props, idpManager)
	opts := c.getServiceProviderConfiguration(f)
	sp := c.sharedServiceProvider(opts)
	return &sp
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
		apptest.WithModules(Module, logout.Module, csrf.Module, request_cache.Module, lanaisaml.Module, access.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(LogoutTestSecurityConfigProvider),
		),
		test.GomegaSubTest(SubTestLogoutWithoutSAML(di), "TestLogoutWithoutSAML"),
		test.GomegaSubTest(SubTestLogoutWithSAML(di), "TestLogoutWithSAML"),
		test.GomegaSubTest(SubTestSLOSuccessCallback(di), "TestSLOSuccessCallback"),
		test.GomegaSubTest(SubTestSLOFailedCallback(di), "TestSLOFailedCallback"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLogoutWithoutSAML(_ *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication())
		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		assertLogoutSuccessResponse(t, g, resp)
		assertSLOState(ctx, t, g, 0)
	}
}

func SubTestLogoutWithSAML(di *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(mockSamlAuth()))

		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(NewRedirectSLORequestMatcher(di.Properties), "correct SLO request should be sent")
		assertSLOState(ctx, t, g, SLOInitiated)
	}
}

func SubTestSLOSuccessCallback(di *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		// send logout request for request_cache to store it
		ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(mockSamlAuth()))
		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		assertSLOState(ctx, t, g, SLOInitiated)

		// mimic callback
		samlResp := mockSamlSloResponse(true, di.MockedSigner)
		body := makeLogoutResponsePostBody(samlResp, uuid.New().String())

		req = webtest.NewRequest(ctx, http.MethodPost, TestSloURL, body, webtest.WithCookies(resp))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp = webtest.MustExec(ctx, req).Response
		assertLogoutReplayResponse(t, g, resp)
		assertSLOState(ctx, t, g, SLOCompletedFully)
	}
}

func SubTestSLOFailedCallback(di *sloTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		// send logout request for request_cache to store it
		ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(mockSamlAuth()))
		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		assertSLOState(ctx, t, g, SLOInitiated)

		// mimic callback
		samlResp := mockSamlSloResponse(false, di.MockedSigner)
		body := makeLogoutResponsePostBody(samlResp, uuid.New().String())

		req = webtest.NewRequest(ctx, http.MethodPost, TestSloURL, body, webtest.WithCookies(resp))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp = webtest.MustExec(ctx, req).Response
		assertLogoutReplayResponse(t, g, resp)
		assertSLOState(ctx, t, g, SLOFailed)
	}
}

/*************************
	Helpers
 *************************/

func mockSamlAuth() security.Authentication {
	assertion := samltest.MockAssertion(func(opt *samltest.AssertionOption) {
		*opt = MockedLogoutAssertionOption
	})
	return &samlAssertionAuthentication{
		Account:       nil,
		SamlAssertion: assertion,
		Perms:         security.Permissions{},
		DetailsMap:    map[string]interface{}{},
	}
}

func mockSamlSloResponse(success bool, signer *saml.ServiceProvider) *saml.LogoutResponse {
	resp := samltest.MockLogoutResponse(func(opt *samltest.LogoutResponseOption) {
		*opt = MockedLogoutResponseOption
		opt.Success = success
	})
	if signer != nil {
		_ = signer.SignLogoutResponse(resp)
	}
	return resp
}

func makeLogoutResponsePostBody(resp *saml.LogoutResponse, relayState string) io.Reader {
	html := resp.Post(relayState)
	doc := etree.NewDocument()
	if e := doc.ReadFromBytes(html); e != nil {
		return nil
	}

	if elem := doc.FindElement("//input[@name='SAMLResponse']"); elem != nil {
		decoded := elem.SelectAttrValue("value", "")
		body := fmt.Sprintf("SAMLResponse=%s", url.QueryEscape(decoded))
		return strings.NewReader(body)
	}
	return nil
}

func assertSLOState(ctx context.Context, _ *testing.T, g *gomega.WithT, expected SLOState) {
	actual := currentSLOState(ctx)
	g.Expect(actual.Is(expected)).To(BeTrue(), "SLO State should be %v, but got %v", expected, actual)
}

func assertRedirectResponse(_ *testing.T, g *gomega.WithT, resp *http.Response) *url.URL {
	g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
	g.Expect(resp.StatusCode).To(BeNumerically(">=", 300), "status code should be >= 300")
	g.Expect(resp.StatusCode).To(BeNumerically("<=", 399), "status code should be <= 399")
	loc, e := resp.Location()
	g.Expect(e).To(Succeed(), "Location header should be a valid URL")
	return loc
}

func assertLogoutReplayResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath+TestLogoutURL), "response should redirect to logout URL")
}

func assertLogoutSuccessResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath+TestLogoutSuccessURL), "response should redirect to logout URL")
}

/*************************
	Helper Types
 *************************/

type SLORequestMatcher struct {
	testdata.SamlRequestMatcher
}

func NewRedirectSLORequestMatcher(props lanaisaml.SamlProperties) *SLORequestMatcher {
	return &SLORequestMatcher{
		testdata.SamlRequestMatcher{
			SamlProperties: props,
			Binding:        saml.HTTPRedirectBinding,
			Subject:        "SLO request",
			ExpectedMsg:    "redirect with queries",
		},
	}
}

func (m SLORequestMatcher) Match(actual interface{}) (success bool, err error) {
	req, e := m.Extract(actual)
	if e != nil {
		return false, e
	}

	if req.Location != "https://dev-32506814.okta.com/app/dev-32506814_golanaisamllocal8765_1/exk668ha29xaI4in25d7/slo/saml" {
		return false, errors.New("incorrect request destination")
	}

	nameId := req.XMLDoc.FindElement("//saml:NameID")
	if nameId == nil {
		return false, errors.New("unable to find <saml:NameID>")
	}

	if e := m.compareAttr(nameId, "NameQualifier", MockedLogoutAssertionOption.Issuer); e != nil {
		return false, e
	}

	if e := m.compareAttr(nameId, "SPNameQualifier", MockedLogoutAssertionOption.Audience); e != nil {
		return false, e
	}

	if e := m.compareAttr(nameId, "Format", MockedLogoutAssertionOption.NameIDFormat); e != nil {
		return false, e
	}

	if e := m.compareValue(nameId, MockedLogoutAssertionOption.NameID); e != nil {
		return false, e
	}

	return true, nil
}

func (m SLORequestMatcher) compareAttr(elem *etree.Element, attrName, expected string) error {
	if actual := elem.SelectAttrValue(attrName, ""); actual != expected {
		return fmt.Errorf(`incorrect attribute <%s %s>, expected "%s", but got "%s"`, elem.Tag, attrName, expected, actual)
	}
	return nil
}

func (m SLORequestMatcher) compareValue(elem *etree.Element, expected string) error {
	if actual := elem.Text(); actual != expected {
		return fmt.Errorf(`incorrect <%s> value, expected "%s", but got "%s"`, elem.Tag, expected, actual)
	}
	return nil
}

type WarningsAwareSuccessHandler string

func (h WarningsAwareSuccessHandler) HandleAuthenticationSuccess(ctx context.Context, r *http.Request, rw http.ResponseWriter, _, _ security.Authentication) {
	redirectUrl := string(h)
	if contextPath, ok := ctx.Value(web.ContextKeyContextPath).(string); ok {
		redirectUrl = contextPath + redirectUrl
	}
	redirectUrl = h.appendWarnings(ctx, redirectUrl)
	http.Redirect(rw, r, redirectUrl, http.StatusFound)
	_, _ = rw.Write([]byte{})
}

func (h WarningsAwareSuccessHandler) appendWarnings(ctx context.Context, redirect string) string {
	warnings := logout.GetWarnings(ctx)
	if len(warnings) == 0 {
		return redirect
	}

	redirectUrl, e := url.Parse(redirect)
	if e != nil {
		return redirect
	}

	q := redirectUrl.Query()
	for _, w := range warnings {
		q.Add("warning", w.Error())
	}
	redirectUrl.RawQuery = q.Encode()
	return redirectUrl.String()
}
