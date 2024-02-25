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

package testdata

import (
    "bytes"
    "compress/flate"
    "encoding/base64"
    "fmt"
    "github.com/beevik/etree"
    lanaisaml "github.com/cisco-open/go-lanai/pkg/security/saml"
    "github.com/crewjam/saml"
    "io"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
)

type ActualSamlRequest struct {
	XMLDoc     *etree.Document
	Location   string
	RelayState string
	SigAlg     string
	Signature  string
}

type SamlRequestMatcher struct {
	SamlProperties lanaisaml.SamlProperties
	Binding        string
	Subject        string
	ExpectedMsg    string
}

func (a SamlRequestMatcher) Extract(actual interface{}) (*ActualSamlRequest, error) {
	switch a.Binding {
	case saml.HTTPPostBinding:
		return a.extractPost(actual)
	case saml.HTTPRedirectBinding:
		return a.extractRedirect(actual)
	default:
		return nil, fmt.Errorf("unable to verify %s with binding '%s'", a.Subject, a.Binding)
	}
}

func (a SamlRequestMatcher) extractPost(actual interface{}) (*ActualSamlRequest, error) {
	w := actual.(*httptest.ResponseRecorder)

	html := etree.NewDocument()
	if _, e := html.ReadFrom(w.Body); e != nil {
		return nil, e
	}

	formElem := html.FindElement("//form[@action]")
	if formElem == nil {
		return nil, fmt.Errorf("form with is not found in HTML")
	}

	reqElem := html.FindElement("//input[@name='SAMLRequest']")
	if reqElem == nil {
		return nil, fmt.Errorf("form doesn't contain 'SAMLRequest' value in HTML")
	}

	reqDecoded, e := base64.StdEncoding.DecodeString(reqElem.SelectAttrValue("value", ""))
	if e != nil {
		return nil, e
	}

	req := ActualSamlRequest{
		XMLDoc:   etree.NewDocument(),
		Location: formElem.SelectAttrValue("action", ""),
	}
	if e := req.XMLDoc.ReadFromBytes(reqDecoded); e != nil {
		return nil, e
	}

	if elem := html.FindElement("//input[@name='RelayState']"); elem != nil {
		req.RelayState = elem.SelectAttrValue("value", "")
	}

	if elem := req.XMLDoc.FindElement("//ds:SignatureMethod"); elem != nil {
		req.SigAlg = elem.SelectAttrValue("Algorithm", "")
	}

	if elem := req.XMLDoc.FindElement("//ds:SignatureValue"); elem != nil {
		req.Signature = elem.Text()
	}

	return &req, nil
}

func (a SamlRequestMatcher) extractRedirect(actual interface{}) (*ActualSamlRequest, error) {
	var resp *http.Response
	switch v := actual.(type) {
	case *httptest.ResponseRecorder:
		resp = v.Result()
	case *http.Response:
		resp = v
	}
	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		return nil, fmt.Errorf("not redirect")
	}

	loc := resp.Header.Get("Location")
	locUrl, e := url.Parse(loc)
	if e != nil {
		return nil, e
	}
	loc = loc[:strings.IndexRune(loc, '?')]

	// Note redirect request is compressed
	compressed, e := base64.StdEncoding.DecodeString(locUrl.Query().Get("SAMLRequest"))
	if e != nil {
		return nil, e
	}
	r := flate.NewReader(bytes.NewReader(compressed))
	defer func() { _ = r.Close() }()
	reqDecoded, e := io.ReadAll(r)
	if e != nil {
		return nil, e
	}

	req := ActualSamlRequest{
		XMLDoc:     etree.NewDocument(),
		Location:   loc,
		RelayState: locUrl.Query().Get("RelayState"),
		SigAlg:     locUrl.Query().Get("SigAlg"),
		Signature:  locUrl.Query().Get("Signature"),
	}
	if e := req.XMLDoc.ReadFromBytes(reqDecoded); e != nil {
		return nil, e
	}
	return &req, nil
}

func (a SamlRequestMatcher) FailureMessage(actual interface{}) (message string) {
	w := actual.(*httptest.ResponseRecorder)
	body := string(w.Body.Bytes())
	return fmt.Sprintf("Expected %s as %s. Actual: %s", a.Subject, a.ExpectedMsg, body)
}

func (a SamlRequestMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	w := actual.(*httptest.ResponseRecorder)
	body := string(w.Body.Bytes())
	return fmt.Sprintf("Expected %s as %s. Actual: %s", a.Subject, a.ExpectedMsg, body)
}
