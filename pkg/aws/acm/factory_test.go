package acm_test

import (
    "context"
    awsconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws"
    acmclient "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws/acm"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/acm"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "gopkg.in/dnaeon/go-vcr.v3/recorder"
    "os"
    "testing"
)

/*************************
	Test Setup
 *************************/

const (
    TestCertPath   = `testdata/test.crt`
    TestKeyPath    = `testdata/test.key`
)

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
    AcmClient *acm.Client
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
        apptest.WithModules(awsconfig.Module, acmclient.Module),
        apptest.WithFxOptions(
            fx.Provide(awsconfig.FxCustomizerProvider(CustomizeAwsClient)),
        ),
        test.GomegaSubTest(SubTestImportCertificate(di), "ConfigWithStaticCredentials"),
    )
}

func SubTestImportCertificate(di *AcmTestDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        input := &acm.ImportCertificateInput{
            Certificate: MustLoadFile(g, TestCertPath),
            PrivateKey:  MustLoadFile(g, TestKeyPath),
        }
        output, e := di.AcmClient.ImportCertificate(ctx, input)
        g.Expect(e).To(Succeed(), "importing certificate should not fail")
        g.Expect(output).ToNot(BeNil(), "output from client should not be nil")
    }
}

/*************************
	Helpers
 *************************/

func MustLoadFile(g *WithT, path string) []byte {
    data, e := os.ReadFile(path)
    g.Expect(e).To(Succeed(), "reading file '%s' should not fail", path)
    g.Expect(data).ToNot(BeEmpty(), "file '%s' should not be empty", path)
    return data
}
