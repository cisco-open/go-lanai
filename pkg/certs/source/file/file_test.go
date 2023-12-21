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

package filecerts_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	filecerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source/file"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"fmt"
	"go.uber.org/fx"
	"os"
	"testing"
)
import . "github.com/onsi/gomega"

type mgrDI struct {
	fx.In
	AppCfg    bootstrap.ApplicationConfig
	Props     certs.Properties
	Factories []certs.SourceFactory `group:"certs"`
}

func ProvideTestManager(di mgrDI) (certs.Manager, certs.Registrar) {
	reg := certs.NewDefaultManager(func(mgr *certs.DefaultManager) {
		mgr.ConfigLoaderFunc = di.AppCfg.Bind
		mgr.Properties = di.Props
	})
	for _, f := range di.Factories {
		if f != nil {
			reg.MustRegister(f)
		}
	}
	return reg, reg
}

func BindTestProperties(appCfg bootstrap.ApplicationConfig) certs.Properties {
	props := certs.NewProperties()
	if e := appCfg.Bind(props, "tls"); e != nil {
		panic(fmt.Errorf("failed to bind certificate properties: %v", e))
	}
	return *props
}

type FileTestDi struct {
	fx.In
	CertsManager certs.Manager
}

func TestFileCertificateSource(t *testing.T) {
	di := &FileTestDi{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(ProvideTestManager, BindTestProperties, filecerts.FxProvider()),
		),
		test.GomegaSubTest(SubTestTLSConfig(di), "SubTestTLSConfig"),
		test.GomegaSubTest(SubTestFiles(di), "TestFiles"),
	)
}

func SubTestTLSConfig(di *FileTestDi) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		p := filecerts.SourceProperties{
			CACertFile: "testdata/ca-cert-test.pem",
			CertFile:   "testdata/client-cert-signed-test.pem",
			KeyFile:    "testdata/client-key-test.pem",
			KeyPass:    "foobar",
		}

		tlsSrc, err := di.CertsManager.Source(ctx, certs.WithType(certs.SourceFile, p))
		g.Expect(err).NotTo(HaveOccurred())

		tlsCfg, err := tlsSrc.TLSConfig(ctx)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(tlsCfg.RootCAs).ToNot(BeNil())
		g.Expect(len(tlsCfg.RootCAs.Subjects())).To(Equal(1))
		g.Expect(tlsCfg.GetClientCertificate).ToNot(BeNil())

		//try with the ca that the cert is signed with
		// the signature scheme and version is captured from a kafka broker that uses tls connection.
		certReqInfo := &tls.CertificateRequestInfo{
			AcceptableCAs: tlsCfg.RootCAs.Subjects(),
			SignatureSchemes: []tls.SignatureScheme{
				tls.ECDSAWithP256AndSHA256,
				tls.ECDSAWithP384AndSHA384,
				tls.ECDSAWithP521AndSHA512,
				tls.PSSWithSHA256,
				tls.PSSWithSHA384,
				tls.PSSWithSHA512,
				2057,
				2058,
				2059,
				tls.PKCS1WithSHA256,
				tls.PKCS1WithSHA384,
				tls.PKCS1WithSHA512,
				tls.ECDSAWithSHA1,
				tls.PKCS1WithSHA1,
			},
			Version: 772,
		}
		clientCert, err := tlsCfg.GetClientCertificate(certReqInfo)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(clientCert).NotTo(BeNil())
		g.Expect(len(clientCert.Certificate)).To(Equal(1))

		//try with a different ca, and expect no cert is returned
		anotherCa, err := os.ReadFile("testdata/ca-cert-test-2")
		g.Expect(err).NotTo(HaveOccurred())
		anotherCaPool := x509.NewCertPool()
		anotherCaPool.AppendCertsFromPEM(anotherCa)

		certReqInfo.AcceptableCAs = anotherCaPool.Subjects()
		clientCert, err = tlsCfg.GetClientCertificate(certReqInfo)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(len(clientCert.Certificate)).To(Equal(0))
	}
}

func SubTestFiles(di *FileTestDi) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		p := filecerts.SourceProperties{
			CACertFile: "testdata/ca-cert-test.pem",
			CertFile:   "testdata/client-cert-signed-test.pem",
			KeyFile:    "testdata/client-key-test.pem",
			KeyPass:    "foobar",
		}

		tlsSrc, err := di.CertsManager.Source(ctx, certs.WithType(certs.SourceFile, p))
		g.Expect(err).NotTo(HaveOccurred())

		tlsFiles, err := tlsSrc.Files(ctx)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(tlsFiles.RootCAPaths).To(ContainElement(ContainSubstring("testdata/ca-cert-test.pem")))
		AssertFilesExist(g, tlsFiles.RootCAPaths)
		g.Expect(tlsFiles.CertificatePath).To(ContainSubstring("testdata/client-cert-signed-test.pem"))
		AssertFileExists(g, tlsFiles.CertificatePath)
		g.Expect(tlsFiles.PrivateKeyPath).To(ContainSubstring("testdata/client-key-test.pem"))
		AssertFileExists(g, tlsFiles.PrivateKeyPath)
		g.Expect(tlsFiles.PrivateKeyPassphrase).To(Equal("foobar"))
	}
}

/*************************
	Helpers
 *************************/

func AssertFilesExist(g *WithT, paths []string) {
	for _, path := range paths {
		AssertFileExists(g, path)
	}
}

func AssertFileExists(g *WithT, path string) {
	data, e := os.ReadFile(path)
	g.Expect(e).To(Succeed(), "reading file '%s' should not fail", path)
	g.Expect(data).ToNot(BeEmpty(), "file '%s' should not be empty", path)
}
