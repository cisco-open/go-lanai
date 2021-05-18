package saml_auth

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_util"
	"encoding/xml"
	"fmt"
	"github.com/crewjam/saml"
	xrv "github.com/mattermost/xml-roundtrip-validator"
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

//This method is similar to the method in saml.IdpAuthnRequest,
//Because the original implementation doesn't support signature check and destination check,
//we reimplement it here to add support for them
func ValidateAuthnRequest(req *saml.IdpAuthnRequest, spDetails SamlSpDetails, spMetadata *saml.EntityDescriptor) error {
	if !spDetails.SkipAuthRequestSignatureVerification {
		if err := verifySignature(req); err != nil {
			return err
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
	data := req.RequestBuffer;
	cert, err := getServiceProviderCert(req,"signing")
	if err != nil {
		return NewSamlRequesterError("request signature cannot be verified, because metadata does not include certificate", err)
	}
	return saml_util.VerifySignature(data, cert)
}

func DetermineACSEndpoint(req *saml.IdpAuthnRequest) error {
	//get by index
	if req.Request.AssertionConsumerServiceIndex != "" {
		for _, spAssertionConsumerService := range req.SPSSODescriptor.AssertionConsumerServices {
			if strconv.Itoa(spAssertionConsumerService.Index) == req.Request.AssertionConsumerServiceIndex {
				req.ACSEndpoint = &spAssertionConsumerService
				return nil
			}
		}
	}

	//get by location
	if req.Request.AssertionConsumerServiceURL != "" {
		for _, spAssertionConsumerService := range req.SPSSODescriptor.AssertionConsumerServices {
			if spAssertionConsumerService.Location == req.Request.AssertionConsumerServiceURL {
				req.ACSEndpoint = &spAssertionConsumerService
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
					req.ACSEndpoint = &spAssertionConsumerService
					return nil
				}
			}
		}

		// if we can't find a default, use *any* ACS binding
		for _, spAssertionConsumerService := range req.SPSSODescriptor.AssertionConsumerServices {
			switch spAssertionConsumerService.Binding {
			case saml.HTTPPostBinding, saml.HTTPRedirectBinding:
				req.ACSEndpoint = &spAssertionConsumerService
				return nil
			}
		}
	}

	return NewSamlRequesterError("assertion consumer service not found")
}