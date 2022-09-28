package samlutils

import "errors"

const SAMLEncodingDeflate   = `urn:oasis:names:tc:SAML:2.0:bindings:URL-Encoding:DEFLATE`
const (
	HttpParamSAMLRequest  = `SAMLRequest`
	HttpParamSAMLResponse = `SAMLResponse`
	HttpParamSigAlg       = `SigAlg`
	HttpParamSignature    = `Signature`
	HttpParamRelayState   = `RelayState`
	HttpParamSAMLEncoding = `SAMLEncoding`
)


var ErrorXMLNotSigned = errors.New("XML document is not signed")
