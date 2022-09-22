package saml_auth

import (
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_util"
	"encoding/base64"
	"github.com/crewjam/saml"
	"net/http"
	"regexp"
	"time"
)

type SamlLogoutRequest struct {
	Request         *saml.LogoutRequest
	RequestBuffer   []byte
	SPMeta          *saml.EntityDescriptor // the requester
	SPSSODescriptor *saml.SPSSODescriptor
	Callback        *saml.Endpoint
	IDP             *saml.IdentityProvider
	Response        *saml.LogoutResponse
}

func (r SamlLogoutRequest) Validate() error {
	now := time.Now()
	if r.Request.Destination != "" &&  r.Request.Destination != r.IDP.LogoutURL.String() {
		return ErrorSamlSloResponder.WithMessage("expected destination to be %q, not %q", r.IDP.LogoutURL.String(), r.Request.Destination)
	}

	if r.Request.IssueInstant.Add(saml.MaxIssueDelay).Before(now) {
		return ErrorSamlSloResponder.WithMessage("request expired at %s", r.Request.IssueInstant.Add(saml.MaxIssueDelay))
	}
	if r.Request.Version != "2.0" {
		return NewSamlRequestVersionMismatch("expected saml version 2.0")
	}
	return nil
}

func (r SamlLogoutRequest) VerifySignature() error {
	// TODO we might need to support redirect binding
	data := r.RequestBuffer
	cert, e := r.serviceProviderCert("signing")
	if e != nil {
		return ErrorSamlSloResponder.WithMessage("logout request signature cannot be verified, because metadata does not include certificate")
	}
	return saml_util.VerifySignature(data, cert)
}

func (r SamlLogoutRequest) WriteResponse(rw http.ResponseWriter) error {
	if r.Response == nil {
		return ErrorSamlSloRequester.WithMessage("logout response is not available")
	}

	// the only supported binding is the HTTP-POST binding, so don't need to apply Redirect fix
	switch r.Callback.Binding {
	case saml.HTTPPostBinding:
		data := r.Response.Post("")
		if e := saml_util.WritePostBindingForm(data, rw); e != nil {
			return ErrorSamlSloRequester.WithMessage("unable to write response: %v",  e)
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
