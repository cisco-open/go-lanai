package samltest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"encoding/base64"
	"fmt"
	"github.com/beevik/etree"
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

type AssertionOptions func(opt *AssertionOption)
type AssertionOption struct {
	Issuer       string // entity ID
	NameID       string
	NameIDFormat string
	Recipient    string
	Audience     string // entity ID
	RequestID    string
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
				Attributes: []saml.Attribute{},
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
