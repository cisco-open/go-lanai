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

package embedded

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"os"
)

type TLSCerts struct {
	FS   fs.FS
	Cert string
	Key  string
	CAs  []string
}

// ServerTLSWithCerts construct a tls.Config with certificates in a given filesystem.
// The setup Server TLS, following config are required:
// - filesystem to load files from. Default to "."
// - path of certificate file in PEM format, default to "testdata/server.crt"
// - path of certificate private key file in unencrypted PEM format, default to "testdata/server.key"
// - path of at least one CA certificate in PEM format, default to "testdata/ca.crt"
// Note: if any file is missing or not readable, the result tls.Config might not works as expected
func ServerTLSWithCerts(opts ...func(src *TLSCerts)) (*tls.Config, error) {
	src := TLSCerts{
		FS:   os.DirFS("."),
		Cert: "testdata/server.crt",
		Key:  "testdata/server.key",
		CAs:  []string{"testdata/ca.crt"},
	}
	for _, fn := range opts {
		fn(&src)
	}
	// start to load
	caPool := x509.NewCertPool()
	for _, path := range src.CAs {
		pemBytes, e := fs.ReadFile(src.FS, path)
		if e != nil {
			return nil, fmt.Errorf("unable to read CA file [%s]: %v", path, e)
		}
		caPool.AppendCertsFromPEM(pemBytes)
	}
	certBytes, e := fs.ReadFile(src.FS, src.Cert)
	if e != nil {
		return nil, fmt.Errorf("unable to read certificate file [%s]: %v", src.Cert, e)
	}
	keyBytes, e := fs.ReadFile(src.FS, src.Key)
	if e != nil {
		return nil, fmt.Errorf("unable to read private key file [%s]: %v", src.Key, e)
	}
	cert, e := tls.X509KeyPair(certBytes, keyBytes)
	if e != nil {
		return nil, fmt.Errorf("unable to parse certificate: %v", e)
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}
