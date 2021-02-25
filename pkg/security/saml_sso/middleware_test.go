package saml_auth

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSsoEndpoint(t *testing.T) {
	testClientStore := newTestSamlClientStore([]DefaultSamlClient{
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             "http://localhost:8000/saml/metadata",
				MetadataSource:                       "testdata/saml_test_sp_metadata.xml",
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
		},
	},)
	testAccountStore := newTestAccountStore()

	r := setupServerForTest(testClientStore, testAccountStore)

	rootURL, _ := url.Parse("http://localhost:8000")
	cert, _ := cryptoutils.LoadCert("testdata/saml_test_sp.cert")
	key, _ := cryptoutils.LoadPrivateKey("testdata/saml_test_sp.key", "")
	sp := samlsp.DefaultServiceProvider(samlsp.Options{
		URL:            *rootURL,
		Key:            key,
		Certificate:    cert,
		SignRequest: true,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/europa/v2/authorize", bytes.NewBufferString(makeAuthnRequest(sp)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.Query().Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")

	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)
	g.Expect(w.Code).To(gomega.BeEquivalentTo(http.StatusOK))
	//TODO: the response is html page with a form post. We need to be able to parse the html page if we want to further
	// examine the content of the post.
}

func TestMetadata(t *testing.T) {
	testClientStore := newTestSamlClientStore([]DefaultSamlClient{
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             "http://localhost:8000/saml/metadata",
				MetadataSource:                       "testdata/saml_test_sp_metadata.xml",
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
		},
	},)
	testAccountStore := newTestAccountStore()

	r := setupServerForTest(testClientStore, testAccountStore)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/europa/metadata", nil)

	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)

	g.Expect(w).To(MetadataMatcher{})
}


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

func setupServerForTest(testClientStore SamlClientStore, testAccountStore security.AccountStore) *gin.Engine {
	prop := security.NewSamlProperties()
	prop.RootUrl = "http://vms.com:8080"
	prop.KeyFile = "testdata/saml_test.key"
	prop.CertificateFile = "testdata/saml_test.cert"

	serverProp := web.NewServerProperties()
	serverProp.ContextPath = "europa"
	c := newSamlAuthorizeEndpointConfigurer(*prop, *serverProp, testClientStore, testAccountStore, nil)

	f := NewEndpoint().
		SsoLocation(&url.URL{Path: "/v2/authorize", RawQuery: "grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"}).
		SsoCondition(matcher.NotRequest(matcher.RequestWithParam("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer"))).
		MetadataPath("/metadata")

	opts := c.getIdentityProviderConfiguration(f)
	mw := NewSamlAuthorizeEndpointMiddleware(opts, c.samlClientStore, c.accountStore, c.attributeGenerator)

	r := gin.Default()
	r.GET(serverProp.ContextPath + f.metadataPath, mw.MetadataHandlerFunc)
	r.Use(MockAuthHandler)
	r.Use(mw.RefreshMetadataHandler(f.ssoCondition))
	r.Use(mw.AuthorizeHandlerFunc(f.ssoCondition))
	r.POST(serverProp.ContextPath + f.ssoLocation.Path, security.NoopHandlerFunc)

	return r
}

/*************
 * Matcher
 *************/
type MetadataMatcher struct {

}

func (m MetadataMatcher) Match(actual interface{}) (success bool, err error) {
	w := actual.(*httptest.ResponseRecorder)
	descriptor, err := samlsp.ParseMetadata(w.Body.Bytes())

	if err != nil {
		return false, err
	}

	if descriptor.EntityID != "http://vms.com:8080/europa" {
		return false, nil
	}

	if len(descriptor.IDPSSODescriptors) != 1 {
		return false, nil
	}

	if len(descriptor.IDPSSODescriptors[0].SingleSignOnServices) != 2{
		return false, nil
	}

	if descriptor.IDPSSODescriptors[0].SingleSignOnServices[0].Binding != saml.HTTPRedirectBinding || descriptor.IDPSSODescriptors[0].SingleSignOnServices[0].Location != "http://vms.com:8080/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"{
		return false, nil
	}

	if descriptor.IDPSSODescriptors[0].SingleSignOnServices[1].Binding != saml.HTTPPostBinding || descriptor.IDPSSODescriptors[0].SingleSignOnServices[1].Location != "http://vms.com:8080/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer" {
		return false, nil
	}

	if len(descriptor.IDPSSODescriptors[0].KeyDescriptors) != 2 {
		return false, nil
	}

	if descriptor.IDPSSODescriptors[0].KeyDescriptors[0].Use != "signing" {
		return false, nil
	}

	if descriptor.IDPSSODescriptors[0].KeyDescriptors[1].Use != "encryption" {
		return false, nil
	}

	return true, nil
}

func (m MetadataMatcher) FailureMessage(actual interface{}) (message string) {
	w := actual.(*httptest.ResponseRecorder)
	return fmt.Sprintf("metadata doesn't match expectation. actual meta is %s", string(w.Body.Bytes()))
}

func (m MetadataMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	w := actual.(*httptest.ResponseRecorder)
	return fmt.Sprintf("metadata doesn't match expectation. actual meta is %s", string(w.Body.Bytes()))
}

/*************************************
 * In memory Implementations for tests
 *************************************/
type TestSamlClientStore struct {
	details []DefaultSamlClient
}

func newTestSamlClientStore(d []DefaultSamlClient) *TestSamlClientStore {
	return &TestSamlClientStore{
		details: d,
	}
}

func (t *TestSamlClientStore) GetAllSamlClient() []SamlClient {
	var result []SamlClient
	for _, v := range t.details {
		result = append(result, v)
	}
	return result
}

func (t *TestSamlClientStore) GetSamlClientById(id string) (SamlClient, error) {
	for _, detail := range t.details {
		if detail.EntityId == id {
			return detail, nil
		}
	}
	return DefaultSamlClient{}, errors.New("not found")
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


func MockAuthHandler(ctx *gin.Context) {
	auth := NewUserAuthentication(func(opt *UserAuthOption){
		opt.Principal = "test_user"
		opt.State = security.StateAuthenticated
	})
	ctx.Set(security.ContextKeySecurity, auth)

}

/******************************
	UserAuthentication
******************************/
type UserAuthOptions func(opt *UserAuthOption)

type UserAuthOption struct {
	Principal   string
	Permissions map[string]interface{}
	State       security.AuthenticationState
	Details     map[string]interface{}
}

// userAuthentication implments security.Authentication.
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

func (a *userAuthentication) Permissions() map[string]interface{} {
	return a.PermissionMap
}

func (a *userAuthentication) State() security.AuthenticationState {
	return a.StateValue
}

func (a *userAuthentication) Details() interface{} {
	return a.DetailsMap
}

