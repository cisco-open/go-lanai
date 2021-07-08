package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"fmt"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetadataEndpoint(t *testing.T) {
	prop := saml.NewSamlProperties()
	prop.KeyFile = "testdata/saml_test.key"
	prop.CertificateFile = "testdata/saml_test.cert"

	serverProp := web.NewServerProperties()
	serverProp.ContextPath = "europa"

	c := newSamlAuthConfigurer(*prop, newTestIdpManager(), newTestFedAccountStore())
	feature := New()
	feature.Issuer(security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
		*opt =security.DefaultIssuerDetails{
		Protocol:    "http",
		Domain:      "vms.com",
		Port:        8080,
		ContextPath: serverProp.ContextPath,
		IncludePort: true,
	}}))
	ws := TestWebSecurity{}

	m := c.makeMiddleware(feature, ws)

	r := gin.Default()
	r.Use(m.RefreshMetadataHandler())
	r.GET(serverProp.ContextPath + feature.metadataPath, m.MetadataHandlerFunc())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/europa/saml/metadata", nil)

	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)
	g.Expect(w).To(MetadataMatcher{})
}

func TestAcsEndpoint(t *testing.T) {
	//TODO Test this when we have facility to create assertions
}

func TestSamlEntryPoint(t *testing.T) {
	prop := saml.NewSamlProperties()
	prop.KeyFile = "testdata/saml_test.key"
	prop.CertificateFile = "testdata/saml_test.cert"

	serverProp := web.NewServerProperties()
	serverProp.ContextPath = "europa"

	c := newSamlAuthConfigurer(*prop, newTestIdpManager(), newTestFedAccountStore())
	feature := New()
	feature.Issuer(security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
		*opt =security.DefaultIssuerDetails{
			Protocol:    "http",
			Domain:      "vms.com",
			Port:        8080,
			ContextPath: serverProp.ContextPath,
			IncludePort: true,
		}}))
	ws := TestWebSecurity{}

	m := c.makeMiddleware(feature, ws)
	refreshHandler := m.RefreshMetadataHandler()
	//trigger the refresh manually to simulate the middleware being called.
	refreshHandler(mockGinContext())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://saml.vms.com:8080/europa/v2/authorize", nil)
	m.Commence(context.TODO(), req, w, errors.New("not authenticated"))

	g := gomega.NewWithT(t)
	g.Expect(w).To(AuthRequestMatcher{})
}

type MetadataMatcher struct {

}

func (m MetadataMatcher) Match(actual interface{}) (success bool, err error) {
	w := actual.(*httptest.ResponseRecorder)
	descriptor, err := samlsp.ParseMetadata(w.Body.Bytes())

	if err != nil {
		return false, err
	}

	if descriptor.EntityID != "http://vms.com:8080/europa/saml/metadata" {
		return false, nil
	}

	if len(descriptor.SPSSODescriptors) != 1 {
		return false, nil
	}

	if len(descriptor.SPSSODescriptors[0].AssertionConsumerServices) != 2{
		return false, nil
	}

	if descriptor.SPSSODescriptors[0].AssertionConsumerServices[0].Location != "http://vms.com:8080/europa/saml/SSO" {
		return false, nil
	}

	if descriptor.SPSSODescriptors[0].AssertionConsumerServices[1].Location != "http://saml.vms.com:8080/europa/saml/SSO" {
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

type AuthRequestMatcher struct {}

func (a AuthRequestMatcher) Match(actual interface{}) (success bool, err error) {
	w := actual.(*httptest.ResponseRecorder)
	body := string(w.Body.Bytes())

	if !strings.Contains(body, "action=\"https://dev-940621.oktapreview.com/app/dev-940621_samlservicelocalgo_1/exkwj65c2kC1vwtYi0h7/sso/saml\"") {
		return false, nil
	}
	if !strings.Contains(body, "id=\"SAMLRequestForm\"") {
		return false, nil
	}
	if !strings.Contains(body, "method=\"post\"") {
		return false, nil
	}
	return true, nil
}

func (a AuthRequestMatcher) FailureMessage(actual interface{}) (message string) {
	w := actual.(*httptest.ResponseRecorder)
	body := string(w.Body.Bytes())
	return fmt.Sprintf("Expected html with form posting auth request. Actual: " + body)
}

func (a AuthRequestMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	w := actual.(*httptest.ResponseRecorder)
	body := string(w.Body.Bytes())
	return fmt.Sprintf("Expected html with form posting auth request. Actual: " + body)
}

type TestIdpProvider struct {
	domain string
	metadataLocation string
	externalIdpName string
	externalIdName string
	entityId string
	metadataRequireSignature bool
	metadataTrustCheck bool
	metadataTrustedKeys []string
}

func (i TestIdpProvider) ShouldMetadataRequireSignature() bool {
	return i.metadataRequireSignature
}

func (i TestIdpProvider) ShouldMetadataTrustCheck() bool {
	return i.metadataTrustCheck
}

func (i TestIdpProvider) GetMetadataTrustedKeys() []string {
	return i.metadataTrustedKeys
}

func (i TestIdpProvider) Domain() string {
	return i.domain
}

func (i TestIdpProvider) EntityId() string {
	return i.entityId
}

func (i TestIdpProvider) MetadataLocation() string {
	return i.metadataLocation
}

func (i TestIdpProvider) ExternalIdName() string {
	return i.externalIdName
}

func (i TestIdpProvider) ExternalIdpName() string {
	return i.externalIdpName
}

type TestIdpManager struct {
	idpDetails TestIdpProvider
}

func newTestIdpManager() *TestIdpManager {
	return &TestIdpManager{
		idpDetails: TestIdpProvider{
			domain:           "saml.vms.com",
			metadataLocation: "testdata/okta_metadata.xml",
			externalIdpName: "okta",
			externalIdName: "email",
			entityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
		},
	}
}

func (t *TestIdpManager) GetIdentityProvidersWithFlow(context.Context, idp.AuthenticationFlow) []idp.IdentityProvider {
	return []idp.IdentityProvider{t.idpDetails}
}

func (t TestIdpManager) GetIdentityProviderByEntityId(_ context.Context, entityId string) (idp.IdentityProvider, error) {
	if entityId == t.idpDetails.entityId {
		return t.idpDetails, nil
	}
	return nil, errors.New("not found")
}

func (t TestIdpManager) GetIdentityProviderByDomain(_ context.Context, domain string) (idp.IdentityProvider, error) {
	if domain == t.idpDetails.domain {
		return t.idpDetails, nil
	}
	return nil, errors.New("not found")
}

type TestFedAccountStore struct {
}

func newTestFedAccountStore() *TestFedAccountStore {
	return &TestFedAccountStore{}
}

//The externalIdName and value matches the test assertion
//The externalIdp matches that from the TestIdpManager
func (t *TestFedAccountStore) LoadAccountByExternalId(ctx context.Context, externalIdName string, externalIdValue string, externalIdpName string, _ security.AutoCreateUserDetails, _ interface{}) (security.Account, error) {
	if externalIdName == "email" && externalIdValue == "test@example.com" && externalIdpName == "okta" {
		return security.NewUsernamePasswordAccount(&security.AcctDetails{
			ID:              "test@example.com",
			Type:            security.AccountTypeFederated,
			Username:        "test"}), nil
	}
	return nil, nil
}

type TestWebSecurity struct {

}

func (t TestWebSecurity) Context() context.Context {
	return context.TODO()
}

func (t TestWebSecurity) AndCondition(mwcm web.RequestMatcher) security.WebSecurity {
	panic("implement me")
}

func (t TestWebSecurity) Route(matcher web.RouteMatcher) security.WebSecurity {
	panic("implement me")
}

func (t TestWebSecurity) Condition(mwcm web.RequestMatcher) security.WebSecurity {
	panic("implement me")
}

func (t TestWebSecurity) Add(i ...interface{}) security.WebSecurity {
	panic("implement me")
}

func (t TestWebSecurity) Remove(i ...interface{}) security.WebSecurity {
	panic("implement me")
}

func (t TestWebSecurity) With(f security.Feature) security.WebSecurity {
	panic("implement me")
}

func (t TestWebSecurity) Shared(key string) interface{} {
	return nil
}

func (t TestWebSecurity) AddShared(key string, value interface{}) error {
	panic("implement me")
}

func (t TestWebSecurity) Authenticator() security.Authenticator {
	panic("implement me")
}

func (t TestWebSecurity) Features() []security.Feature {
	panic("implement me")
}

func mockGinContext() *gin.Context {
	req, _ := http.NewRequest("GET", "/unit-test", strings.NewReader(""))
	gc := gin.Context{
		Request:  req,
		Writer:   nil,
		Params:   gin.Params{},
		Keys:     map[string]interface{}{},
		Errors:   nil,
		Accepted: []string{},
	}
	return &gc
}
