package cryptoutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
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
	if keyPassword != "" {
		unEncryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(keyPassword))
		if err != nil {
			return nil, err
		}
		key, err := x509.ParsePKCS1PrivateKey(unEncryptedKey)
		return key, err
	} else {
		key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)

		if err != nil {
			return nil, err
		}
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, err
		} else {
			return nil, errors.New("private key is not rsa key")
		}
	}
}

// RandReader is the io.Reader that produces cryptographically random
// bytes when they are need by the library. The default value is
// rand.Reader, but it can be replaced for testing.
var RandReader = rand.Reader

func RandomBytes(n int) []byte {
	rv := make([]byte, n)

	if _, err := io.ReadFull(RandReader, rv); err != nil {
		panic(err)
	}
	return rv
}