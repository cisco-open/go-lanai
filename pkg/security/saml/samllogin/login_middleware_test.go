package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	lanaisaml "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"errors"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAcsEndpoint(t *testing.T) {
	//TODO Test this when we have facility to create assertions
}

func TestSamlEntryPoint(t *testing.T) {
	tests := []struct {
		name string
		samlProperties lanaisaml.SamlProperties
	}{
		{
			name: "NameID Format not configured",
			samlProperties: lanaisaml.SamlProperties{
				KeyFile:         "testdata/saml_test.key",
				CertificateFile: "testdata/saml_test.cert"},
		},
		{
			name: "NameID Format set to unspecified.",
			samlProperties: lanaisaml.SamlProperties{
			KeyFile:         "testdata/saml_test.key",
			CertificateFile: "testdata/saml_test.cert",
			NameIDFormat:    "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"},
		},
		{
			name: "NameID Format set to emailAddress.",
			samlProperties: lanaisaml.SamlProperties{
			KeyFile:         "testdata/saml_test.key",
			CertificateFile: "testdata/saml_test.cert",
			NameIDFormat:    "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverProp := web.NewServerProperties()
			serverProp.ContextPath = "europa"

			idpManager := samltest.NewMockedIdpManager(func(opt *samltest.IdpManagerMockOption) {
				opt.IDPList = testdata.DefaultIdpProviders
			})
			c := newSamlAuthConfigurer(newSamlConfigurer(tt.samlProperties, idpManager), testdata.NewTestFedAccountStore())
			feature := New()
			feature.Issuer(samltest.DefaultIssuer)
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

type AuthRequestMatcher struct {
	testdata.SamlRequestMatcher
}

func NewPostAuthRequestMatcher(props lanaisaml.SamlProperties) *AuthRequestMatcher {
	return &AuthRequestMatcher{
		testdata.SamlRequestMatcher{
			SamlProperties: props,
			Binding:        saml.HTTPPostBinding,
			Subject:        "auth request",
			ExpectedMsg:    "HTML with form posting",
		},
	}
}

func NewRedirectAuthRequestMatcher(props lanaisaml.SamlProperties) *AuthRequestMatcher {
	return &AuthRequestMatcher{
		testdata.SamlRequestMatcher{
			SamlProperties: props,
			Binding:        saml.HTTPRedirectBinding,
			Subject:        "auth request",
			ExpectedMsg:    "redirect with queries",
		},
	}
}

func (a AuthRequestMatcher) Match(actual interface{}) (success bool, err error) {
	req, e := a.Extract(actual)
	if e != nil {
		return false, e
	}

	if req.Location != "https://dev-940621.oktapreview.com/app/dev-940621_samlservicelocalgo_1/exkwj65c2kC1vwtYi0h7/sso/saml" {
		return false, errors.New("incorrect request destination")
	}

	nameIdPolicy := req.XMLDoc.FindElement("//samlp:NameIDPolicy")
	if a.SamlProperties.NameIDFormat == "" {
		if nameIdPolicy.SelectAttr("Format").Value != "urn:oasis:names:tc:SAML:2.0:nameid-format:transient" {
			return false, errors.New("NameIDPolicy format should be transient if it's not configured in our properties")
		}
	} else if a.SamlProperties.NameIDFormat == "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified" {
		if nameIdPolicy.SelectAttr("Format") != nil {
			return false, errors.New("NameIDPolicy should not have a format, if we configure it to be unspecified")
		}
	} else {
		if nameIdPolicy.SelectAttr("Format").Value != a.SamlProperties.NameIDFormat {
			return false, errors.New("NameIDPolicy format should match our configuration")
		}
	}
	return true, nil
}
