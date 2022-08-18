package saml_util

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/httperr"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/russellhaering/goxmldsig/etreeutils"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func ParseMetadataFromXml(xml string) (*saml.EntityDescriptor, []byte, error) {
	data := []byte(xml)
	metadata, err := samlsp.ParseMetadata(data)
	return metadata, data, err
}

func ParseMetadataFromFile(fileLocation string) (*saml.EntityDescriptor, []byte, error) {
	file, err := os.Open(fileLocation)
	if err != nil {
		return nil, nil, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}
	metadata, err := samlsp.ParseMetadata(data)
	return metadata, data, err
}

func FetchMetadata(ctx context.Context, httpClient *http.Client, metadataURL url.URL) (*saml.EntityDescriptor, []byte, error) {
	req, err := http.NewRequest("GET", metadataURL.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, nil, httperr.Response(*resp)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, data, err
	}

	metadata, err := samlsp.ParseMetadata(data)
	return metadata, data, err
}

func ResolveMetadata(metadataSource string, httpClient *http.Client) (*saml.EntityDescriptor, []byte, error) {
	if strings.HasPrefix(metadataSource, "<") {
		return ParseMetadataFromXml(metadataSource)
	}
	metadataUrl, err := url.Parse(metadataSource)

	if err != nil {
		return nil, nil, err
	}
	//if it's not url or file url, assume it's relative path
	if metadataUrl.Scheme == "file" || metadataUrl.Scheme == "" {
		return ParseMetadataFromFile(metadataUrl.Path)
	} else {
		return FetchMetadata(context.TODO(), httpClient, *metadataUrl)
	}
}

func VerifySignature(data []byte, trustedCerts ...*x509.Certificate) error {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return errors.New("error parsing metadata for signature verification")
	}

	el := doc.Root()

	sigEl, err := FindChild(el, "http://www.w3.org/2000/09/xmldsig#", "Signature")

	if err != nil || sigEl == nil {
		return errors.New("metadata is not signed")
	}

	certificateStore := dsig.MemoryX509CertificateStore{
		Roots: trustedCerts,
	}

	validationContext := dsig.NewDefaultValidationContext(&certificateStore)
	validationContext.IdAttribute = "ID"
	if saml.Clock != nil {
		validationContext.Clock = saml.Clock
	}

	//if there's signature but keyInfo is not X509, then we remove the key info element, and just use the
	//default public key to verify.
	//if keyinfo is x509, it'll be verified that it's a trusted key before being used to verify the signature
	//See the logic in validationContext.Validate
	if el.FindElement("./Signature/KeyInfo/X509Data/X509Certificate") == nil {
		if keyInfo := sigEl.FindElement("KeyInfo"); keyInfo != nil {
			sigEl.RemoveChild(keyInfo)
		}
	}

	ctx, err := etreeutils.NSBuildParentContext(el)
	if err != nil {
		return errors.New("error getting document context for signature check")
	}
	ctx, err = ctx.SubContext(el)
	if err != nil {
		return errors.New("error getting document sub context for signature check")
	}
	//makes a copy of the element
	el, err = etreeutils.NSDetatch(ctx, el)
	if err != nil {
		return errors.New("error getting document for signature check")
	}

	_, err = validationContext.Validate(el)

	if err != nil {
		return errors.New("invalid signature")
	}

	return nil
}

func FindChild(parentEl *etree.Element, childNS string, childTag string) (*etree.Element, error) {
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

type ParsableSamlTypes interface {
	saml.LogoutRequest | saml.LogoutResponse | saml.AuthnRequest | saml.Response
}

type SAMLObjectParseResult struct {
	Binding string
	Encoded string
	Decoded []byte
	Err     error
}

func ParseSAMLObject[T ParsableSamlTypes](gc *gin.Context, dest *T) (ret SAMLObjectParseResult) {
	param := "SAMLResponse"
	var i interface{} = dest
	switch i.(type) {
	case *saml.LogoutRequest, *saml.AuthnRequest:
		param = "SAMLRequest"
	}

	ret.Binding = saml.HTTPRedirectBinding
	if ret.Encoded, _ = gc.GetQuery(param); len(ret.Encoded) == 0 {
		ret.Encoded = gc.PostForm(param)
		ret.Binding = saml.HTTPPostBinding
	}
	if len(ret.Encoded) == 0 {
		ret.Err = fmt.Errorf("unable to find %s in http request", param)
		return
	}

	ret.Decoded, ret.Err = base64.StdEncoding.DecodeString(ret.Encoded)
	if ret.Err != nil {
		return
	}

	ret.Err = xml.Unmarshal(ret.Decoded, dest)
	return
}
