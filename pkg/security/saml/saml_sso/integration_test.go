package saml_auth_test

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"net/url"
	"testing"
	"time"
)
import 	. "github.com/onsi/gomega"

var logger = log.New("saml_sso_test")

//go:embed application.yml
var customConfigFS embed.FS

//go:embed testdata/*
var whiteLabelContent embed.FS

type DIForTest struct {
	fx.In
	Register *web.Registrar
	Server *web.Engine
}

//TODO: need to invoke something similar to ConfigureAuthorizationServer in security/config/authserver

func Test_Saml_Sso (t *testing.T) {
	di := &DIForTest{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(webinit.Module),
		apptest.WithModules(security.Module, saml_auth.Module, TestModule),
		apptest.WithDI(di), // tell test framework to do dependencies injection
		apptest.WithTimeout(300*time.Second),
		apptest.WithConfigFS(customConfigFS),
		apptest.WithFxOptions(
			fx.Provide(newTestSamlClientStore, newTestAccountStore)),
		test.GomegaSubTest(SubTestTenantRestriction(di), "SubTestTenantRestriction"))
}

func SubTestTenantRestriction(di *DIForTest) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		//rootURL, _ := url.Parse("http://localhost:8000")
		//cert, _ := cryptoutils.LoadCert("testdata/saml_test_sp.cert")
		//key, _ := cryptoutils.LoadPrivateKey("testdata/saml_test_sp.key", "")
		//sp := samlsp.DefaultServiceProvider(samlsp.Options{
		//	URL:            *rootURL,
		//	Key:            key,
		//	Certificate:    cert[0],
		//	SignRequest: true,
		//})
		//
		//w := httptest.NewRecorder()
		//req, _ := http.NewRequest("POST", "/auth/v2/authorize", bytes.NewBufferString(makeAuthnRequest(sp)))
		//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		//q := req.URL.Query()
		//q.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
		//req.URL.RawQuery = q.Encode()
		//di.Server.ServeHTTP(w, req)
		//
		//g.Expect(w.Code).To(BeEquivalentTo(http.StatusOK))

		port := di.Register.ServerPort()

		rootURL, _ := url.Parse(fmt.Sprintf("http://localhost:8000"))
		cert, _ := cryptoutils.LoadCert("testdata/saml_test_sp.cert")
		key, _ := cryptoutils.LoadPrivateKey("testdata/saml_test_sp.key", "")
		sp := samlsp.DefaultServiceProvider(samlsp.Options{
			URL:            *rootURL,
			Key:            key,
			Certificate:    cert[0],
			SignRequest: true,
		})

		resp, e := http.DefaultClient.Post(fmt.Sprintf("http://localhost:%d/auth/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", port),
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(makeAuthnRequest(sp)))

		if e == nil {
			logger.Infof("got response %s", resp.Status)
		}

		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
	}
}

//TODO: deduplicate
func MockAuthHandler(ctx *gin.Context) {
	auth := NewUserAuthentication(func(opt *UserAuthOption){
		opt.Principal = "test_user"
		opt.State = security.StateAuthenticated
	})
	ctx.Set(security.ContextKeySecurity, auth)

}

type UserAuthOptions func(opt *UserAuthOption)

type UserAuthOption struct {
	Principal   string
	Permissions map[string]interface{}
	State       security.AuthenticationState
	Details     map[string]interface{}
}

type userAuthentication struct {
	Subject       string
	PermissionMap map[string]interface{}
	StateValue    security.AuthenticationState
	DetailsMap    map[string]interface{}
}

func NewUserAuthentication(opts...UserAuthOptions) *userAuthentication {
	opt := UserAuthOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &userAuthentication{
		Subject:       opt.Principal,
		PermissionMap: opt.Permissions,
		StateValue:    opt.State,
		DetailsMap:    opt.Details,
	}
}

func (a *userAuthentication) Principal() interface{} {
	return a.Subject
}

func (a *userAuthentication) Permissions() security.Permissions {
	return a.PermissionMap
}

func (a *userAuthentication) State() security.AuthenticationState {
	return a.StateValue
}

func (a *userAuthentication) Details() interface{} {
	return a.DetailsMap
}

//TODO: de-duplicate
func makeAuthnRequest(sp saml.ServiceProvider) string {
	authnRequest, _ := sp.MakeAuthenticationRequest("http://vms.com:8080/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer")
	doc := etree.NewDocument()
	doc.SetRoot(authnRequest.Element())
	reqBuf, _ := doc.WriteToBytes()
	encodedReqBuf := base64.StdEncoding.EncodeToString(reqBuf)

	data := url.Values{}
	data.Set("SAMLRequest", encodedReqBuf)
	data.Add("RelayState", "my_relay_state")

	return data.Encode()
}

/*************************************
Test Configuration Module
 */
var TestModule = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(ConfigureAuthorizationServer),
	},
}

type initDI struct {
	fx.In
	SecurityRegistrar    security.Registrar
	WebRegister *web.Registrar
	Server *web.Engine
}

// ConfigureAuthorizationServer is the Configuration entry point
func ConfigureAuthorizationServer(di initDI) {
	di.Server.Use(MockAuthHandler)

	di.WebRegister.MustRegister(web.OrderedFS(whiteLabelContent, 0))

	di.SecurityRegistrar.Register(&AuthorizeEndpointConfigurer{di.WebRegister.ServerPort()})

}

type AuthorizeEndpointConfigurer struct {
	Port int
}

func(c *AuthorizeEndpointConfigurer) Configure(ws security.WebSecurity) {
	location := &url.URL{Path: "/v2/authorize"}
	ws.Route(matcher.RouteWithPattern(location.Path)).
		With(saml_auth.NewEndpoint().
			Issuer(security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
				*opt = security.DefaultIssuerDetails{
					Protocol:    "http",
					Domain:      "localhost",
					Port:        c.Port,
					ContextPath: "/auth",
					IncludePort: true,
				}})).
			SsoCondition(matcher.RequestWithParam(oauth2.ParameterGrantType, samlctx.GrantTypeSamlSSO)).
			SsoLocation(&url.URL{Path: "/v2/authorize", RawQuery: fmt.Sprintf("%s=%s", oauth2.ParameterGrantType, samlctx.GrantTypeSamlSSO)}).
			MetadataPath("/metadata"))
}

/*************************************
 * In memory Implementations for tests
 *************************************/
type TestSamlClientStore struct {
	details []saml_auth.DefaultSamlClient
}

func newTestSamlClientStore() saml_auth.SamlClientStore {
	return &TestSamlClientStore{
		details: []saml_auth.DefaultSamlClient{
			{
				SamlSpDetails: saml_auth.SamlSpDetails{
					EntityId:                             "http://localhost:8000/saml/metadata",
					MetadataSource:                       "testdata/saml_test_sp_metadata.xml",
					SkipAssertionEncryption:              false,
					SkipAuthRequestSignatureVerification: false,
				},
			},
		},
	}
}

func (t *TestSamlClientStore) GetAllSamlClient(_ context.Context) ([]saml_auth.SamlClient, error) {
	var result []saml_auth.SamlClient
	for _, v := range t.details {
		result = append(result, v)
	}
	return result, nil
}

func (t *TestSamlClientStore) GetSamlClientByEntityId(_ context.Context, id string) (saml_auth.SamlClient, error) {
	for _, detail := range t.details {
		if detail.EntityId == id {
			return detail, nil
		}
	}
	return saml_auth.DefaultSamlClient{}, errors.New("not found")
}

type TestAccountStore struct {

}

func newTestAccountStore() *TestAccountStore {
	return &TestAccountStore{}
}

func (t *TestAccountStore) LoadAccountById(ctx context.Context, id interface{}) (security.Account, error) {
	panic("implement me")
}

func (t *TestAccountStore) LoadAccountByUsername(ctx context.Context, username string) (security.Account, error) {
	panic("implement me")
}

func (t *TestAccountStore) LoadLockingRules(ctx context.Context, acct security.Account) (security.AccountLockingRule, error) {
	panic("implement me")
}

func (t *TestAccountStore) LoadPwdAgingRules(ctx context.Context, acct security.Account) (security.AccountPwdAgingRule, error) {
	panic("implement me")
}

func (t *TestAccountStore) Save(ctx context.Context, acct security.Account) error {
	panic("implement me")
}