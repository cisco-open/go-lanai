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

package sp

import (
    "context"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/security"
    lanaisaml "github.com/cisco-open/go-lanai/pkg/security/saml"
    "github.com/cisco-open/go-lanai/pkg/security/saml/sp/testdata"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/test/samltest"
    "github.com/cisco-open/go-lanai/test/sectest"
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
			c := newSamlAuthConfigurer(newSamlConfigurer(tt.samlProperties, idpManager),
				sectest.NewMockedFederatedAccountStore(testdata.DefaultFedUserProperties...))
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
