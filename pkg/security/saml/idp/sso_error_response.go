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
    "crypto/tls"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
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