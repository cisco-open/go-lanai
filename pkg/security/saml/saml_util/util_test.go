package saml_util

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"io/ioutil"
	"os"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	trusts, err := cryptoutils.LoadCert("testdata/cert_for_metadata.crt")
	if err != nil {
		t.Errorf("error setting up test, test certificate not found")
	}

	metadataFile, err := os.Open("testdata/test_metadata.xml")
	if err != nil {
		t.Errorf("error setting up test, test metadata not found")
	}

	metadata, err := ioutil.ReadAll(metadataFile)
	if err != nil {
		t.Errorf("error setting up test, can't read test metadata")
	}

	err = VerifySignature(MetadataSignature(metadata, trusts...))

	if err != nil {
		t.Errorf("expected signature verification to pass, but it failed with error %v", err)
	}

	trusts, err = cryptoutils.LoadCert("testdata/unrelated_cert.crt")
	if err != nil {
		t.Errorf("error setting up test, test certificate not found")
	}

	err = VerifySignature(MetadataSignature(metadata, trusts...))

	if err == nil {
		t.Errorf("expected signature verification to fail, but no error was thrown")
	}
}

func TestResolveXMLMetadata(t *testing.T) {
	metadataSource := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<md:EntityDescriptor entityID=\"some_entity_id\" xmlns:md=\"urn:oasis:names:tc:SAML:2.0:metadata\">\n\t<md:IDPSSODescriptor WantAuthnRequestsSigned=\"false\" protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\">\n\t\t<md:KeyDescriptor use=\"signing\">\n\t\t\t<ds:KeyInfo xmlns:ds=\"http://www.w3.org/2000/09/xmldsig#\">\n\t\t\t\t<ds:X509Data>\n\t\t\t\t\t<ds:X509Certificate>some_certificate</ds:X509Certificate>\n\t\t\t\t</ds:X509Data>\n\t\t\t</ds:KeyInfo>\n\t\t</md:KeyDescriptor>\n\t\t<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>\n\t\t<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>\n\t\t<md:SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST\" Location=\"https://some_url\"/>\n\t\t<md:SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect\" Location=\"http://some_url\"/>\n\t</md:IDPSSODescriptor>\n</md:EntityDescriptor>"
	descriptor, data, err := ResolveMetadata(context.Background(), metadataSource, nil)

	if err != nil {
		t.Errorf("xml should be parsed without error")
	}

	if metadataSource != string(data) {
		t.Errorf("data should be the same as the xml content")
	}

	if descriptor.EntityID != "some_entity_id" ||
		len(descriptor.IDPSSODescriptors) != 1 ||
		len(descriptor.IDPSSODescriptors[0].SingleSignOnServices) != 2{
		t.Errorf("data was not parsed correctly")
	}
}