package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
)

func printPrivateKey(key *rsa.PrivateKey) string {
	//bytes := x509.MarshalPKCS1PrivateKey(key)
	bytes, _ := x509.MarshalPKCS8PrivateKey(key)
	return base64.StdEncoding.EncodeToString(bytes)
}

func printPublicKey(key *rsa.PublicKey) string {
	bytes := x509.MarshalPKCS1PublicKey(key)
	return base64.StdEncoding.EncodeToString(bytes)
}