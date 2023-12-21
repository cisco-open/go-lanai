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
	"bytes"
	"crypto/x509"
	samlutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/utils"
	"encoding/xml"
	"fmt"
	"github.com/crewjam/saml"
	xrv "github.com/mattermost/xml-roundtrip-validator"
	"net/http"
	"strconv"
)

func UnmarshalRequest(req *saml.IdpAuthnRequest) error {
	if err := xrv.Validate(bytes.NewReader(req.RequestBuffer)); err != nil {
		return NewSamlRequesterError("authentication request is not valid xml", err)
	}

	if err := xml.Unmarshal(req.RequestBuffer, &req.Request); err != nil {
		return NewSamlInternalError("error unmarshal authentication request xml", err)
	}

	return nil
}

// ValidateAuthnRequest This method is similar to the method in saml.IdpAuthnRequest,
// Because the original implementation doesn't support signature check and destination check,
// we reimplement it here to add support for them
func ValidateAuthnRequest(req *saml.IdpAuthnRequest, spDetails SamlSpDetails, spMetadata *saml.EntityDescriptor) error {
	if !spDetails.SkipAuthRequestSignatureVerification {
		if err := verifySignature(req); err != nil {
			return NewSamlRequesterError("request signature cannot be verified", err)
		}
	}

	if req.Request.Destination != "" &&  req.Request.Destination != req.IDP.SSOURL.String() {
		return NewSamlResponderError(fmt.Sprintf("expected destination to be %q, not %q", req.IDP.SSOURL.String(), req.Request.Destination))
	}

	if req.Request.IssueInstant.Add(saml.MaxIssueDelay).Before(req.Now) {
		return NewSamlResponderError(fmt.Sprintf("request expired at %s",req.Request.IssueInstant.Add(saml.MaxIssueDelay)))
	}
	if req.Request.Version != "2.0" {
		return NewSamlRequestVersionMismatch("expected saml version 2.0")
	}

	return nil
}

func verifySignature(req *saml.IdpAuthnRequest) error {
	binding := saml.HTTPPostBinding
	if req.HTTPRequest.Method == http.MethodGet {
		binding = saml.HTTPRedirectBinding
	}
	cert, err := getServiceProviderCert(req,"signing")
	if err != nil {
		return NewSamlRequesterError("request signature cannot be verified, because metadata does not include certificate", err)
	}
	return samlutils.VerifySignature(func(sc *samlutils.SignatureContext) {
		sc.Binding = binding
		sc.XMLData = req.RequestBuffer
		sc.Certs = []*x509.Certificate{cert}
		sc.Request = req.HTTPRequest

	})
}

func DetermineACSEndpoint(req *saml.IdpAuthnRequest) error {
	//get by index
	if req.Request.AssertionConsumerServiceIndex != "" {
		for _, spAssertionConsumerService := range req.SPSSODescriptor.AssertionConsumerServices {
			if strconv.Itoa(spAssertionConsumerService.Index) == req.Request.AssertionConsumerServiceIndex {
				v := spAssertionConsumerService
				req.ACSEndpoint = &v
				return nil
			}
		}
	}

	//get by location
	if req.Request.AssertionConsumerServiceURL != "" {
		for _, spAssertionConsumerService := range req.SPSSODescriptor.AssertionConsumerServices {
			if spAssertionConsumerService.Location == req.Request.AssertionConsumerServiceURL {
				v := spAssertionConsumerService
				req.ACSEndpoint = &v
				return nil
			}
		}
	}

	// Some service providers, like the Microsoft Azure AD service provider, issue
	// assertion requests that don't specify an ACS url at all.
	if req.Request.AssertionConsumerServiceURL == "" && req.Request.AssertionConsumerServiceIndex == "" {
		// find a default ACS binding in the metadata that we can use
		for _, spAssertionConsumerService := range req.SPSSODescriptor.AssertionConsumerServices {
			if spAssertionConsumerService.IsDefault != nil && *spAssertionConsumerService.IsDefault {
				switch spAssertionConsumerService.Binding {
				case saml.HTTPPostBinding, saml.HTTPRedirectBinding:
					v := spAssertionConsumerService
					req.ACSEndpoint = &v
					return nil
				}
			}
		}

		// if we can't find a default, use *any* ACS binding
		for _, spAssertionConsumerService := range req.SPSSODescriptor.AssertionConsumerServices {
			switch spAssertionConsumerService.Binding {
			case saml.HTTPPostBinding, saml.HTTPRedirectBinding:
				v := spAssertionConsumerService
				req.ACSEndpoint = &v
				return nil
			}
		}
	}

	return NewSamlRequesterError("assertion consumer service not found")
}