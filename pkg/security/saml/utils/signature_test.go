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
    "crypto/rsa"
    "crypto/x509"
    "encoding/xml"
    "fmt"
    "github.com/beevik/etree"
    "github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
    "github.com/cisco-open/go-lanai/test"
    "github.com/crewjam/saml"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    dsig "github.com/russellhaering/goxmldsig"
    "io"
    "net/http"
    "net/url"
    "os"
    "testing"
)

/********************
	Setup
 ********************/

const TestAuthnRelayState = `MjJkNjBhNWYtMzAzMS00NmZkLWE2NjktMjRlZTFjNTZiZDBj`

var Certs []*x509.Certificate
var PrivateKey *rsa.PrivateKey
var MetadataCerts []*x509.Certificate
var UnrelatedCerts []*x509.Certificate
var TestSP *saml.ServiceProvider

func LoadCertificates(ctx context.Context, _ *testing.T) (context.Context, error) {
	var e error
	if Certs != nil {
		return ctx, nil
	}
	if Certs, e = cryptoutils.LoadCert("testdata/saml_test.cert"); e != nil {
		return nil, e
	}
	if MetadataCerts, e = cryptoutils.LoadCert("testdata/cert_for_metadata.crt"); e != nil {
		return nil, e
	}
	if UnrelatedCerts, e = cryptoutils.LoadCert("testdata/unrelated_cert.crt"); e != nil {
		return nil, e
	}
	if PrivateKey, e = cryptoutils.LoadPrivateKey("testdata/saml_test.key", ""); e != nil {
		return nil, e
	}
	TestSP = &saml.ServiceProvider{
		Key:                   PrivateKey,
		Certificate:           Certs[0],
		SignatureMethod:       dsig.RSASHA1SignatureMethod,
	}
	return ctx, nil
}

/********************
	Test
 ********************/

// TestGenerateSignatures this test is used to prepare signed requests.
// Unless test need to be regenerated, we don't need to run this test
func TestGenerateSignatures(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.SubTestSetup(LoadCertificates),
		test.GomegaSubTest(SubTestSignPostBinding[saml.AuthnRequest]("raw_authn_request.xml", "signed_authn_request.xml"), "TestSignAuthnPost"),
		test.GomegaSubTest(SubTestSignPostBinding[saml.LogoutRequest]("raw_logout_request.xml", "signed_logout_request.xml"), "TestSignLogoutPost"),
		test.GomegaSubTest(SubTestSignRedirectBinding[saml.AuthnRequest]("raw_authn_request.xml", "signed_authn_request.txt", TestAuthnRelayState), "TestSignLogoutRedirect"),
		test.GomegaSubTest(SubTestSignRedirectBinding[saml.LogoutRequest]("raw_logout_request.xml", "signed_logout_request.txt", ""), "TestSignAuthnRedirect"),
	)
}

func TestVerifySignature(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.SubTestSetup(LoadCertificates),
		test.GomegaSubTest(SubTestMetadata("test_metadata.xml"), "TestMetadataSignature"),
		test.GomegaSubTest(SubTestPostBinding("signed_authn_request.xml"), "TestAuthnRequestPostBinding"),
		test.GomegaSubTest(SubTestPostBinding("signed_logout_request.xml"), "TestLogoutRequestPostBinding"),
		test.GomegaSubTest(SubTestRedirectBinding("signed_authn_request.txt"), "TestAuthnRequestRedirectBinding"),
		test.GomegaSubTest(SubTestRedirectBinding("signed_logout_request.txt"), "TestLogoutRequestRedirectBinding"),
	)
}

/********************
	Sub Test
 ********************/

func SubTestMetadata(filepath string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		data := loadFile(t, g, filepath)
		var e error
		e = VerifySignature(MetadataSignature(data, MetadataCerts...))
		g.Expect(e).To(Succeed(), "signature should be valid")

		e = VerifySignature(MetadataSignature(data, UnrelatedCerts...))
		g.Expect(e).To(HaveOccurred(), "signature should not be valid")
	}
}

func SubTestPostBinding(filepath string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		data := loadFile(t, g, filepath)
		var e error
		e = VerifySignature(func(sc *SignatureContext) {
			*sc = SignatureContext{
				Binding: saml.HTTPPostBinding,
				Certs:   Certs,
				XMLData: data,
			}
		})
		g.Expect(e).To(Succeed(), "signature should be valid")

		e = VerifySignature(func(sc *SignatureContext) {
			*sc = SignatureContext{
				Binding: saml.HTTPPostBinding,
				Certs:   UnrelatedCerts,
				XMLData: data,
			}
		})
		g.Expect(e).To(HaveOccurred(), "signature should not be valid")
	}
}

func SubTestRedirectBinding(filepath string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		data := loadFile(t, g, filepath)
		link := fmt.Sprintf("http://localhost:9876/whatever/?%s", string(data))
		req, e := http.NewRequest(http.MethodGet, link, nil)
		g.Expect(e).To(Succeed(), "creating http request should succeed")

		e = VerifySignature(func(sc *SignatureContext) {
			*sc = SignatureContext{
				Binding: saml.HTTPRedirectBinding,
				Certs:   Certs,
				Request: req,
			}
		})
		g.Expect(e).To(Succeed(), "signature should be valid")

		e = VerifySignature(func(sc *SignatureContext) {
			*sc = SignatureContext{
				Binding: saml.HTTPRedirectBinding,
				Certs:   UnrelatedCerts,
				Request: req,
			}
		})
		g.Expect(e).To(HaveOccurred(), "signature should not be valid")

		e = VerifySignature(func(sc *SignatureContext) {
			*sc = SignatureContext{
				Binding: saml.HTTPPostBinding,
				Certs:   UnrelatedCerts,
				Request: req,
			}
		})
		g.Expect(e).To(HaveOccurred(), "signature should not be valid")
	}
}

func SubTestSignPostBinding[T saml.AuthnRequest|saml.LogoutRequest](inputFile, outputFile string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		data := loadFile(t, g, inputFile)
		var req T
		e := xml.Unmarshal(data, &req)
		g.Expect(e).To(Succeed(), "XML unmarshalling should succeed")

		var el *etree.Element
		var i interface{} = &req
		switch v := i.(type) {
		case *saml.AuthnRequest:
			e = TestSP.SignAuthnRequest(v)
			g.Expect(e).To(Succeed(), "Signing AuthnRequest should succeed")
			el = v.Element()
		case *saml.LogoutRequest:
			e = TestSP.SignLogoutRequest(v)
			g.Expect(e).To(Succeed(), "Signing LogoutRequest should succeed")
			el = v.Element()
		}

		doc := etree.NewDocument()
		doc.SetRoot(el)
		outData, e := doc.WriteToBytes()
		g.Expect(e).To(Succeed(), "marshalling should succeed")
		writeFile(t, g, outputFile, outData)
	}
}

func SubTestSignRedirectBinding[T saml.AuthnRequest|saml.LogoutRequest](inputFile, outputFile string, relayState string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		data := loadFile(t, g, inputFile)
		var req T
		e := xml.Unmarshal(data, &req)
		g.Expect(e).To(Succeed(), "XML unmarshalling should succeed")

		var rUrl *url.URL
		var i interface{} = &req
		switch v := i.(type) {
		case *saml.AuthnRequest:
			r := &FixedAuthnRequest{*v}
			rUrl, e = r.Redirect(relayState, TestSP)
		case *saml.LogoutRequest:
			r := &FixedLogoutRequest{*v}
			rUrl, e = r.Redirect(relayState, TestSP)
		}
		g.Expect(e).To(Succeed(), "singing request for redirect binding should succeed")

		rawQuery := rUrl.Query().Encode()
		writeFile(t, g, outputFile, []byte(rawQuery))
	}
}

/********************
	Sub Test
 ********************/

func loadFile(_ *testing.T, g *gomega.WithT, filepath string) []byte {
	file, e := os.Open("testdata/" + filepath)
	defer func() { _ = file.Close() }()
	g.Expect(e).To(Succeed(), "file should exists")
	data, e := io.ReadAll(file)
	g.Expect(e).To(Succeed(), "file should be readable")
	return data
}

func writeFile(_ *testing.T, g *gomega.WithT, filepath string, data []byte) {
	file, e := os.OpenFile("testdata/" + filepath, os.O_CREATE | os.O_TRUNC | os.O_RDWR, 0666)
	defer func() { _ = file.Close() }()
	g.Expect(e).To(Succeed(), "file should be opened")
	_, e = file.Write(data)
	g.Expect(e).To(Succeed(), "data should be written")
}

