package samllogin

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
)

func LoadCert(file string) (*x509.Certificate, error) {
	certFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	certBytes, err := ioutil.ReadAll(certFile)
	if err != nil {
		return nil, err
	}
	certBlock, _ := pem.Decode(certBytes)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	return cert, err
}

func LoadPrivateKey(file string, keyPassword string) (*rsa.PrivateKey, error){
	keyFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	keyBytes, err := ioutil.ReadAll(keyFile)
	if err != nil {
		return nil, err
	}
	keyBlock, _ := pem.Decode(keyBytes)
	unEncryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(keyPassword))
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKCS1PrivateKey(unEncryptedKey)
	return key, err
}