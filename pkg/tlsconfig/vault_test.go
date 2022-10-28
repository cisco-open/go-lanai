package tlsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	vaultinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"errors"
	"fmt"
	"go.uber.org/fx"
	"net/http"
	"os"
	"testing"
	"time"
)
import . "github.com/onsi/gomega"

type VaultTestDi struct {
	fx.In
	ProviderFactory *ProviderFactory
	VaultClient *vault.Client
}

// This test assumes your vault has PKI backend enabled (i.e. vault secrets enable pki)
func TestVaultProvider(t *testing.T) {
	t.Skipf("skipped because this test requires real vault server")
	di := &VaultTestDi{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		apptest.WithModules(Module, vaultinit.Module),
		test.SubTestSetup(SubTestSetupSubmitCA(di)),
		test.GomegaSubTest(SubTestVaultProvider(di), "SubTestVaultProvider"),
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
		resp, err := di.VaultClient.RawRequestWithContext(ctx, r)
		if err != nil {
			return ctx, err
		}
		if resp.StatusCode != http.StatusNoContent {
			return ctx, errors.New(fmt.Sprintf("could not submit ca to vault, vault responded with %s", resp.Status))
		}

		return ctx, err
	}
}

func SubTestVaultProvider(di *VaultTestDi) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		p := Properties{
			Type:             "vault",
			Path:             "pki/",
			Role:             "localhost",
			CN:               "localhost",
			Ttl:              "10s",
			MinRenewInterval: "2s",
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

		parsedCert, err := x509.ParseCertificate(clientCert.Certificate[0])
		g.Expect(err).NotTo(HaveOccurred())
		//expect the cert to be valid
		g.Expect(time.Now().After(parsedCert.NotBefore)).To(BeTrue())
		g.Expect(time.Now().Before(parsedCert.NotAfter)).To(BeTrue())
		// because we specified our ttl to be 10s, we expect the cert to be expired after 10 seconds
		g.Expect(time.Now().Add(11 * time.Second).After(parsedCert.NotAfter)).To(BeTrue())

		//Sleep for 15 seconds, so the original cert is expired
		//we expect the renew process to kick in and got a new cert
		time.Sleep(13 * time.Second)
		clientCert, err = getClientCert(certReqInfo)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(clientCert).NotTo(BeNil())
		g.Expect(len(clientCert.Certificate)).To(Equal(1))

		parsedCert, err = x509.ParseCertificate(clientCert.Certificate[0])
		g.Expect(err).NotTo(HaveOccurred())
		//we expect the cert to be valid
		g.Expect(time.Now().After(parsedCert.NotBefore)).To(BeTrue())
		g.Expect(time.Now().Before(parsedCert.NotAfter)).To(BeTrue())

		//try with a different ca, and expect no cert is returned
		anotherCa, err := os.ReadFile("testdata/ca-cert-test-2")
		g.Expect(err).NotTo(HaveOccurred())
		anotherCaPool := x509.NewCertPool()
		anotherCaPool.AppendCertsFromPEM(anotherCa)

		certReqInfo.AcceptableCAs = anotherCaPool.Subjects()
		clientCert, err = getClientCert(certReqInfo)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(len(clientCert.Certificate)).To(Equal(0))
	}
}
