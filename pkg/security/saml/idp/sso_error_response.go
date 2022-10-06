package samlidp

import (
	"crypto/tls"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"fmt"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
)

func MakeErrorResponse(req *saml.IdpAuthnRequest, code string, message string) error {
	response := &saml.Response{
		Destination:  req.ACSEndpoint.Location,
		ID:           fmt.Sprintf("id-%x", cryptoutils.RandomBytes(20)),
		InResponseTo: req.Request.ID,
		IssueInstant: req.Now,
		Version:      "2.0",
		Issuer: &saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  req.IDP.MetadataURL.String(),
		},
		Status: saml.Status{
			StatusCode: saml.StatusCode{
				Value: code,
			},
			StatusMessage: &saml.StatusMessage{
				Value: message,
			},
		},
	}

	responseEl := response.Element()
	// Sign the response element
	{
		keyPair := tls.Certificate{
			Certificate: [][]byte{req.IDP.Certificate.Raw},
			PrivateKey:  req.IDP.Key,
			Leaf:        req.IDP.Certificate,
		}
		for _, cert := range req.IDP.Intermediates {
			keyPair.Certificate = append(keyPair.Certificate, cert.Raw)
		}
		keyStore := dsig.TLSCertKeyStore(keyPair)

		signatureMethod := req.IDP.SignatureMethod
		if signatureMethod == "" {
			signatureMethod = dsig.RSASHA1SignatureMethod
		}

		signingContext := dsig.NewDefaultSigningContext(keyStore)
		signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(canonicalizerPrefixList)
		if err := signingContext.SetSignatureMethod(signatureMethod); err != nil {
			return err
		}

		signedResponseEl, err := signingContext.SignEnveloped(responseEl)
		if err != nil {
			return err
		}

		sigEl := signedResponseEl.ChildElements()[len(signedResponseEl.ChildElements())-1]
		response.Signature = sigEl
		responseEl = response.Element()
	}

	req.ResponseEl = responseEl
	return nil
}