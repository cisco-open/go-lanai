package samlutils

import (
	"crypto"
	"crypto/rsa"
	//nolint:gosec // weak cryptographic primitive, but we still need to support it
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/russellhaering/goxmldsig/etreeutils"
	"hash"
	"net/http"
	"strings"
)

type SignatureVerifyOptions func(sc *SignatureContext)
type SignatureContext struct {
	Binding string
	Certs   []*x509.Certificate
	Request *http.Request
	XMLData []byte
}

// MetadataSignature returns SignatureVerifyOptions for metadata validation
func MetadataSignature(data []byte, certs ...*x509.Certificate) SignatureVerifyOptions {
	return func(sc *SignatureContext) {
		sc.Binding = ""
		sc.Certs = certs
		sc.XMLData = data
	}
}

// VerifySignature verify signature of SAML Request/Response/Metadata
// This function would choose signing protocol based on bindings and provided information.
// - saml.HTTPRedirectBinding uses Deflated Encoding. SignatureContext.Request and SignatureContext.Certs is required in this mode
// - saml.HTTPPostBinding uses enveloped XMLDSign. SignatureContext.XMLData is required in this mode
// - Enveloped XMLDSign is used when Binding is any other value. SignatureContext.XMLData is required in this mode
func VerifySignature(opts...SignatureVerifyOptions) error {
	sc := SignatureContext{}
	for _, fn := range opts {
		fn(&sc)
	}
	switch sc.Binding {
	case saml.HTTPRedirectBinding:
		return verifyDeflateEncSign(&sc)
	default:
		return verifyXMLDSign(&sc)
	}
}

// verifyXMLDSign validate Enveloped XMLDSign signature, typically used for PostBinding or Metadata
func verifyXMLDSign(sc *SignatureContext) error {
	if len(sc.XMLData) == 0 {
		return errors.New("XML document is missing for signature verification")
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(sc.XMLData); err != nil {
		return errors.New("error parsing XML document for signature verification")
	}

	el := doc.Root()
	sigEl, e := FindChild(el, "http://www.w3.org/2000/09/xmldsig#", "Signature")
	if e != nil || sigEl == nil {
		return ErrorXMLNotSigned
	}

	certificateStore := dsig.MemoryX509CertificateStore{
		Roots: sc.Certs,
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

	ctx, e := etreeutils.NSBuildParentContext(el)
	if e != nil {
		return errors.New("error getting document context for signature check")
	}

	if ctx, e = ctx.SubContext(el); e != nil {
		return errors.New("error getting document sub context for signature check")
	}

	//makes a copy of the element
	if el, e = etreeutils.NSDetatch(ctx, el); e != nil {
		return errors.New("error getting document for signature check")
	}

	if _, e = validationContext.Validate(el); e != nil {
		return errors.New("invalid signature")
	}
	return nil
}

// verifyDeflateEncSign validate DEFLATE URL encoding signature, typically used for RedirectBinding of SAML Request
// https://www.oasis-open.org/committees/download.php/35387/sstc-saml-bindings-errata-2.0-wd-05-diff.pdf
func verifyDeflateEncSign(sc *SignatureContext) error {
	// some sanity check
	if sc.Request == nil {
		return fmt.Errorf("HTTP Request is required for DEFLATE Encoding signature verification")
	}
	if enc := queryValue(sc, HttpParamSAMLEncoding); len(enc) != 0 && enc != SAMLEncodingDeflate {
		return fmt.Errorf("unsupported SAML encoding [%s]", enc)
	}

	// find signature
	alg := queryValue(sc, HttpParamSigAlg)
	encodedSig := queryValue(sc, HttpParamSignature)
	if len(alg) == 0 || len(encodedSig) == 0 {
		return ErrorXMLNotSigned
	}

	sig, e := base64.StdEncoding.DecodeString(encodedSig)
	if e != nil || len(sig) == 0 {
		return fmt.Errorf("failed to decode signature")
	}

	// extract to-be-verified data
	toVerify := toBeVerifiedQuery(sc)

	// verify
	var err error
	for _, cert := range sc.Certs {
		if err = rsaVerify([]byte(toVerify), sig, cert.PublicKey, alg); err == nil {
			return nil
		}
	}
	return err
}

func toBeVerifiedQuery(sc *SignatureContext) string {
	// SAMLRequest=value&RelayState=value&SigAlg=value
	// SAMLResponse=value&RelayState=value&SigAlg=value
	// note: per SAML spec, we need to use the original URL encoded query instead of re-encoding the query
	candidates := make([]string, 3)
	rawKVs := strings.Split(sc.Request.URL.RawQuery, "&")
	for _, pair := range rawKVs {
		kv := strings.SplitN(pair, "=", 2)
		var i int
		switch kv[0] {
		case HttpParamSAMLRequest, HttpParamSAMLResponse:
			i = 0
		case HttpParamRelayState:
			i = 1
		case HttpParamSigAlg:
			i = 2
		default:
			continue
		}
		candidates[i] = pair
	}
	toVerify := make([]string, 0, len(candidates))
	for _, v := range candidates {
		if len(v) != 0 {
			toVerify = append(toVerify, v)
		}
	}
	return strings.Join(toVerify, "&")
}

func rsaVerify(data, signature []byte, pubKey any, method string) error {
	var h hash.Hash
	var hashAlg crypto.Hash
	switch method {
	case dsig.RSASHA1SignatureMethod:
		//nolint:gosec // weak cryptographic primitive, but we still need to support it
		h = sha1.New()
		hashAlg = crypto.SHA1
	case dsig.RSASHA256SignatureMethod:
		h = sha256.New()
		hashAlg = crypto.SHA256
	case dsig.RSASHA512SignatureMethod:
		h = sha512.New()
		hashAlg = crypto.SHA512
	default:
		return fmt.Errorf("unsupported signature method: %s", method)
	}
	_, _ = h.Write(data)
	hashed := h.Sum(nil)

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("RSA public key is required to verify signature")
	}

	return rsa.VerifyPKCS1v15(rsaPubKey, hashAlg, hashed, signature)
}

func queryValue(sc *SignatureContext, key string) string {
	return sc.Request.URL.Query().Get(key)
}
