package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin/testdata"
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

	c := newSamlAuthConfigurer(newSamlConfigurer(*prop, testdata.NewTestIdpManager()), testdata.NewTestFedAccountStore())
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
	tests := []struct {
		name string
		samlProperties saml.SamlProperties
	}{
		{
			name: "NameID Format not configured",
			samlProperties: saml.SamlProperties{
				KeyFile:         "testdata/saml_test.key",
				CertificateFile: "testdata/saml_test.cert"},
		},
		{
			name: "NameID Format set to unspecified.",
			samlProperties: saml.SamlProperties{
			KeyFile:         "testdata/saml_test.key",
			CertificateFile: "testdata/saml_test.cert",
			NameIDFormat:    "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"},
		},
		{
			name: "NameID Format set to emailAddress.",
			samlProperties: saml.SamlProperties{
			KeyFile:         "testdata/saml_test.key",
			CertificateFile: "testdata/saml_test.cert",
			NameIDFormat:    "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverProp := web.NewServerProperties()
			serverProp.ContextPath = "europa"

			c := newSamlAuthConfigurer(newSamlConfigurer(tt.samlProperties, testdata.NewTestIdpManager()), testdata.NewTestFedAccountStore())
			feature := New()
			feature.Issuer(testdata.TestIssuer)
			ws := TestWebSecurity{}

			m := c.makeMiddleware(feature, ws)
			refreshHandler := m.RefreshMetadataHandler()
			//trigger the refresh manually to simulate the middleware being called.
			refreshHandler(mockGinContext())

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "http://saml.vms.com:8080/europa/v2/authorize", nil)
			m.Commence(context.TODO(), req, w, errors.New("not authenticated"))

			g := gomega.NewWithT(t)
			//g.Expect(w).To(NewPostAuthRequestMatcher(tt.samlProperties))
			g.Expect(w).To(NewRedirectAuthRequestMatcher(tt.samlProperties))
		})
	}
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

	if len(descriptor.SPSSODescriptors[0].AssertionConsumerServices) != 1{
		return false, nil
	}

	if descriptor.SPSSODescriptors[0].AssertionConsumerServices[0].Location != "http://saml.vms.com:8080/europa/saml/SSO" {
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
