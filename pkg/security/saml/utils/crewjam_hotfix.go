package samlutils

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"net/url"
	"strings"
)

// Redirect this is copied from saml.AuthnRequest.Redirect.
// As of crewjam/saml 0.4.8, crewjam/saml made an attempt of fixing saml.AuthnRequest.Redirect with correct Signature:
// 	https://github.com/crewjam/saml/pull/339
// However, per SAML 2.0 Binding protocol specs, the signing should only apply to query "SAMLRequest=value&RelayState=value&SigAlg=value",
// but crewjam/saml 0.4.8 uses all query string for signing.
// See https://www.oasis-open.org/committees/download.php/35387/sstc-saml-bindings-errata-2.0-wd-05-diff.pdf
// TODO revisit this part later when newer crewjam/saml library become available
func redirectUrl(relayState string, sp *saml.ServiceProvider, rootEl *etree.Element, dest string) (*url.URL, error) {
	w := &bytes.Buffer{}
	w1 := base64.NewEncoder(base64.StdEncoding, w)
	w2, _ := flate.NewWriter(w1, 9)
	doc := etree.NewDocument()
	doc.SetRoot(rootEl)
	if _, err := doc.WriteTo(w2); err != nil {
		return nil, err
	}
	_ = w2.Close()
	_ = w1.Close()

	rawKVs := make([]string, 1, 3)
	rv, _ := url.Parse(dest)
	query := rv.Query()

	rawKVs[0] = HttpParamSAMLRequest + "=" + url.QueryEscape(string(w.Bytes()))
	query.Set(HttpParamSAMLRequest, string(w.Bytes()))

	if relayState != "" {
		rawKVs = append(rawKVs, HttpParamRelayState+ "=" + url.QueryEscape(relayState))
		query.Set(HttpParamRelayState, relayState)
	}
	if len(sp.SignatureMethod) > 0 {
		rawKVs = append(rawKVs, HttpParamSigAlg+ "=" + url.QueryEscape(sp.SignatureMethod))
		query.Set(HttpParamSigAlg, sp.SignatureMethod)

		signingContext, e := saml.GetSigningContext(sp)
		if e != nil {
			return nil, e
		}

		sig, e := signingContext.SignString(strings.Join(rawKVs, "&"))
		if e != nil {
			return nil, e
		}

		sigVal := base64.StdEncoding.EncodeToString(sig)
		query.Set(HttpParamSignature, sigVal)
	}
	rv.RawQuery = query.Encode()
	return rv, nil
}

/***********************
	AuthnRequest
 ***********************/

type FixedAuthnRequest struct {
	saml.AuthnRequest
}

func NewFixedAuthenticationRequest(sp *saml.ServiceProvider, idpURL string, binding string, resultBinding string) (*FixedAuthnRequest, error) {
	req, e := sp.MakeAuthenticationRequest(idpURL, binding, resultBinding)
	if e != nil {
		return nil, e
	}
	return &FixedAuthnRequest{*req}, nil
}

// Redirect crewjam/saml 0.4.8 hotfix.
func (req *FixedAuthnRequest) Redirect(relayState string, sp *saml.ServiceProvider) (*url.URL, error) {
	// per SAML 2.0 spec, Signature element should be removed from xml in case of redirect binding
	req.Signature = nil
	return redirectUrl(relayState, sp, req.Element(), req.Destination)
}

/***********************
	LogoutRequest
 ***********************/

type FixedLogoutRequest struct {
	saml.LogoutRequest
}

func NewFixedLogoutRequest(sp *saml.ServiceProvider, idpURL, nameID string) (*FixedLogoutRequest, error) {
	req, e := sp.MakeLogoutRequest(idpURL, nameID)
	if e != nil {
		return nil, e
	}
	return &FixedLogoutRequest{*req}, nil
}

// Redirect crewjam/saml 0.4.8 hotfix.
func (req *FixedLogoutRequest) Redirect(relayState string, sp *saml.ServiceProvider) (*url.URL, error) {
	// per SAML 2.0 spec, Signature element should be removed from xml in case of redirect binding
	req.Signature = nil
	return redirectUrl(relayState, sp, req.Element(), req.Destination)
}


