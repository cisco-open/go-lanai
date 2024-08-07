// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package cryptoutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
	"strings"
)

func LoadCert(file string) ([]*x509.Certificate, error) {
	var result []*x509.Certificate

	certFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	certBytes, err := io.ReadAll(certFile)
	if err != nil {
		return nil, err
	}
	for block, r := pem.Decode(certBytes); block != nil; block, r = pem.Decode(r) {
		var cert *x509.Certificate
		switch {
		case block.Type == "CERTIFICATE":
			cert, err = x509.ParseCertificate(block.Bytes)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, cert)
	}
	return result, err
}

func LoadPrivateKey(file string, keyPassword string) (*rsa.PrivateKey, error) {
	keyFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	keyBytes, err := io.ReadAll(keyFile)
	if err != nil {
		return nil, err
	}
	keyBlock, _ := pem.Decode(keyBytes)
	if keyPassword != "" {
		//nolint:staticcheck // TODO find alternative
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

// LoadMultiBlockPem load items (cert, private key, public key, etc.) from pem file.
// Supported block types are
//   - * PRIVATE KEY
//   - PUBLIC KEY
//   - CERTIFICATE
func LoadMultiBlockPem(path string, password string) ([]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	result := []interface{}{}
	for block, r := pem.Decode(data); block != nil; block, r = pem.Decode(r) {
		var item interface{}
		var e error
		switch {
		case strings.HasSuffix(block.Type, "PRIVATE KEY"):
			item, e = parsePrivateKey(block, password)
		case strings.HasSuffix(block.Type, "PUBLIC KEY"):
			item, e = parsePublicKey(block)
		case block.Type == "CERTIFICATE":
			item, e = parseX509Cert(block)
		default:
			item = block
		}
		if e != nil {
			return nil, e
		}
		result = append(result, item)
	}
	return result, nil
}

func parsePrivateKey(block *pem.Block, password string) (interface{}, error) {
	data := block.Bytes
	if password != "" {
		//nolint:staticcheck // TODO find alternative
		decrypted, e := x509.DecryptPEMBlock(block, []byte(password))
		if e != nil {
			return nil, e
		}
		data = decrypted
	}

	// try PKCS8 first
	if key, e := x509.ParsePKCS8PrivateKey(data); e == nil {
		return key, nil
	}

	// fallback to PKCS1
	// this only handles RSA keys
	if key, e := x509.ParsePKCS1PrivateKey(data); e == nil {
		return key, nil
	}
	// this handles EC keys
	return x509.ParseECPrivateKey(data)
}

func parsePublicKey(block *pem.Block) (interface{}, error) {
	// try PKIX first (there's no pkcs8 for public keys because it's for private keys only)
	if key, e := x509.ParsePKIXPublicKey(block.Bytes); e == nil {
		return key, nil
	}
	// fallback to PKCS1
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func parseX509Cert(block *pem.Block) (*x509.Certificate, error) {
	return x509.ParseCertificate(block.Bytes)
}
