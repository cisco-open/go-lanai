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

package vaultcerts_test

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/certs"
    vaultcerts "github.com/cisco-open/go-lanai/pkg/certs/source/vault"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/pkg/vault"
    vaultinit "github.com/cisco-open/go-lanai/pkg/vault/init"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/ittest"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "gopkg.in/dnaeon/go-vcr.v3/recorder"
    "os"
    "sync"
    "testing"
    "time"
)

/*************************
	Test Setup
 *************************/

var TestCAExpiration = utils.ParseTimeISO8601("2033-11-27T23:04:45Z")

var TestCertReqInfoTmpl = tls.CertificateRequestInfo{
	AcceptableCAs: [][]byte{},
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

func RecordedVaultProvider() fx.Annotated {
	return fx.Annotated{
		Group: "vault",
		Target: func(recorder *recorder.Recorder) vault.Options {
			return func(cfg *vault.ClientConfig) error {
				recorder.SetRealTransport(cfg.HttpClient.Transport)
				cfg.HttpClient.Transport = recorder
				return nil
			}
		},
	}
}

type VaultTestDi struct {
	fx.In
	Manager     certs.Manager
	VaultClient *vault.Client
}

/*************************
	Tests
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		ittest.PackageHttpRecordingMode(),
//	)
//}

// This test assumes your vault has PKI backend enabled (i.e. vault secrets enable pki)
func TestVaultProvider(t *testing.T) {
	di := &VaultTestDi{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t, ittest.DisableHttpRecordOrdering()),
		apptest.WithDI(di),
		apptest.WithModules(vaultinit.Module, vaultcerts.Module),
		apptest.WithFxOptions(
			fx.Provide(RecordedVaultProvider()),
			fx.Provide(ProvideTestManager, BindTestProperties),
		),
		test.SubTestSetup(SubTestSetupCleanupTempDir()),
		test.SubTestSetup(SubTestSetupSubmitCA(di)),
		test.GomegaSubTest(SubTestVaultTLSConfig(di), "SubTestVaultTLSConfig"),
		test.GomegaSubTest(SubTestVaultCertFiles(di), "SubTestVaultCertFiles"),
		test.GomegaSubTest(SubTestVaultRenewal(di), "SubTestVaultRenewal"),
	)
}

func SubTestSetupSubmitCA(di *VaultTestDi) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		data, err := os.ReadFile("testdata/ca-bundle-test.json") //this file has the ca bundle matching ca-cert-test.pem
		if err != nil {
			return ctx, err
		}
		r := di.VaultClient.NewRequest("POST", "/v1/pki/config/ca")
		r.BodyBytes = data
		_, err = di.VaultClient.Logical(ctx).WriteBytesWithContext(ctx, "/pki/config/ca", data)
		return ctx, err
	}
}

func SubTestSetupCleanupTempDir() test.SetupFunc {
	var once sync.Once
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		once.Do(func() {
			_ = os.RemoveAll("testdata/.tmp/certs")
		})
		return ctx, nil
	}
}

func SubTestVaultTLSConfig(di *VaultTestDi) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// For recorded HTTP, the certificate should be valid for very long time for repeated tests
		p := vaultcerts.SourceProperties{
			Path:             "pki/",
			Role:             "localhost",
			CN:               "localhost",
			TTL:              MaxCertificateTTL(), // many years
			MinRenewInterval: utils.Duration(2 * time.Second),
		}

		tlsSrc, err := di.Manager.Source(ctx, certs.WithType(certs.SourceVault, p))
		g.Expect(err).To(Succeed())

		tlsCfg, err := tlsSrc.TLSConfig(ctx)
		g.Expect(err).To(Succeed())
		g.Expect(tlsCfg.RootCAs).ToNot(BeNil())
		g.Expect(tlsCfg.RootCAs.Subjects()).ToNot(BeEmpty())
		g.Expect(tlsCfg.RootCAs).ToNot(BeNil())

		// try with the ca that the cert is signed with
		// the signature scheme and version is captured from server that uses tls connection.
		certReqInfo := NewTestCertificateRequestInfo(tlsCfg)
		clientCert, err := tlsCfg.GetClientCertificate(certReqInfo)
		g.Expect(err).To(Succeed())
		g.Expect(clientCert).NotTo(BeNil())
		g.Expect(len(clientCert.Certificate)).To(Equal(1))

		parsedCert, err := x509.ParseCertificate(clientCert.Certificate[0])
		g.Expect(err).To(Succeed())
		//expect the cert to be valid
		g.Expect(time.Now().After(parsedCert.NotBefore)).To(BeTrue())
		g.Expect(time.Now().Before(parsedCert.NotAfter)).To(BeTrue())

		//try with a different ca, and expect no cert is returned
		anotherCa, err := os.ReadFile("testdata/ca-cert-test-2")
		g.Expect(err).To(Succeed())
		anotherCaPool := x509.NewCertPool()
		anotherCaPool.AppendCertsFromPEM(anotherCa)

		certReqInfo.AcceptableCAs = anotherCaPool.Subjects()
		clientCert, err = tlsCfg.GetClientCertificate(certReqInfo)
		g.Expect(err).To(Succeed())
		g.Expect(len(clientCert.Certificate)).To(Equal(0))
	}
}

func SubTestVaultCertFiles(di *VaultTestDi) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// For recorded HTTP, the certificate should be valid for very long time for repeated tests
		p := vaultcerts.SourceProperties{
			Path:             "pki/",
			Role:             "localhost",
			CN:               "localhost",
			TTL:              MaxCertificateTTL(), // many years
			MinRenewInterval: utils.Duration(2 * time.Second),
			CachePath:        "testdata/.tmp/certs",
		}

		tlsSrc, err := di.Manager.Source(ctx, certs.WithType(certs.SourceVault, p))
		g.Expect(err).To(Succeed())

		tlsFiles, err := tlsSrc.Files(ctx)
		g.Expect(err).To(Succeed())
		const fileRegexTmpl = `testdata/\.tmp/certs/vault/localhost-localhost-[0-9]+-%s\.pem`
		g.Expect(tlsFiles.RootCAPaths).To(ContainElement(MatchRegexp(fileRegexTmpl, "ca")))
		AssertFilesExist(g, tlsFiles.RootCAPaths)
		g.Expect(tlsFiles.CertificatePath).To(MatchRegexp(fileRegexTmpl, "cert"))
		AssertFileExists(g, tlsFiles.CertificatePath)
		g.Expect(tlsFiles.PrivateKeyPath).To(MatchRegexp(fileRegexTmpl, "key"))
		AssertFileExists(g, tlsFiles.PrivateKeyPath)
		g.Expect(tlsFiles.PrivateKeyPassphrase).To(Equal(""))
	}
}

func SubTestVaultRenewal(di *VaultTestDi) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		var ttl = 5 * time.Second // short
		p := vaultcerts.SourceProperties{
			Path:             "pki/",
			Role:             "localhost",
			CN:               "localhost",
			TTL:              utils.Duration(ttl),
			MinRenewInterval: utils.Duration(1 * time.Second),
			CachePath:        "testdata/.tmp/certs",
		}

		tlsSrc, err := di.Manager.Source(ctx, certs.WithType(certs.SourceVault, p))
		g.Expect(err).To(Succeed())

		// Note: In this test case, the certificate have short TTL, we don't check certificate's validity due to HTTP playback.
		// 		 Instead, we focus on certificate been renewed (different after delay)
		tlsFiles, err := tlsSrc.Files(ctx)
		g.Expect(err).To(Succeed())
		beforeCert := LoadFile(g, tlsFiles.CertificatePath)
		beforeKey := LoadFile(g, tlsFiles.PrivateKeyPath)

		//Sleep for more than half of the TTL, so the original cert is renewed
		//we expect the renew process to kick in and got a new cert
		time.Sleep(ttl-time.Second)
		tlsFiles, err = tlsSrc.Files(ctx)
		g.Expect(err).To(Succeed())

		// verify new cert is different from the old one
		afterCert := LoadFile(g, tlsFiles.CertificatePath)
		g.Expect(afterCert).ToNot(BeEquivalentTo(beforeCert), "new cert should be issued")
		afterKey := LoadFile(g, tlsFiles.PrivateKeyPath)
		g.Expect(afterKey).ToNot(BeEquivalentTo(beforeKey), "new key should be issued")
	}
}

/*************************
	Helpers
 *************************/

func MaxCertificateTTL() utils.Duration {
	return utils.Duration(TestCAExpiration.Sub(time.Now()) - time.Minute)
}

func NewTestCertificateRequestInfo(tlsCfg *tls.Config) *tls.CertificateRequestInfo {
	info := TestCertReqInfoTmpl
	info.AcceptableCAs = tlsCfg.RootCAs.Subjects()
	return &info
}

func AssertFilesExist(g *WithT, paths []string) {
	for _, path := range paths {
		AssertFileExists(g, path)
	}
}

func AssertFileExists(g *WithT, path string) {
	LoadFile(g, path)
}

func LoadFile(g *WithT, path string) []byte {
	data, e := os.ReadFile(path)
	g.Expect(e).To(Succeed(), "reading file '%s' should not fail", path)
	g.Expect(data).ToNot(BeEmpty(), "file '%s' should not be empty", path)
	return data
}
