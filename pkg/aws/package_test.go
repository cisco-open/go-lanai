package aws_test

import (
    "context"
    awsconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
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
	TargetRegion   = `us-east-1`
	TargetEndpoint = `http://localhost:4566`
	STSJwtPath     = `testdata/sts_jwt.txt`
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

type AwsTestDI struct {
	fx.In
	ConfigLoader awsconfig.ConfigLoader
	Recorder     *recorder.Recorder
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
func TestAWSConfigLoading(t *testing.T) {
	di := &AwsTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t, AwsHTTPVCROptions()),
		apptest.WithDI(di),
		apptest.WithModules(awsconfig.Module),
		apptest.WithFxOptions(
			fx.Provide(awsconfig.FxCustomizerProvider(CustomizeAwsClient)),
		),
		test.SubTestSetup(SubTestSetupSTS(di)),
		test.GomegaSubTest(SubTestConfigWithStaticCredentials(di), "ConfigWithStaticCredentials"),
		test.GomegaSubTest(SubTestConfigWithEnvCredentials(di), "ConfigWithEnvCredentials"),
		test.GomegaSubTest(SubTestConfigWithSTSCredentials(di), "ConfigWithSTSCredentials"),
	)
}

func SubTestSetupSTS(di *AwsTestDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		_, e := config.LoadDefaultConfig(ctx,
			config.WithRegion(TargetRegion),
			config.WithEndpointResolverWithOptions(NewTestEndpointResolver(TargetEndpoint)),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "static_auth")),
			func(opt *config.LoadOptions) error {
				opt.HTTPClient = di.Recorder.GetDefaultClient()
				return nil
			})
		g.Expect(e).To(Succeed(), "load config should not fail")
		return ctx, nil
	}
}

func SubTestConfigWithStaticCredentials(di *AwsTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		props := awsconfig.Properties{
			Region:   TargetRegion,
			Endpoint: TargetEndpoint,
			Credentials: awsconfig.Credentials{
				Type:   awsconfig.CredentialsTypeStatic,
				Id:     "test",
				Secret: "test",
			},
		}
		loader := ConfigLoadWithProperties(di.ConfigLoader, props)
		cfg, e := loader.Load(ctx)
		g.Expect(e).To(Succeed(), "load config should not fail")
		AssertAwsConfig(ctx, g, cfg)
	}
}

func SubTestConfigWithEnvCredentials(di *AwsTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		props := awsconfig.Properties{
			Region:   TargetRegion,
			Endpoint: TargetEndpoint,
			Credentials: awsconfig.Credentials{
				Type: "none",
			},
		}
		loader := ConfigLoadWithProperties(di.ConfigLoader, props)
		cfg, e := loader.Load(ctx)
		g.Expect(e).To(Succeed(), "load config should not fail")
		AssertAwsConfig(ctx, g, cfg)
	}
}

func SubTestConfigWithSTSCredentials(di *AwsTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		props := awsconfig.Properties{
			Region:   TargetRegion,
			Endpoint: TargetEndpoint,
			Credentials: awsconfig.Credentials{
				Type:            awsconfig.CredentialsTypeSTS,
				TokenFile:       STSJwtPath,
				RoleSessionName: "test_session",
			},
		}
		loader := ConfigLoadWithProperties(di.ConfigLoader, props)
		cfg, e := loader.Load(ctx)
		g.Expect(e).To(Succeed(), "load config should not fail")
		AssertAwsConfig(ctx, g, cfg)
	}
}

/*************************
	Helpers
 *************************/

func ConfigLoadWithProperties(loader awsconfig.ConfigLoader, p awsconfig.Properties) awsconfig.ConfigLoader {
	if v, ok := loader.(*awsconfig.PropertiesBasedConfigLoader); ok {
		newLoader := *v
		newLoader.Properties = &p
		return &newLoader
	}
	return loader
}

func AssertAwsConfig(ctx context.Context, g *gomega.WithT, cfg aws.Config) {
	g.Expect(cfg).ToNot(BeZero(), "loaded config should not be empty")

	client := acm.NewFromConfig(cfg)
	g.Expect(client).ToNot(BeNil(), "client using loaded config should not be nil")

	input := &acm.ImportCertificateInput{
		Certificate: MustLoadFile(g, TestCertPath),
		PrivateKey:  MustLoadFile(g, TestKeyPath),
	}
	output, e := client.ImportCertificate(ctx, input)
	g.Expect(e).To(Succeed(), "call to AWS should not fail")
	g.Expect(output).ToNot(BeNil(), "output from AWS client should not be nil")
}

func MustLoadFile(g *WithT, path string) []byte {
	data, e := os.ReadFile(path)
	g.Expect(e).To(Succeed(), "reading file '%s' should not fail", path)
	g.Expect(data).ToNot(BeEmpty(), "file '%s' should not be empty", path)
	return data
}

func NewTestEndpointResolver(url string) aws.EndpointResolverWithOptions {
	return aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: url}, nil
		},
	)
}
