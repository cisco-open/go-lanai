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
    "context"
    "crypto/tls"
    "fmt"
    "github.com/beevik/etree"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
    "github.com/crewjam/saml"
    "github.com/crewjam/saml/xmlenc"
    dsig "github.com/russellhaering/goxmldsig"
)

const canonicalizerPrefixList = ""

type AttributeGenerator func(account security.Account) []saml.Attribute

// MakeAssertion This is similar to the method in saml.IdpAuthnRequest
// but we have our own logic for generating attributes.
func MakeAssertion(ctx context.Context, req *saml.IdpAuthnRequest, authentication security.Authentication, generator AttributeGenerator) error {
	username, err := security.GetUsername(authentication)

	if err != nil {
		return NewSamlInternalError("can't get username from authentication", err)
	}

	attributes := []saml.Attribute{}

	attributes = append(attributes, saml.Attribute{
		Name:         "urn:mace:dir:attribute-def:uid",
		NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
		Values: []saml.AttributeValue{{
			Type:  "xs:string",
			Value: username,
		}},
	})

	attributes = append(attributes, saml.Attribute{
		Name:         "Username",
		NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
		Values: []saml.AttributeValue{{
			Type:  "xs:string",
			Value: username,
		}},
	})

	acct, ok := authentication.Principal().(security.Account)
	if generator != nil && ok {
		additionalAttributes := generator(acct)
		if len(additionalAttributes) > 0 {
			attributes = append(attributes, additionalAttributes...)
		}
	}

	// allow for some clock skew in the validity period using the
	// issuer's apparent clock.
	notBefore := req.Now.Add(-1 * saml.MaxClockSkew)
	notOnOrAfterAfter := req.Now.Add(saml.MaxIssueDelay)
	if notBefore.Before(req.Request.IssueInstant) {
		notBefore = req.Request.IssueInstant
		notOnOrAfterAfter = notBefore.Add(saml.MaxIssueDelay)
	}


	authCtxClassRef := "urn:oasis:names:tc:SAML:2.0:ac:classes:unspecified"
	details, ok := authentication.Details().(map[string]interface{})
	if ok {
		if authMethod, exist := details[security.DetailsKeyAuthMethod]; exist {
			switch authMethod {
			case security.AuthMethodPassword:
				authCtxClassRef = "urn:oasis:names:tc:SAML:2.0:ac:classes:Password"
			case security.AuthMethodExternalSaml:
				authCtxClassRef = "urn:oasis:names:tc:SAML:2.0:ac:classes:InternetProtocol"
			case security.AuthMethodExternalOpenID:
				authCtxClassRef = "urn:oasis:names:tc:SAML:2.0:ac:classes:InternetProtocol"
			}
		}
	}

	req.Assertion = &saml.Assertion{
		ID:           fmt.Sprintf("id-%x", cryptoutils.RandomBytes(20)),
		IssueInstant: saml.TimeNow(),
		Version:      "2.0",
		Issuer: saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  req.IDP.Metadata().EntityID,
		},
		Subject: &saml.Subject{
			NameID: &saml.NameID{
				Format:          "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
				Value:           username,
			},
			SubjectConfirmations: []saml.SubjectConfirmation{
				{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: &saml.SubjectConfirmationData{
						InResponseTo: req.Request.ID,
						NotOnOrAfter: req.Now.Add(saml.MaxIssueDelay),
						Recipient:    req.ACSEndpoint.Location,
					},
				},
			},
		},
		Conditions: &saml.Conditions{
			NotBefore:    notBefore,
			NotOnOrAfter: notOnOrAfterAfter,
			AudienceRestrictions: []saml.AudienceRestriction{
				{
					Audience: saml.Audience{Value: req.ServiceProviderMetadata.EntityID},
				},
			},
		},
		AuthnStatements: []saml.AuthnStatement{
			{
				AuthnInstant: security.DetermineAuthenticationTime(ctx, authentication),
				AuthnContext: saml.AuthnContext{
					AuthnContextClassRef: &saml.AuthnContextClassRef{
						Value: authCtxClassRef,
					},
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{
				Attributes: attributes,
			},
		},
	}
	return nil
}

// MakeAssertionEl This is similar to the implementation in saml.IdpAuthnRequest
// we re-implement it here because we need to optionally skip encryption
func MakeAssertionEl(req *saml.IdpAuthnRequest, skipEncryption bool) error {
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

	//This canonicalizer is used to canonicalize a subdocument in such a way that it is substantially independent of its XML context
	//because we want to sign the assertion payload indpendent of the response envelope,
	//so we don't want to canonicalize the assertion's element with the evenlope's name space.
	//we give an empty prefix list because we don't want any of the envelope's name space.
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(canonicalizerPrefixList)
	if err := signingContext.SetSignatureMethod(signatureMethod); err != nil {
		return NewSamlResponderError("unsupported signature method for signing assertion", err)
	}

	assertionEl := req.Assertion.Element()

	signedAssertionEl, err := signingContext.SignEnveloped(assertionEl)
	if err != nil {
		return NewSamlResponderError("error signing assertion", err)
	}

	sigEl := signedAssertionEl.Child[len(signedAssertionEl.Child)-1]
	req.Assertion.Signature = sigEl.(*etree.Element)
	signedAssertionEl = req.Assertion.Element()

	if skipEncryption {
		req.AssertionEl = signedAssertionEl
		return nil
	}

	certBuf, err := getServiceProviderCert(req, "encryption")
	if err != nil {
		return NewSamlRequesterError("requester doesn't provide encryption key in metadata")
	}

	var signedAssertionBuf []byte
	{
		doc := etree.NewDocument()
		doc.SetRoot(signedAssertionEl)
		signedAssertionBuf, err = doc.WriteToBytes()
		if err != nil {
			return err
		}
	}

	encryptor := xmlenc.OAEP()
	encryptor.BlockCipher = xmlenc.AES128CBC
	encryptor.DigestMethod = &xmlenc.SHA1
	encryptedDataEl, err := encryptor.Encrypt(certBuf, signedAssertionBuf, nil)
	if err != nil {
		return NewSamlResponderError("error signing assertion")
	}
	encryptedDataEl.CreateAttr("Type", "http://www.w3.org/2001/04/xmlenc#Element")

	encryptedAssertionEl := etree.NewElement("saml:EncryptedAssertion")
	encryptedAssertionEl.AddChild(encryptedDataEl)
	req.AssertionEl = encryptedAssertionEl

	return nil
}

