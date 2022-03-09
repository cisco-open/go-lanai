package saml_auth

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samlssotest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"encoding/base64"
	"encoding/xml"
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

func TestSPInitiatedSso(t *testing.T) {
	sp := samlssotest.NewSamlSp("http://localhost:8000", "testdata/saml_test_sp.cert", "testdata/saml_test_sp.key")
	metadata, _ := xml.MarshalIndent(sp.Metadata(), "", "  ")

	testClientStore := samlssotest.NewMockedSamlClientStore(
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             sp.EntityID,
				MetadataSource:                       string(metadata),
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
		})
	testAccountStore := sectest.NewMockedAccountStore()

	r := setupServerForTest(testClientStore, testAccountStore)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/europa/v2/authorize", bytes.NewBufferString(makeAuthnRequest(sp)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	q := req.URL.Query()
	q.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	req.URL.RawQuery = q.Encode()
	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)
	g.Expect(w.Code).To(gomega.BeEquivalentTo(http.StatusOK))

	samlResponseXml, err := samlssotest.ParseSamlResponse(w.Body)
	if err != nil {
		t.Errorf("error parsing saml response xml")
		return
	}

	status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Success']")
	g.Expect(status).ToNot(gomega.BeNil())
}

//In this test we use a different cert key pair so that the SP's actual cert and key do not match the ones that are
// in its metadata. This way the signature of the auth request won't match the expected signature based on the metadata
func TestSPInitiatedSsoAuthRequestWithBadSignature(t *testing.T) {
	sp := samlssotest.NewSamlSp("http://localhost:8000", "testdata/saml_test_sp.cert", "testdata/saml_test_sp.key")
	metadata, _ := xml.MarshalIndent(sp.Metadata(), "", "  ")

	testClientStore := samlssotest.NewMockedSamlClientStore(
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             sp.EntityID,
				MetadataSource:                       string(metadata),
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
		})
	testAccountStore := sectest.NewMockedAccountStore()

	unknownSp := samlssotest.NewSamlSp("http://localhost:8000", "testdata/saml_test_unknown_sp.cert", "testdata/saml_test_unknown_sp.key")

	r := setupServerForTest(testClientStore, testAccountStore)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/europa/v2/authorize", bytes.NewBufferString(makeAuthnRequest(unknownSp)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	q := req.URL.Query()
	q.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	req.URL.RawQuery = q.Encode()
	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)
	g.Expect(w.Code).To(gomega.BeEquivalentTo(http.StatusOK))

	samlResponseXml, err := samlssotest.ParseSamlResponse(w.Body)

	if err != nil {
		t.Errorf("error parsing saml response xml")
		return
	}

	// StatusCode Responder tells the auth requester that there's a problem with the request
	status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Responder']")
	g.Expect(status).ToNot(gomega.BeNil())
}

func TestIDPInitiatedSso(t *testing.T) {
	sp := samlssotest.NewSamlSp("http://localhost:8000", "testdata/saml_test_sp.cert", "testdata/saml_test_sp.key")
	metadata, _ := xml.MarshalIndent(sp.Metadata(), "", "  ")

	testClientStore := samlssotest.NewMockedSamlClientStore(
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             sp.EntityID,
				MetadataSource:                       string(metadata),
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
		})
	testAccountStore := sectest.NewMockedAccountStore()

	r := setupServerForTest(testClientStore, testAccountStore)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/europa/v2/authorize", nil)
	q := req.URL.Query()
	q.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	q.Add("idp_init", "true")
	q.Add("entity_id", sp.EntityID)
	req.URL.RawQuery = q.Encode()

	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)
	g.Expect(w.Code).To(gomega.BeEquivalentTo(http.StatusOK))

	samlResponseXml, err := samlssotest.ParseSamlResponse(w.Body)
	if err != nil {
		t.Errorf("error parsing saml response xml")
		return
	}

	status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Success']")
	g.Expect(status).ToNot(gomega.BeNil())
}

func TestMetadata(t *testing.T) {
	testClientStore := samlssotest.NewMockedSamlClientStore(
		DefaultSamlClient{
			SamlSpDetails: SamlSpDetails{
				EntityId:                             "http://localhost:8000/saml/metadata",
				MetadataSource:                       "testdata/saml_test_sp_metadata.xml",
				SkipAssertionEncryption:              false,
				SkipAuthRequestSignatureVerification: false,
			},
		})
	testAccountStore := sectest.NewMockedAccountStore()

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

func setupServerForTest(testClientStore saml_auth_ctx.SamlClientStore, testAccountStore security.AccountStore) *gin.Engine {
	prop := samlctx.NewSamlProperties()
	prop.KeyFile = "testdata/saml_test.key"
	prop.CertificateFile = "testdata/saml_test.cert"

	serverProp := web.NewServerProperties()
	serverProp.ContextPath = "europa"
	c := newSamlAuthorizeEndpointConfigurer(*prop, testClientStore, testAccountStore, nil)

	f := NewEndpoint().
		SsoLocation(&url.URL{Path: "/v2/authorize", RawQuery: "grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"}).
		SsoCondition(matcher.RequestWithParam("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")).
		MetadataPath("/metadata").
		Issuer(security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
		*opt =security.DefaultIssuerDetails{
			Protocol:    "http",
			Domain:      "vms.com",
			Port:        8080,
			ContextPath: serverProp.ContextPath,
			IncludePort: true,
		}}))

	opts := c.getIdentityProviderConfiguration(f)
	mw := NewSamlAuthorizeEndpointMiddleware(opts, c.samlClientStore, c.accountStore, c.attributeGenerator)

	r := gin.Default()
	r.GET(serverProp.ContextPath + f.metadataPath, mw.MetadataHandlerFunc())
	r.Use(samlErrorHandlerFunc())
	r.Use(sectest.NewMockAuthenticationMiddleware(sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption){
		opt.Principal = "test_user"
		opt.State = security.StateAuthenticated
	})).AuthenticationHandlerFunc())
	r.Use(mw.RefreshMetadataHandler(f.ssoCondition))
	r.Use(mw.AuthorizeHandlerFunc(f.ssoCondition))
	r.POST(serverProp.ContextPath + f.ssoLocation.Path, security.NoopHandlerFunc())
	r.GET(serverProp.ContextPath + f.ssoLocation.Path, security.NoopHandlerFunc())

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

	if descriptor.IDPSSODescriptors[0].SingleSignOnServices[0].Binding != saml.HTTPPostBinding || descriptor.IDPSSODescriptors[0].SingleSignOnServices[0].Location != "http://vms.com:8080/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"{
		return false, nil
	}

	if descriptor.IDPSSODescriptors[0].SingleSignOnServices[1].Binding != saml.HTTPRedirectBinding || descriptor.IDPSSODescriptors[0].SingleSignOnServices[1].Location != "http://vms.com:8080/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer" {
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

func samlErrorHandlerFunc() gin.HandlerFunc {
	samlErrorHandler := NewSamlErrorHandler()
	return func(ctx *gin.Context) {
		ctx.Next()

		for _,e := range ctx.Errors {
			if errors.Is(e.Err, security.ErrorTypeSecurity) {
				samlErrorHandler.HandleError(ctx, ctx.Request, ctx.Writer, e)
				break
			}
		}
	}
}