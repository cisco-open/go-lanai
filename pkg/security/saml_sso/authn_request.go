package saml_auth

import (
	"bytes"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	xrv "github.com/mattermost/xml-roundtrip-validator"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/russellhaering/goxmldsig/etreeutils"
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
	doc := etree.NewDocument()
	// this shouldn't occur because we have already parsed the request buffer into the auth request structure
	// so this is just a sanity check
	if err := doc.ReadFromBytes(req.RequestBuffer); err != nil {
		return NewSamlInternalError("error parsing request for signature verification", err)
	}

	el := doc.Root()

	sigEl, err := findChild(el, "http://www.w3.org/2000/09/xmldsig#", "Signature")

	if err != nil || sigEl == nil {
		return NewSamlRequesterError("auth request is not signed")
	}

	cert, err := getServiceProviderCert(req,"signing")

	if err != nil {
		return NewSamlRequesterError("request signature cannot be verified, because metadata does not include certificate", err)
	}

	certificateStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{cert},
	}

	validationContext := dsig.NewDefaultValidationContext(&certificateStore)
	validationContext.IdAttribute = "ID"
	if saml.Clock != nil {
		validationContext.Clock = saml.Clock
	}

	//if there's signature but keyInfo is not X509, then we remove the key info element, and just use the
	//default public key to verify.
	//if keyinfo is x509, it'll be verified that it's a trusted key before being used to verify the signature
	//See the logic in validationContext.ValidateAuthnRequest
	if el.FindElement("./Signature/KeyInfo/X509Data/X509Certificate") == nil {
		if keyInfo := sigEl.FindElement("KeyInfo"); keyInfo != nil {
			sigEl.RemoveChild(keyInfo)
		}
	}

	ctx, err := etreeutils.NSBuildParentContext(el)
	if err != nil {
		return NewSamlInternalError("error getting document context for signature check", err)
	}
	ctx, err = ctx.SubContext(el)
	if err != nil {
		return NewSamlInternalError("error getting document sub context for signature check", err)
	}
	//makes a copy of the element
	el, err = etreeutils.NSDetatch(ctx, el)
	if err != nil {
		return NewSamlInternalError("error getting document for signature check", err)
	}

	_, err = validationContext.Validate(el)

	if err!= nil {
		return NewSamlRequesterError("Invalid signature", err)
	}

	return nil
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

func findChild(parentEl *etree.Element, childNS string, childTag string) (*etree.Element, error) {
	for _, childEl := range parentEl.ChildElements() {
		if childEl.Tag != childTag {
			continue
		}

		ctx, err := etreeutils.NSBuildParentContext(childEl)
		if err != nil {
			return nil, err
		}
		ctx, err = ctx.SubContext(childEl)
		if err != nil {
			return nil, err
		}

		ns, err := ctx.LookupPrefix(childEl.Space)
		if err != nil {
			return nil, fmt.Errorf("[%s]:%s cannot find prefix %s: %v", childNS, childTag, childEl.Space, err)
		}
		if ns != childNS {
			continue
		}

		return childEl, nil
	}
	return nil, nil
}