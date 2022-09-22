package saml_util

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"net/url"
)

/***********************
	Workaround
 ***********************/

type FixedLogoutRequest struct {
	saml.LogoutRequest
}

func MakeFixedLogoutRequest(sp *saml.ServiceProvider, idpURL, nameID string) (*FixedLogoutRequest, error) {
	req, e := sp.MakeLogoutRequest(idpURL, nameID)
	if e != nil {
		return nil, e
	}
	return &FixedLogoutRequest{*req}, nil
}

// Redirect this is copied from saml.AuthnRequest.Redirect.
// As of crewjam/saml 0.4.8, AuthnRequest's Redirect is fixed for properly setting Signature in redirect URL:
// 	https://github.com/crewjam/saml/pull/339
// However, saml.LogoutRequest.Redirect is not fixed. We need to do that by ourselves
// TODO revisit this part later when newer crewjam/saml library become available
func (req *FixedLogoutRequest) Redirect(relayState string, sp *saml.ServiceProvider) (*url.URL, error) {
	w := &bytes.Buffer{}
	w1 := base64.NewEncoder(base64.StdEncoding, w)
	w2, _ := flate.NewWriter(w1, 9)
	doc := etree.NewDocument()
	doc.SetRoot(req.Element())
	if _, err := doc.WriteTo(w2); err != nil {
		return nil, err
	}
	_ = w2.Close()
	_ = w1.Close()

	rv, _ := url.Parse(req.Destination)
	// We can't depend on Query().set() as order matters for signing
	query := rv.RawQuery
	if len(query) > 0 {
		query += "&SAMLRequest=" + url.QueryEscape(string(w.Bytes()))
	} else {
		query += "SAMLRequest=" + url.QueryEscape(string(w.Bytes()))
	}

	if relayState != "" {
		query += "&RelayState=" + relayState
	}
	if len(sp.SignatureMethod) > 0 {
		query += "&SigAlg=" + url.QueryEscape(sp.SignatureMethod)
		signingContext, err := saml.GetSigningContext(sp)

		if err != nil {
			return nil, err
		}

		sig, err := signingContext.SignString(query)
		if err != nil {
			return nil, err
		}
		query += "&Signature=" + url.QueryEscape(base64.StdEncoding.EncodeToString(sig))
	}

	rv.RawQuery = query

	return rv, nil
}

