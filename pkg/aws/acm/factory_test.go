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

/*************************
	Sub Test
 *************************/

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
