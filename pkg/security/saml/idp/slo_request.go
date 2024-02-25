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
    "crypto/x509"
    "encoding/base64"
    samlutils "github.com/cisco-open/go-lanai/pkg/security/saml/utils"
    "github.com/crewjam/saml"
    "net/http"
    "regexp"
    "time"
)

type SamlLogoutRequest struct {
	HTTPRequest     *http.Request
	Binding         string
	Request         *saml.LogoutRequest
	RequestBuffer   []byte
	RelayState      string
	IDP             *saml.IdentityProvider
	SPMeta          *saml.EntityDescriptor // the requester
	SPSSODescriptor *saml.SPSSODescriptor
	Callback        *saml.Endpoint
	Response        *saml.LogoutResponse
}

func (r SamlLogoutRequest) Validate() error {
	now := time.Now()
	if r.Request.Destination != "" && r.Request.Destination != r.IDP.LogoutURL.String() {
		return ErrorSamlSloResponder.WithMessage("expected destination to be %q, not %q", r.IDP.LogoutURL.String(), r.Request.Destination)
	}

	if r.Request.IssueInstant.Add(saml.MaxIssueDelay).Before(now) {
		return ErrorSamlSloResponder.WithMessage("request expired at %s", r.Request.IssueInstant.Add(saml.MaxIssueDelay))
	}

	if r.Request.Version != "2.0" {
		return NewSamlRequestVersionMismatch("expected saml version 2.0")
	}

	if r.Request.NameID == nil || len(r.Request.NameID.Value) == 0 {
		return ErrorSamlSloResponder.WithMessage("request missing saml:NameID")
	}

	return nil
}

func (r SamlLogoutRequest) VerifySignature() error {
	cert, e := r.serviceProviderCert("signing")
	if e != nil {
		return ErrorSamlSloResponder.WithMessage("logout request signature cannot be verified, because metadata does not include certificate")
	}
	return samlutils.VerifySignature(func(sc *samlutils.SignatureContext) {
		sc.Binding = r.Binding
		sc.XMLData = r.RequestBuffer
		sc.Certs = []*x509.Certificate{cert}
		sc.Request = r.HTTPRequest
	})
}

func (r SamlLogoutRequest) WriteResponse(rw http.ResponseWriter) error {
	if r.Response == nil {
		return ErrorSamlSloRequester.WithMessage("logout response is not available")
	}

	// the only supported binding is the HTTP-POST binding, so don't need to apply Redirect fix
	switch r.Callback.Binding {
	case saml.HTTPPostBinding:
		data := r.Response.Post(r.RelayState)
		if e := samlutils.WritePostBindingHTML(data, rw); e != nil {
			return ErrorSamlSloRequester.WithMessage("unable to write response: %v", e)
		}
	default:
		return ErrorSamlSloRequester.WithMessage("%s: unsupported binding %s", r.SPMeta.EntityID, r.Callback.Binding)
	}
	return nil
}

func (r SamlLogoutRequest) serviceProviderCert(usage string) (*x509.Certificate, error) {
	certStr := ""
	for _, keyDescriptor := range r.SPSSODescriptor.KeyDescriptors {
		if keyDescriptor.Use == usage && len(keyDescriptor.KeyInfo.X509Data.X509Certificates) > 0 {
			certStr = keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data
			break
		}
	}

	// If there are no certs explicitly labeled for encryption, return the first
	// non-empty cert we find.
	if certStr == "" {
		for _, keyDescriptor := range r.SPSSODescriptor.KeyDescriptors {
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
