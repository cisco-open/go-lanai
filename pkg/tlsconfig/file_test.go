package tlsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"go.uber.org/fx"
	"io/ioutil"
	"testing"
)
import . "github.com/onsi/gomega"


type FileTestDi struct {
	fx.In
	ProviderFactory *ProviderFactory
}


func TestFileProvider(t *testing.T) {
	di := &FileTestDi{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		apptest.WithModules(Module),
		test.GomegaSubTest(SubTestFileProvider(di), "SubTestFileProvider"),
	)
}

func SubTestFileProvider(di *FileTestDi) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		p := Properties{
			Type: "file",
			CaCertFile: "testdata/ca-cert-test.pem",
			CertFile: "testdata/client-cert-signed-test.pem",
			KeyFile: "testdata/client-key-test.pem",
			KeyPass: "foobar",
		}

		provider, err := di.ProviderFactory.GetProvider(p)
		g.Expect(err).NotTo(HaveOccurred())

		caPool, err := provider.RootCAs(ctx)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(len(caPool.Subjects())).To(Equal(1))

		getClientCert, err := provider.GetClientCertificate(ctx)
		g.Expect(err).NotTo(HaveOccurred())

		//try with the ca that the cert is signed with
		// the signature scheme and version is captured from a kafka broker that uses tls connection.
		certReqInfo := &tls.CertificateRequestInfo{
			AcceptableCAs: caPool.Subjects(),
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
		clientCert, err := getClientCert(certReqInfo)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(clientCert).NotTo(BeNil())
		g.Expect(len(clientCert.Certificate)).To(Equal(1))

		//try with a different ca, and expect no cert is returned
		anotherCa, err := ioutil.ReadFile("testdata/ca-cert-test-2")
		g.Expect(err).NotTo(HaveOccurred())
		anotherCaPool := x509.NewCertPool()
		anotherCaPool.AppendCertsFromPEM(anotherCa)

		certReqInfo.AcceptableCAs = anotherCaPool.Subjects()
		clientCert, err = getClientCert(certReqInfo)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(len(clientCert.Certificate)).To(Equal(0))
	}
}