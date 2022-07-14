package saml_auth

import (
	"crypto/x509"
	"encoding/base64"
	"github.com/crewjam/saml"
	"regexp"
)

func getServiceProviderCert(req *saml.IdpAuthnRequest, usage string) (*x509.Certificate, error) {
	certStr := ""
	for _, keyDescriptor := range req.SPSSODescriptor.KeyDescriptors {
		if keyDescriptor.Use == usage && len(keyDescriptor.KeyInfo.X509Data.X509Certificates) > 0 {
			certStr = keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data
			break
		}
	}

	// If there are no certs explicitly labeled for encryption, return the first
	// non-empty cert we find.
	if certStr == "" {
		for _, keyDescriptor := range req.SPSSODescriptor.KeyDescriptors {
			if keyDescriptor.Use == "" &&
				len(keyDescriptor.KeyInfo.X509Data.X509Certificates) > 0 &&
				keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data != "" {

				certStr = keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data
				break
			}
		}
	}

	if certStr == "" {
		return nil, NewSamlInternalError("certificate not found")
	}

	// cleanup whitespace and re-encode a PEM
	certStr = regexp.MustCompile(`\s+`).ReplaceAllString(certStr, "")
	certBytes, err := base64.StdEncoding.DecodeString(certStr)
	if err != nil {
		return nil, NewSamlInternalError("cannot decode certificate base64: %v", err)
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, NewSamlInternalError("cannot parse certificate: %v", err)
	}
	return cert, nil
}