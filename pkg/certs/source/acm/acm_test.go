package acmcerts_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	awsconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	acmcerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source/acm"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

/*************************
	Test Setup
 *************************/

const (
	CtxKeyARNValid      = "arn:valid"
	CtxKeyARNShortLived = "arn:short"
)

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

func AwsHTTPVCROptions() ittest.HTTPVCROptions {
	return ittest.HttpRecordMatching(ittest.FuzzyHeaders("Amz-Sdk-Invocation-Id", "X-Amz-Date", "User-Agent"))
}

type RecordedAwsDI struct {
	fx.In
	Recorder *recorder.Recorder
}

func CustomizeAwsClient(di RecordedAwsDI) config.LoadOptionsFunc {
	return func(opt *config.LoadOptions) error {
		opt.HTTPClient = di.Recorder.GetDefaultClient()
		return nil
	}
}

type AcmTestDI struct {
	fx.In
	Manager         certs.Manager
	AwsConfigLoader awsconfig.ConfigLoader
}

/*************************
	Tests
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		ittest.PackageHttpRecordingMode(),
//	)
//}

// This test assumes you are running LocatStack (https://docs.localstack.cloud/user-guide/aws/feature-coverage/) at localhost
func TestDefaultClient(t *testing.T) {
	di := &AcmTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t, AwsHTTPVCROptions()),
		apptest.WithDI(di),
		apptest.WithModules(awsconfig.Module, acmcerts.Module),
		apptest.WithFxOptions(
			fx.Provide(ProvideTestManager, BindTestProperties),
			fx.Provide(awsconfig.FxCustomizerProvider(CustomizeAwsClient)),
		),
		test.SubTestSetup(SubTestSetupCleanupTempDir()),
		test.SubTestSetup(SubTestSetupImportCerts(di)),
		test.GomegaSubTest(SubTestAcmTLSConfig(di), "TestAcmTLSConfig"),
		test.GomegaSubTest(SubTestAcmCertFiles(di), "TestAcmCertFiles"),
		test.GomegaSubTest(SubTestAcmRenewal(di), "TestAcmRenewal"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestSetupImportCerts(di *AcmTestDI) test.SetupFunc {
	var once sync.Once
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		var err error
		once.Do(func() {
			g := gomega.NewWithT(t)
			cfg, e := di.AwsConfigLoader.Load(ctx)
			g.Expect(e).To(Succeed(), "load AWS config should not fail")
			client := acm.NewFromConfig(cfg)

			var arn string
			// the valid one
			arn, err = ImportCertificate(ctx, g, client, "testdata/test-client.crt", "testdata/test-client.key", "testdata/test-ca.crt")
			if err != nil {
				return
			}
			ctx = context.WithValue(ctx, CtxKeyARNValid, arn)
			// for renewal, we cannot test renewal because AWS
			//arn, err = ImportCertificate(ctx, g, client,
			//	"testdata/test-client-short.crt", "testdata/test-client-short.key", "testdata/test-ca-short.crt")
			//if err != nil {
			//	return
			//}
			//ctx = context.WithValue(ctx, CtxKeyARNShortLived, arn)
		})
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

func SubTestAcmTLSConfig(di *AcmTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// For recorded HTTP, the certificate should be valid for very long time for repeated tests
		p := acmcerts.SourceProperties{
			ARN:              ctx.Value(CtxKeyARNValid).(string),
			Passphrase:       "doesn't matter",
			MinRenewInterval: utils.Duration(2 * time.Second),
		}

		tlsSrc, err := di.Manager.Source(ctx, certs.WithType(certs.SourceACM, p))
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

func SubTestAcmCertFiles(di *AcmTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// For recorded HTTP, the certificate should be valid for very long time for repeated tests
		arn := ctx.Value(CtxKeyARNValid).(string)
		p := acmcerts.SourceProperties{
			ARN:              arn,
			Passphrase:       "doesn't matter",
			MinRenewInterval: utils.Duration(2 * time.Second),
			CachePath:        "testdata/.tmp/certs",
		}

		tlsSrc, err := di.Manager.Source(ctx, certs.WithType(certs.SourceACM, p))
		g.Expect(err).To(Succeed())

		tlsFiles, err := tlsSrc.Files(ctx)
		g.Expect(err).To(Succeed())

		expectedId := arn[strings.LastIndex(arn, "/")+1:]
		var fileRegexTmpl = fmt.Sprintf(`testdata/\.tmp/certs/acm/certificate-%s-[0-9]+-%%s\.pem`, expectedId)
		g.Expect(tlsFiles.RootCAPaths).To(ContainElement(MatchRegexp(fileRegexTmpl, "ca")))
		AssertFilesExist(g, tlsFiles.RootCAPaths)
		g.Expect(tlsFiles.CertificatePath).To(MatchRegexp(fileRegexTmpl, "cert"))
		AssertFileExists(g, tlsFiles.CertificatePath)
		g.Expect(tlsFiles.PrivateKeyPath).To(MatchRegexp(fileRegexTmpl, "key"))
		AssertFileExists(g, tlsFiles.PrivateKeyPath)
		g.Expect(tlsFiles.PrivateKeyPassphrase).To(Equal(""))
	}
}

func SubTestAcmRenewal(di *AcmTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		t.Skipf("Renewal test is not possible for imported certificates")
		var ttl = 5 * time.Second // short
		p := acmcerts.SourceProperties{
			ARN:              ctx.Value(CtxKeyARNShortLived).(string),
			Passphrase:       "doesn't matter",
			MinRenewInterval: utils.Duration(2 * time.Second),
			CachePath:        "testdata/.tmp/certs",
		}

		tlsSrc, err := di.Manager.Source(ctx, certs.WithType(certs.SourceACM, p))
		g.Expect(err).To(Succeed())

		// Note: In this test case, the certificate have short TTL, we don't check certificate's validity due to HTTP playback.
		// 		 Instead, we focus on certificate been renewed (different after delay)
		tlsFiles, err := tlsSrc.Files(ctx)
		g.Expect(err).To(Succeed())
		beforeCert := LoadFile(g, tlsFiles.CertificatePath)
		beforeKey := LoadFile(g, tlsFiles.PrivateKeyPath)

		//Sleep for more than half of the TTL, so the original cert is renewed
		//we expect the renew process to kick in and got a new cert
		time.Sleep(ttl - time.Second)
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

func ImportCertificate(ctx context.Context, g *gomega.WithT, acmClient *acm.Client, certPath, keyPath, caPath string) (string, error) {
	certBytes := LoadFile(g, certPath)
	keyBytes := LoadFile(g, keyPath)
	caBytes := LoadFile(g, caPath)
	output, e := acmClient.ImportCertificate(ctx, &acm.ImportCertificateInput{
		Certificate:      certBytes,
		PrivateKey:       keyBytes,
		CertificateChain: caBytes,
		Tags: []types.Tag{{
			Key:   utils.ToPtr("name"),
			Value: utils.ToPtr("test"),
		}},
	})
	if e != nil {
		return "", e
	}
	if output.CertificateArn == nil {
		return "", fmt.Errorf("ImportCertificate didn't return any ARN")
	}
	return *output.CertificateArn, nil
}
