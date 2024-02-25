// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"errors"
	"fmt"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	targetSSOUrl = "http://vms.com:8080/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"
)

var knownSP = samltest.MustNewMockedSP(func(opt *samltest.SPMockOption) {
	opt.Properties.EntityID = "http://localhost:8000/saml/metadata"
	opt.Properties.PrivateKeySource = "testdata/saml_test_sp.key"
	opt.Properties.CertsSource = "testdata/saml_test_sp.cert"
	opt.Properties.ACSPath = "/saml/acs"
	opt.Properties.SLOPath = "/saml/slo"
})

var unknownSP = samltest.MustNewMockedSP(func(opt *samltest.SPMockOption) {
	opt.Properties.EntityID = "http://localhost:8000/saml/metadata"
	opt.Properties.PrivateKeySource = "testdata/saml_test_unknown_sp.key"
	opt.Properties.CertsSource = "testdata/saml_test_unknown_sp.cert"
	opt.Properties.ACSPath = "/saml/acs"
	opt.Properties.SLOPath = "/saml/slo"
})

func TestSPInitiatedSso(t *testing.T) {
	sp := *knownSP

	testClientStore := samltest.NewMockedClientStore(samltest.ClientsWithSPs(&sp))
	testAccountStore := sectest.NewMockedAccountStore(
		[]*sectest.MockedAccountProperties{},
	)
	g := gomega.NewWithT(t)
	r := setupServerForTest(testClientStore, testAccountStore)

	authnReq, e := sp.MakeAuthenticationRequest(targetSSOUrl, saml.HTTPPostBinding, saml.HTTPPostBinding)
	g.Expect(e).To(gomega.Succeed())
	req := httptest.NewRequest("POST", "/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", nil)
	samltest.RequestWithSAMLPostBinding(authnReq, "")(req)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	g.Expect(w.Code).To(gomega.BeEquivalentTo(http.StatusOK))

	var samlResp saml.Response
	_, err := samltest.ParseBinding(w.Result(), &samlResp)
	if err != nil {
		t.Errorf("error parsing saml response xml")
		return
	}
	samlResponseXml := samlResp.Element()

	status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Success']")
	g.Expect(status).ToNot(gomega.BeNil())
}

// In this test we use a different cert key pair so that the SP's actual cert and key do not match the ones that are
// in its metadata. This way the signature of the auth request won't match the expected signature based on the metadata
func TestSPInitiatedSsoAuthRequestWithBadSignature(t *testing.T) {
	sp := *knownSP

	testClientStore := samltest.NewMockedClientStore(samltest.ClientsWithSPs(&sp))
	testAccountStore := sectest.NewMockedAccountStore(
		[]*sectest.MockedAccountProperties{},
	)
	g := gomega.NewWithT(t)
	r := setupServerForTest(testClientStore, testAccountStore)
	unknownSp := *unknownSP

	authnReq, e := unknownSp.MakeAuthenticationRequest(targetSSOUrl, saml.HTTPPostBinding, saml.HTTPPostBinding)
	g.Expect(e).To(gomega.Succeed())
	req := httptest.NewRequest("POST", "/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer", nil)
	samltest.RequestWithSAMLPostBinding(authnReq, "")(req)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	g.Expect(w.Code).To(gomega.BeEquivalentTo(http.StatusOK))

	var samlResp saml.Response
	_, err := samltest.ParseBinding(w.Result(), &samlResp)
	if err != nil {
		t.Errorf("error parsing saml response xml")
		return
	}
	samlResponseXml := samlResp.Element()

	// StatusCode Responder tells the auth requester that there's a problem with the request
	status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Responder']")
	g.Expect(status).ToNot(gomega.BeNil())
}

func TestIDPInitiatedSso(t *testing.T) {
	sp := *knownSP

	testClientStore := samltest.NewMockedClientStore(samltest.ClientsWithSPs(&sp))
	testAccountStore := sectest.NewMockedAccountStore(
		[]*sectest.MockedAccountProperties{},
	)

	r := setupServerForTest(testClientStore, testAccountStore)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/europa/v2/authorize", nil)
	q := req.URL.Query()
	q.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	q.Add("idp_init", "true")
	q.Add("entity_id", sp.EntityID)
	req.URL.RawQuery = q.Encode()

	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)
	g.Expect(w.Code).To(gomega.BeEquivalentTo(http.StatusOK))

	var samlResp saml.Response
	_, err := samltest.ParseBinding(w.Result(), &samlResp)
	if err != nil {
		t.Errorf("error parsing saml response xml")
		return
	}
	samlResponseXml := samlResp.Element()
	status := samlResponseXml.FindElement("//samlp:StatusCode[@Value='urn:oasis:names:tc:SAML:2.0:status:Success']")
	g.Expect(status).ToNot(gomega.BeNil())
}

func TestMetadata(t *testing.T) {
	testClientStore := samltest.NewMockedClientStore(samltest.ClientsWithSPs(knownSP))
	testAccountStore := sectest.NewMockedAccountStore(
		[]*sectest.MockedAccountProperties{},
	)

	r := setupServerForTest(testClientStore, testAccountStore)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/europa/metadata", nil)

	r.ServeHTTP(w, req)

	g := gomega.NewWithT(t)

	g.Expect(w).To(MetadataMatcher{})
}

// FIXME: this test setup is not same as our normally "web" package initialized web server.
// 		  Should use webtest and sectest package to mimic exact configuration
func setupServerForTest(testClientStore samlctx.SamlClientStore, testAccountStore security.AccountStore) *gin.Engine {
	prop := samlctx.NewSamlProperties()
	prop.KeyFile = "testdata/saml_test.key"
	prop.CertificateFile = "testdata/saml_test.cert"

	serverProp := web.NewServerProperties()
	serverProp.ContextPath = "europa"
	c := newSamlAuthorizeEndpointConfigurer(*prop, testClientStore, testAccountStore, nil)

	f := New().
		SsoLocation(&url.URL{Path: "/v2/authorize", RawQuery: "grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"}).
		SsoCondition(matcher.RequestWithForm("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")).
		MetadataPath("/metadata").
		Issuer(security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
			*opt = security.DefaultIssuerDetails{
				Protocol:    "http",
				Domain:      "vms.com",
				Port:        8080,
				ContextPath: serverProp.ContextPath,
				IncludePort: true,
			}
		}))

	opts := c.getIdentityProviderConfiguration(f)
	metaMw := NewMetadataMiddleware(opts, c.samlClientStore)
	mw := NewSamlAuthorizeEndpointMiddleware(metaMw, c.accountStore, c.attributeGenerator)

	r := gin.Default()
	r.ContextWithFallback = true
	r.Use(web.GinContextMerger())
	r.GET(serverProp.ContextPath+f.metadataPath, mw.MetadataHandlerFunc())
	r.Use(samlErrorHandlerFunc())
	r.Use(sectest.NewMockAuthenticationMiddleware(sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption) {
		opt.Principal = "test_user"
		opt.State = security.StateAuthenticated
	})).AuthenticationHandlerFunc())
	r.Use(mw.RefreshMetadataHandler(f.ssoCondition))
	r.Use(mw.AuthorizeHandlerFunc(f.ssoCondition))
	r.POST(serverProp.ContextPath+f.ssoLocation.Path, security.NoopHandlerFunc())
	r.GET(serverProp.ContextPath+f.ssoLocation.Path, security.NoopHandlerFunc())

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

	if len(descriptor.IDPSSODescriptors[0].SingleSignOnServices) != 2 {
		return false, nil
	}

	if descriptor.IDPSSODescriptors[0].SingleSignOnServices[0].Binding != saml.HTTPPostBinding || descriptor.IDPSSODescriptors[0].SingleSignOnServices[0].Location != "http://vms.com:8080/europa/v2/authorize?grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer" {
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

		for _, e := range ctx.Errors {
			if errors.Is(e.Err, security.ErrorTypeSecurity) {
				samlErrorHandler.HandleError(ctx, ctx.Request, ctx.Writer, e)
				break
			}
		}
	}
}
