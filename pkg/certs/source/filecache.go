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

package certsource

import (
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultCacheRoot         = `.tmp/certs`
	CachedFileKeyCertificate = `cert`
	CachedFileKeyPrivateKey  = `key`
	CachedFileKeyCA          = `ca`
)

type FileCacheOptions func(opt *FileCacheOption)
type FileCacheOption struct {
	Root   string
	Type   certs.SourceType
	Prefix string
}

func NewFileCache(opts ...FileCacheOptions) (*FileCache, error) {
	opt := FileCacheOption{}
	for _, fn := range opts {
		fn(&opt)
	}

	if len(opt.Root) == 0 {
		opt.Root = DefaultCacheRoot
	}
	if len(opt.Prefix) == 0 {
		opt.Prefix = utils.RandomString(12)
	}

	dir := filepath.Clean(filepath.Join(opt.Root, string(opt.Type)))
	e := os.MkdirAll(dir, 0755)
	if e != nil {
		return nil, e
	}
	return &FileCache{Dir: dir, Prefix: opt.Prefix}, nil
}

type FileCache struct {
	Dir    string
	Prefix string
}

// CacheCertificate will write out a cert and key to files based on configured path and prefix
func (c *FileCache) CacheCertificate(cert *tls.Certificate) error {
	if len(cert.Certificate) < 1 {
		return fmt.Errorf("no certificates present in provided tls.Certificate")
	}

	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate[0],
	}

	certBytes := pem.EncodeToMemory(pemBlock)
	if certBytes == nil {
		return fmt.Errorf("failed to encode certificate to PEM")
	}

	if err := c.CachePEM(certBytes, CachedFileKeyCertificate); err != nil {
		return fmt.Errorf("failed to write PEM data to file: %v", err)
	}

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to marshal private key: %v", err)
	}

	privKeyPem := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	}

	keyBytes := pem.EncodeToMemory(privKeyPem)
	if keyBytes == nil {
		return fmt.Errorf("failed to encode private key to PEM")
	}

	if err = c.CachePEM(keyBytes, CachedFileKeyPrivateKey); err != nil {
		return fmt.Errorf("failed to write private key PEM data to file: %v", err)
	}
	return nil
}

// CachePEM write given data into file. The file name is determined by "key" and "suffix"
func (c *FileCache) CachePEM(pemData []byte, key string) error {
	path := c.ResolvePath(key)
	err := os.WriteFile(path, pemData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write PEM data to file: %v", err)
	}
	return nil
}

func (c *FileCache) ResolvePath(key string) string {
	filename := fmt.Sprintf(`%s-%s.pem`, c.Prefix, key)
	return filepath.Clean(filepath.Join(c.Dir, filename))
}
