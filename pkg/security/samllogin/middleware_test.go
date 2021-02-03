package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"fmt"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetadataEndpoint(t *testing.T) {
	prop := NewServiceProviderProperties()
	prop.RootUrl = "http://vms.com:8080"
	prop.KeyFile = "testdata/saml_test.key"
	prop.CertificateFile = "testdata/saml_test.cert"

	serverProp := web.NewServerProperties()
	serverProp.ContextPath = "europa"

	c := newSamlAuthConfigurer(*prop, *serverProp, newTestIdpManager(), newTestFedAccountStore())
	feature := New()
	ws := TestWebSecurity{}

	m := c.makeMiddleware(feature, ws)

	r := gin.Default()
	r.GET(serverProp.ContextPath + feature.metadataPath, m.MetadataHandlerFunc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/europa/saml/metadata", nil)

	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)
	g.Expect(w).To(MetadataMatcher{})
}

func TestAcsEndpoint(t *testing.T) {
	//TODO
}

func TestSamlEntryPoint(t *testing.T) {
	//TODO
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
	//TODO: assert acs

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

type TestIdpManager struct {
	idpDetails IdentityProviderDetails
}

func newTestIdpManager() *TestIdpManager {
	return &TestIdpManager{
		idpDetails: IdentityProviderDetails{
			Domain:           "saml.example.com",
			MetadataLocation: "testdata/okta_metadata.xml",
			ExternalIdpName: "okta",
			ExternalIdName: "email",
			EntityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
		},
	}
}

func (t *TestIdpManager) GetAllIdentityProvider() []IdentityProviderDetails {
	return []IdentityProviderDetails{t.idpDetails}
}

func (t TestIdpManager) GetIdentityProviderByEntityId(entityId string) (IdentityProviderDetails, error) {
	if entityId == t.idpDetails.EntityId {
		return t.idpDetails, nil
	}
	return IdentityProviderDetails{}, errors.New("not found")
}

type TestFedAccountStore struct {
}

func newTestFedAccountStore() *TestFedAccountStore {
	return &TestFedAccountStore{}
}

//The externalIdName and value matches the test assertion
//The externalIdp matches that from the TestIdpManager
func (t *TestFedAccountStore) LoadAccountByExternalId(externalIdName string, externalIdValue string, externalIdpName string) (security.Account, error) {
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
