package samlutils

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"io"
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

