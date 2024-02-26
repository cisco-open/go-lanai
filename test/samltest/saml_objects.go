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
    "encoding/base64"
    "fmt"
    "github.com/beevik/etree"
    "github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
    "github.com/crewjam/saml"
    "github.com/google/uuid"
    "net/url"
    "time"
)

// MakeAuthnRequest create a SAML AuthnRequest, sign it and returns
func MakeAuthnRequest(sp saml.ServiceProvider, idpUrl string) string {
	authnRequest, _ := sp.MakeAuthenticationRequest(idpUrl, saml.HTTPPostBinding, saml.HTTPPostBinding)
	doc := etree.NewDocument()
	doc.SetRoot(authnRequest.Element())
	reqBuf, _ := doc.WriteToBytes()
	encodedReqBuf := base64.StdEncoding.EncodeToString(reqBuf)

	data := url.Values{}
	data.Set("SAMLRequest", encodedReqBuf)
	data.Add("RelayState", "my_relay_state")
	return data.Encode()
}

type AttributeOptions func(attr *saml.Attribute)
func MockAttribute(name, value string, opts ...AttributeOptions) saml.Attribute {
	attr := saml.Attribute{
		FriendlyName: name,
		Name:         name,
		NameFormat:   "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
		Values:       []saml.AttributeValue{{
			Type:   "xs:string",
			Value:  value,
		}},
	}
	for _, fn := range opts {
		fn(&attr)
	}
	return attr
}

type AssertionOptions func(opt *AssertionOption)
type AssertionOption struct {
	Issuer       string // entity ID
	NameID       string
	NameIDFormat string
	Recipient    string
	Audience     string // entity ID
	RequestID    string
	Attributes   []saml.Attribute
}

func MockAssertion(opts ...AssertionOptions) *saml.Assertion {
	opt := AssertionOption{
		NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
		RequestID:    uuid.New().String(),
	}
	for _, fn := range opts {
		fn(&opt)
	}
	now := time.Now()
	assertion := &saml.Assertion{
		ID:           fmt.Sprintf("id-%x", cryptoutils.RandomBytes(20)),
		IssueInstant: saml.TimeNow(),
		Version:      "2.0",
		Issuer: saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  opt.Issuer,
		},
		Subject: &saml.Subject{
			NameID: &saml.NameID{
				Format: opt.NameIDFormat,
				Value:  opt.NameID,
			},
			SubjectConfirmations: []saml.SubjectConfirmation{
				{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: &saml.SubjectConfirmationData{
						InResponseTo: opt.RequestID,
						NotOnOrAfter: now.Add(saml.MaxIssueDelay),
						Recipient:    opt.Recipient,
					},
				},
			},
		},
		Conditions: &saml.Conditions{
			NotBefore:    now,
			NotOnOrAfter: now.Add(saml.MaxIssueDelay),
			AudienceRestrictions: []saml.AudienceRestriction{
				{
					Audience: saml.Audience{Value: opt.Audience},
				},
			},
		},
		AuthnStatements: []saml.AuthnStatement{
			{
				AuthnInstant: now,
				AuthnContext: saml.AuthnContext{
					AuthnContextClassRef: &saml.AuthnContextClassRef{
						Value: "urn:oasis:names:tc:SAML:2.0:ac:classes:Password",
					},
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{
				Attributes: opt.Attributes,
			},
		},
	}
	return assertion
}

type LogoutResponseOptions func(opt *LogoutResponseOption)
type LogoutResponseOption struct {
	Issuer    string // entity ID
	Recipient string
	Audience  string // entity ID
	RequestID string
	Success   bool
}

func MockLogoutResponse(opts ...LogoutResponseOptions) *saml.LogoutResponse {
	opt := LogoutResponseOption{
		RequestID: uuid.New().String(),
		Success:   true,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	status := saml.StatusSuccess
	if !opt.Success {
		status = saml.StatusAuthnFailed
	}
	now := time.Now()
	resp := &saml.LogoutResponse{
		ID:           fmt.Sprintf("id-%x", uuid.New()),
		InResponseTo: opt.RequestID,
		Version:      "2.0",
		IssueInstant: now,
		Destination:  opt.Recipient,
		Issuer: &saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  opt.Issuer,
		},
		Status: saml.Status{
			StatusCode: saml.StatusCode{
				Value: status,
			},
		},
	}
	return resp
}
