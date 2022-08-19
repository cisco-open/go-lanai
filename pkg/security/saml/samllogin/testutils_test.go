package samllogin

import (
	"bytes"
	"compress/flate"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	samllib "github.com/crewjam/saml"
	"io/ioutil"
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
	SamlProperties saml.SamlProperties
	Binding        string
	Subject        string
	ExpectedMsg    string
}

func (a SamlRequestMatcher) extract(actual interface{}) (*ActualSamlRequest, error) {
	switch a.Binding {
	case samllib.HTTPPostBinding:
		return a.extractPost(actual)
	case samllib.HTTPRedirectBinding:
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
	resp := actual.(*httptest.ResponseRecorder).Result()
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
	defer func(){ _ = r.Close() }()
	reqDecoded, e := ioutil.ReadAll(r)
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

type AuthRequestMatcher struct {
	SamlRequestMatcher
}

func NewPostAuthRequestMatcher(props saml.SamlProperties) *AuthRequestMatcher {
	return &AuthRequestMatcher{
		SamlRequestMatcher{
			SamlProperties: props,
			Binding:        samllib.HTTPPostBinding,
			Subject:        "auth request",
			ExpectedMsg:    "HTML with form posting",
		},
	}
}

func NewRedirectAuthRequestMatcher(props saml.SamlProperties) *AuthRequestMatcher {
	return &AuthRequestMatcher{
		SamlRequestMatcher{
			SamlProperties: props,
			Binding:        samllib.HTTPRedirectBinding,
			Subject:        "auth request",
			ExpectedMsg:    "redirect with queries",
		},
	}
}

func (a AuthRequestMatcher) Match(actual interface{}) (success bool, err error) {
	req, e := a.extract(actual)
	if e != nil {
		return false, e
	}

	if req.Location != "https://dev-940621.oktapreview.com/app/dev-940621_samlservicelocalgo_1/exkwj65c2kC1vwtYi0h7/sso/saml" {
		return false, errors.New("incorrect request destination")
	}

	nameIdPolicy := req.XMLDoc.FindElement("//samlp:NameIDPolicy")
	if a.SamlProperties.NameIDFormat == "" {
		if nameIdPolicy.SelectAttr("Format").Value != "urn:oasis:names:tc:SAML:2.0:nameid-format:transient" {
			return false, errors.New("NameIDPolicy format should be transient if it's not configured in our properties")
		}
	} else if a.SamlProperties.NameIDFormat == "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified" {
		if nameIdPolicy.SelectAttr("Format") != nil {
			return false, errors.New("NameIDPolicy should not have a format, if we configure it to be unspecified")
		}
	} else {
		if nameIdPolicy.SelectAttr("Format").Value != a.SamlProperties.NameIDFormat {
			return false, errors.New("NameIDPolicy format should match our configuration")
		}
	}
	return true, nil
}


/********************
	Mocks
 ********************/






