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

package samltest

import (
	"bytes"
	"compress/flate"
	samlutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type BindableSamlTypes interface {
	saml.LogoutRequest | saml.LogoutResponse | saml.AuthnRequest | saml.Response
}

// xmlSerializer is a common interface of saml.LogoutRequest, saml.LogoutResponse, saml.AuthnRequest, saml.Response
type xmlSerializer interface {
	Element() *etree.Element
}

// RequestWithSAMLPostBinding returns a webtest.RequestOptions that inject given SAML Request/Response using Post binding.
// Note: request need to be POST
func RequestWithSAMLPostBinding[T BindableSamlTypes](samlObj *T, relayState string) webtest.RequestOptions {
	return func(req *http.Request) {
		var i interface{} = samlObj
		serializer, ok := i.(xmlSerializer)
		if !ok {
			panic(fmt.Sprintf("%T doess not have Element() *etree.Element", i))
		}
		doc := etree.NewDocument()
		doc.SetRoot(serializer.Element())
		decoded, e := doc.WriteToBytes()
		if e != nil {
			panic(e)
		}
		encoded := base64.StdEncoding.EncodeToString(decoded)
		values := url.Values{}
		values.Set("SAMLRequest", encoded)
		values.Add("RelayState", relayState)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Body = io.NopCloser(strings.NewReader(values.Encode()))
	}
}

type BindingParseResult struct {
	Binding string
	Values  url.Values
	Encoded string
	Decoded []byte
}

// ParseBinding parse redirect/post binding from given HTTP response
func ParseBinding[T samlutils.ParsableSamlTypes](resp *http.Response, dest *T) (ret BindingParseResult, err error) {
	param := samlutils.HttpParamSAMLResponse
	var i interface{} = dest
	switch i.(type) {
	case *saml.LogoutRequest, *saml.AuthnRequest:
		param = samlutils.HttpParamSAMLRequest
	}

	switch {
	case resp.StatusCode < 300:
		ret.Binding = saml.HTTPPostBinding
		ret.Values, err = extractPostBindingValues(resp)
	default:
		ret.Binding = saml.HTTPRedirectBinding
		ret.Values, err = extractRedirectBindingValues(resp)
	}
	if err != nil {
		return
	}

	ret.Encoded = ret.Values.Get(param)
	if len(ret.Encoded) == 0 {
		return ret, fmt.Errorf("unable to find %s in http response", param)
	}

	ret.Decoded, err = base64.StdEncoding.DecodeString(ret.Encoded)
	if err != nil {
		return
	}

	// try de-compress
	r := flate.NewReader(bytes.NewReader(ret.Decoded))
	if data, e := io.ReadAll(r); e == nil {
		ret.Decoded = data
	}
	err = xml.Unmarshal(ret.Decoded, dest)
	return
}

func extractRedirectBindingValues(resp *http.Response) (url.Values, error) {
	loc := resp.Header.Get("Location")
	locUri, e := url.Parse(loc)
	if e != nil {
		return nil, e
	}
	return locUri.Query(), nil
}

func extractPostBindingValues(resp *http.Response) (url.Values, error) {
	htmlDoc := etree.NewDocument()
	if _, e := htmlDoc.ReadFrom(resp.Body); e != nil {
		return nil, e
	}

	values := url.Values{}
	elems := htmlDoc.FindElements("//input")
	for _, el := range elems {
		if typ := el.SelectAttrValue("type", ""); typ == "submit" {
			continue
		}
		name := el.SelectAttrValue("name", "unknown")
		value := el.SelectAttrValue("value", "")
		values.Add(name, value)
	}
	return values, nil
}
