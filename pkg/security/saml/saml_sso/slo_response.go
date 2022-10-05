package saml_auth

import (
	"crypto/tls"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
)

func MakeLogoutResponse(req *SamlLogoutRequest, code string, message string) (*saml.LogoutResponse, error) {
	now := saml.TimeNow()
	response := saml.LogoutResponse{
		Destination:  req.Callback.ResponseLocation,
		ID:           fmt.Sprintf("id-%x", cryptoutils.RandomBytes(20)),
		InResponseTo: req.Request.ID,
		IssueInstant: now,
		Version:      "2.0",
		Issuer: &saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  req.IDP.MetadataURL.String(),
		},
		Status: saml.Status{
			StatusCode: saml.StatusCode{
				Value: code,
			},
		},
	}

	if len(message) != 0 {
		response.Status.StatusMessage = &saml.StatusMessage{
			Value: message,
		}
	}

	if len(req.IDP.SignatureMethod) == 0 {
		req.IDP.SignatureMethod = dsig.RSASHA1SignatureMethod
	}

	if e := SignLogoutResponse(req.IDP, &response); e != nil {
		return nil, e
	}
	req.Response = &response
	return &response, nil
}

// SignLogoutResponse is similar to saml.ServiceProvider.SignLogoutResponse, but for IDP
func SignLogoutResponse(idp *saml.IdentityProvider, resp *saml.LogoutResponse) error {
	keyPair := tls.Certificate{
		Certificate: [][]byte{idp.Certificate.Raw},
		PrivateKey:  idp.Key,
		Leaf:        idp.Certificate,
	}
	// TODO: add intermediates for SP
	//for _, cert := range sp.Intermediates {
	//	keyPair.Certificate = append(keyPair.Certificate, cert.Raw)
	//}
	keyStore := dsig.TLSCertKeyStore(keyPair)

	if idp.SignatureMethod != dsig.RSASHA1SignatureMethod &&
		idp.SignatureMethod != dsig.RSASHA256SignatureMethod &&
		idp.SignatureMethod != dsig.RSASHA512SignatureMethod {
		return fmt.Errorf("invalid signing method %s", idp.SignatureMethod)
	}
	signatureMethod := idp.SignatureMethod
	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(canonicalizerPrefixList)
	if err := signingContext.SetSignatureMethod(signatureMethod); err != nil {
		return err
	}

	assertionEl := resp.Element()

	signedRequestEl, err := signingContext.SignEnveloped(assertionEl)
	if err != nil {
		return err
	}

	sigEl := signedRequestEl.Child[len(signedRequestEl.Child)-1]
	resp.Signature = sigEl.(*etree.Element)
	return nil
}