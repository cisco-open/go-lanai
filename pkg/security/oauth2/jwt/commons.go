package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
)

func printPrivateKey(key *rsa.PrivateKey) string {
	bytes := x509.MarshalPKCS1PrivateKey(key)
	return base64.StdEncoding.EncodeToString(bytes)
}
