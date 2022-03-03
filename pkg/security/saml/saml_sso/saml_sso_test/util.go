package saml_sso_test

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"encoding/base64"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"io"
	"net/url"
)

func MakeAuthnRequest(sp saml.ServiceProvider, idpUrl string) string {
	authnRequest, _ := sp.MakeAuthenticationRequest(idpUrl)
	doc := etree.NewDocument()
	doc.SetRoot(authnRequest.Element())
	reqBuf, _ := doc.WriteToBytes()
	encodedReqBuf := base64.StdEncoding.EncodeToString(reqBuf)

	data := url.Values{}
	data.Set("SAMLRequest", encodedReqBuf)
	data.Add("RelayState", "my_relay_state")
	return data.Encode()
}

func ParseSamlResponse(r io.Reader) (*etree.Document, error) {
	html := etree.NewDocument()
	if _, err := html.ReadFrom(r); err != nil {
		return nil, err
	}

	input := html.FindElement("//input[@name='SAMLResponse']")
	samlResponse := input.SelectAttrValue("value", "")
	data, err := base64.StdEncoding.DecodeString(samlResponse)

	if err != nil {
		return nil, err
	}

	samlResponseXml := etree.NewDocument()
	err = samlResponseXml.ReadFromBytes(data)

	if err != nil {
		return nil, err
	}
	return samlResponseXml, nil
}

func NewSamlSp(spUrl string, certFilePath string, keyFilePath string) saml.ServiceProvider {
	rootURL, _ := url.Parse(spUrl)
	cert, _ := cryptoutils.LoadCert(certFilePath)
	key, _ := cryptoutils.LoadPrivateKey(keyFilePath, "")
	sp := samlsp.DefaultServiceProvider(samlsp.Options{
		URL:            *rootURL,
		Key:            key,
		Certificate:    cert[0],
		SignRequest: true,
		EntityID: fmt.Sprintf("%s/saml/metadata", spUrl),
	})
	return sp
}