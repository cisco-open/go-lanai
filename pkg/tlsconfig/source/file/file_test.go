package filecerts_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	filecerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source/file"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"go.uber.org/fx"
	"os"
	"testing"
)
import . "github.com/onsi/gomega"

type mgrDI struct {
	fx.In
	AppCfg bootstrap.ApplicationConfig
	Factories []tlsconfig.SourceFactory `group:"certs"`
}

func ProvideTestManager(di mgrDI) (tlsconfig.Manager, tlsconfig.Registrar) {
	reg := tlsconfig.NewDefaultManager(di.AppCfg.Bind)
	for _, f := range di.Factories {
		if f != nil {
			reg.MustRegister(f)
		}
	}
	return reg, reg
}

type FileTestDi struct {
	fx.In
	CertsManager tlsconfig.Manager
}

func TestFileCertificateSource(t *testing.T) {
	di := &FileTestDi{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(tlsconfig.BindProperties, ProvideTestManager, filecerts.FxProvider()),
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

		tlsSrc, err := di.CertsManager.Source(ctx, tlsconfig.WithType(tlsconfig.SourceFile, p))
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

		tlsSrc, err := di.CertsManager.Source(ctx, tlsconfig.WithType(tlsconfig.SourceFile, p))
		g.Expect(err).NotTo(HaveOccurred())

		tlsFiles, err := tlsSrc.Files(ctx)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(tlsFiles.RootCAPaths).To(ContainElement(ContainSubstring("testdata/ca-cert-test.pem")))
		g.Expect(tlsFiles.CertificatePath).To(ContainSubstring("testdata/client-cert-signed-test.pem"))
		g.Expect(tlsFiles.PrivateKeyPath).To(ContainSubstring("testdata/client-key-test.pem"))
		g.Expect(tlsFiles.PrivateKeyPassphrase).To(Equal("foobar"))
	}
}
