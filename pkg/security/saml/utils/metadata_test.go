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

package samlutils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"fmt"
	"github.com/crewjam/saml"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"testing"
)

const validXML = `<?xml version="1.0" encoding="UTF-8"?>\n<md:EntityDescriptor entityID="some_entity_id" xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata">\n\t<md:IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">\n\t\t<md:KeyDescriptor use="signing">\n\t\t\t<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">\n\t\t\t\t<ds:X509Data>\n\t\t\t\t\t<ds:X509Certificate>some_certificate</ds:X509Certificate>\n\t\t\t\t</ds:X509Data>\n\t\t\t</ds:KeyInfo>\n\t\t</md:KeyDescriptor>\n\t\t<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>\n\t\t<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>\n\t\t<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://some_url"/>\n\t\t<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="http://some_url"/>\n\t</md:IDPSSODescriptor>\n</md:EntityDescriptor>`
const invalidXML = `<?xml version="1.0" encoding="UTF-8"?><md:EntityDescriptor entityID="some_entity_id" xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata">\n\t<md:IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">\n\t\t<md:KeyDescriptor use="signing">\n\t\t\t<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">\n\t\t\t\t<ds:X509Data>\n\t\t\t\t\t<ds:X509Certificate>some_certificate</ds:X509Certificate>\n\t\t\t\t</ds:X509Data>\n\t\t\t</ds:KeyInfo>\n\t\t</md:KeyDescriptor>\n\t\t<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>\n\t\t<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>\n\t\t<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://some_url"/>\n\t\t<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="http://some_url"/>\n\t</md:IDPSSODescriptor>\n`

/********************
	Test
 ********************/

func TestResolveMetadata(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestXMLDocumentMode(), "TestXMLDocumentMode"),
		test.GomegaSubTest(SubTestFileLocationMode(), "TestFileLocationMode"),
		test.GomegaSubTest(SubTestHttpMode(), "TestHttpMode"),
	)
}

/********************
	Sub Tests
 ********************/

func SubTestXMLDocumentMode() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var entity *saml.EntityDescriptor
		var data []byte
		var e error
		entity, data, e = ResolveMetadata(ctx, validXML)
		g.Expect(e).To(Succeed(), "ResolveMetadata should not fail")
		g.Expect(data).To(BeEquivalentTo(validXML), "loaded data should be identical to input")
		assertMetadata(t, g, entity)

		entity, data, e = ResolveMetadata(ctx, invalidXML)
		g.Expect(e).To(HaveOccurred(), "ResolveMetadata should fail")
	}
}

func SubTestFileLocationMode() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var entity *saml.EntityDescriptor
		var data []byte
		var e error
		entity, data, e = ResolveMetadata(ctx, "testdata/dummy_idp_metadata.xml")
		g.Expect(e).To(Succeed(), "ResolveMetadata should not fail")
		g.Expect(data).To(Not(BeEmpty()), "loaded data should not be empty")
		assertMetadata(t, g, entity)

		entity, data, e = ResolveMetadata(ctx, "testdata/not_exist.xml")
		g.Expect(e).To(HaveOccurred(), "ResolveMetadata should fail")
	}
}

func SubTestHttpMode() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		svr := httptest.NewServer(http.HandlerFunc(serveHTTP))
		defer svr.Close()
		addr := svr.Listener.Addr()

		var entity *saml.EntityDescriptor
		var data []byte
		var e error
		entity, data, e = ResolveMetadata(ctx, fmt.Sprintf("http://%v/metadata/valid", addr))
		g.Expect(e).To(Succeed(), "ResolveMetadata should not fail")
		g.Expect(data).To(Not(BeEmpty()), "loaded data should not be empty")
		assertMetadata(t, g, entity)

		entity, data, e = ResolveMetadata(ctx, fmt.Sprintf("http://%v/metadata/valid", addr), WithHttpClient(http.DefaultClient))
		g.Expect(e).To(Succeed(), "ResolveMetadata should not fail")
		g.Expect(data).To(Not(BeEmpty()), "loaded data should not be empty")
		assertMetadata(t, g, entity)

		entity, data, e = ResolveMetadata(ctx, fmt.Sprintf("http://%v/metadata/invalid", addr))
		g.Expect(e).To(HaveOccurred(), "ResolveMetadata should fail")

		entity, data, e = ResolveMetadata(ctx, fmt.Sprintf("http://%v/metadata/notexist", addr))
		g.Expect(e).To(HaveOccurred(), "ResolveMetadata should fail")
	}
}

/********************
	Helpers
 ********************/

func assertMetadata(_ *testing.T, g *gomega.WithT, entity *saml.EntityDescriptor) {
	g.Expect(entity).To(Not(BeNil()), "metadata shouldn't be nil")
	g.Expect(entity.EntityID).To(Equal("some_entity_id"), "entity ID should be correct")
	g.Expect(entity.IDPSSODescriptors).To(HaveLen(1), "IDP descriptors should have correct size")
	g.Expect(entity.IDPSSODescriptors[0].SingleSignOnServices).To(HaveLen(2), "IDP SSO endpoints should have correct size")
}

func serveHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		serve404(rw)
		return
	}
	switch r.URL.Path {
	case "/metadata/valid":
		serveXML(rw, validXML)
	case "/metadata/invalid":
		serveXML(rw, invalidXML)
	default:
		serve404(rw)
	}
}

func serve404(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusNotFound)
	_, _ = rw.Write(nil)
}

func serveXML(rw http.ResponseWriter, xmlDoc string) {
	rw.Header().Set("Content-Type", "text/xml")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(xmlDoc))

}