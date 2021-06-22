package saml_util

import (
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

	err = VerifySignature(metadata, trusts...)

	if err != nil {
		t.Errorf("expected signature verification to pass, but it failed with error %v", err)
	}

	trusts, err = cryptoutils.LoadCert("testdata/unrelated_cert.crt")
	if err != nil {
		t.Errorf("error setting up test, test certificate not found")
	}

	err = VerifySignature(metadata, trusts...)

	if err == nil {
		t.Errorf("expected signature verification to fail, but no error was thrown")
	}
}
