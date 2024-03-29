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

package filecerts

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "encoding/pem"
    "github.com/cisco-open/go-lanai/pkg/certs"
    certsource "github.com/cisco-open/go-lanai/pkg/certs/source"
    "io"
    "os"
    "path/filepath"
)

type FileProvider struct {
	p SourceProperties
}

func NewFileProvider(p SourceProperties) certs.Source {
	return &FileProvider{
		p: p,
	}
}

func (f *FileProvider) TLSConfig(ctx context.Context, _ ...certs.TLSOptions) (*tls.Config, error) {
	rootCAs, e := f.RootCAs(ctx)
	if e != nil {
		return nil, e
	}
	minVer, e := certsource.ParseTLSVersion(f.p.MinTLSVersion)
	if e != nil {
		return nil, e
	}
	//nolint:gosec // false positive -  G402: TLS MinVersion too low
	return &tls.Config{
		GetClientCertificate: f.toGetClientCertificateFunc(),
		RootCAs: rootCAs,
		MinVersion: minVer,
	}, nil
}

func (f *FileProvider) Files(_ context.Context) (*certs.CertificateFiles, error) {
	return &certs.CertificateFiles{
		RootCAPaths:          []string{f.toAbsPath(f.p.CACertFile)},
		CertificatePath:      f.toAbsPath(f.p.CertFile),
		PrivateKeyPath:       f.toAbsPath(f.p.KeyFile),
		PrivateKeyPassphrase: f.p.KeyPass,
	}, nil
}

func (f *FileProvider) RootCAs(_ context.Context) (*x509.CertPool, error) {
	caPem, err := os.ReadFile(f.p.CACertFile)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caPem)
	return certPool, nil
}

func (f *FileProvider) toGetClientCertificateFunc() func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	return func(certificateReq *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		keyFile, err := os.Open(f.p.KeyFile)
		if err != nil {
			return nil, err
		}

		keyBytes, err := io.ReadAll(keyFile)
		if err != nil {
			return nil, err
		}
		if f.p.KeyPass != "" {
			keyBlock, _ := pem.Decode(keyBytes)
			//nolint:staticcheck
			unEncryptedKey, e := x509.DecryptPEMBlock(keyBlock, []byte(f.p.KeyPass))
			if e != nil {
				return nil, e
			}
			keyBlock.Bytes = unEncryptedKey
			keyBlock.Headers = nil
			keyBytes = pem.EncodeToMemory(keyBlock)
		}
		certfile, err := os.Open(f.p.CertFile)
		if err != nil {
			return nil, err
		}
		certBytes, err := io.ReadAll(certfile)
		if err != nil {
			return nil, err
		}
		clientCert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, err
		}

		e := certificateReq.SupportsCertificate(&clientCert)
		if e != nil {
			// No acceptable certificate found. Don't send a certificate. Don't need to treat as error.
			// see tls package's tls.Conn.getClientCertificate(cri *CertificateRequestInfo) (*Certificate, error)
			return new(tls.Certificate), nil //nolint:nilerr
		} else {
			return &clientCert, nil
		}
	}
}

func (f *FileProvider) toAbsPath(path string) string {
	abs, e := filepath.Abs(path)
	if e != nil {
		return path
	}
	return abs
}
