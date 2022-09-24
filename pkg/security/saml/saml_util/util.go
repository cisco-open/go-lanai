package saml_util

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/httperr"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/russellhaering/goxmldsig/etreeutils"
	"golang.org/x/net/html"
	"io"
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

func ResolveMetadata(ctx context.Context, metadataSource string, httpClient *http.Client) (*saml.EntityDescriptor, []byte, error) {
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
		return FetchMetadata(ctx, httpClient, *metadataUrl)
	}
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

const (
	HttpParamSAMLRequest  = `SAMLRequest`
	HttpParamSAMLResponse = `SAMLResponse`
)

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
	param := HttpParamSAMLResponse
	var i interface{} = dest
	switch i.(type) {
	case *saml.LogoutRequest, *saml.AuthnRequest:
		param = HttpParamSAMLRequest
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

	// try de-compress
	r := flate.NewReader(bytes.NewReader(ret.Decoded))
	if data, e := io.ReadAll(r); e == nil {
		ret.Decoded = data
	}

	ret.Err = xml.Unmarshal(ret.Decoded, dest)
	return
}

// WritePostBindingForm takes HTML of a request/response submitting form and wrap it in HTML document with proper
// script security tags and send it to given ResponseWriter
func WritePostBindingForm(formHtml []byte, rw http.ResponseWriter) error {
	rw.Header().Add("Content-Type", "text/html")

	body := append([]byte(`<!DOCTYPE html><html><body>`), formHtml...)
	body = append(body, []byte(`</body></html>`)...)
	csp := fmt.Sprintf("default-src; script-src %s; reflected-xss block; referrer no-referrer;", scriptSrcHash(body))
	rw.Header().Add("Content-Security-Policy", csp)
	_, e := rw.Write(body)
	return e
}

// scriptSrcHash returns '<hash-algorithm>-<base64-value>' of all inline <script></script> found in given html, delimited by space
// See CSP specs
func scriptSrcHash(htmlBytes []byte) string {
	const fallback = `'unsafe-inline'`
	root, e := html.Parse(bytes.NewReader(htmlBytes))
	if e != nil {
		return fallback
	}
	scripts := findAllHtmlNodes(root, func(node *html.Node) bool {
		return node.Type == html.TextNode && node.Parent != nil && node.Parent.Data == "script"
	})

	srcs := make([]string, len(scripts))
	for i, node := range scripts {
		hash := sha256.Sum256([]byte(node.Data))
		srcs[i] = fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(hash[:]))
	}
	return strings.Join(srcs, " ")
}

func findAllHtmlNodes(node *html.Node, matcher func(*html.Node) bool) (found []*html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if matcher(child) {
			found = append(found, child)
		}
		if sub := findAllHtmlNodes(child, matcher); len(sub) != 0 {
			found = append(found, sub...)
		}
	}
	return
}
