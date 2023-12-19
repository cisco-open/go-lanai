package vault_test

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "testing"
)

/*************************
	Test Setup
 *************************/

var TestK8sAuthProperties = vault.ConnectionProperties{
    Host:           "127.0.0.1",
    Port:           8200,
    Scheme:         "http",
    Authentication: "kubernetes",
    Kubernetes:     vault.KubernetesConfig{
        JWTPath: "testdata/k8s-jwt-valid.txt",
        Role:    "devweb-app",
    },
}

/*************************
	Tests
 *************************/

type TestK8sDI struct {
    fx.In
    ittest.RecorderDI
}

func TestAuthenticateWithK8s(t *testing.T) {
    di := TestK8sDI{}
    test.RunTest(context.Background(), t,
        apptest.Bootstrap(),
        test.Setup(SetupTestConvertV1HttpRecords("testdata/authentication_kubernetes/successful_client.yaml", "testdata/TestAuthenticateWithK8s.httpvcr.yaml")),
        ittest.WithHttpPlayback(t),
        apptest.WithFxOptions(
            fx.Provide(RecordedVaultProvider()),
        ),
        apptest.WithDI(&di),
        test.GomegaSubTest(SubTestSuccessfulK8sAuth(&di), "TestSuccessfulK8sAuth"),
    )
}

func TestFailedAuthenticateWithK8s(t *testing.T) {
    di := TestK8sDI{}
    test.RunTest(context.Background(), t,
        apptest.Bootstrap(),
        //test.Setup(SetupTestConvertV1HttpRecords("testdata/authentication_kubernetes/invalid_role.yaml", "testdata/TestFailedAuthenticateWithK8s.httpvcr.yaml")),
        ittest.WithHttpPlayback(t),
        apptest.WithFxOptions(
            fx.Provide(RecordedVaultProvider()),
        ),
        apptest.WithDI(&di),
        test.GomegaSubTest(SubTestFailedK8sAuth(&di), "TestFailedK8sAuth"),
    )
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSuccessfulK8sAuth(di *TestK8sDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        p := TestK8sAuthProperties
        p.Kubernetes.Role = "devweb-app"
        client, e := vault.New(vault.WithProperties(p), VaultWithRecorder(di.Recorder))
        g.Expect(e).To(Succeed(), "client with k8s auth should not fail")
        g.Expect(client).ToNot(BeNil(), "client with k8s auth should not be nil")
        token := client.Token()
        g.Expect(token).ToNot(BeEmpty(), "client's token should not be empty")
    }
}

func SubTestFailedK8sAuth(di *TestK8sDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        p := TestK8sAuthProperties
        p.Kubernetes.Role = "invalid-role"
        client, e := vault.New(vault.WithProperties(p), VaultWithRecorder(di.Recorder))
        g.Expect(e).To(Succeed(), "client with k8s auth should not fail")
        g.Expect(client).ToNot(BeNil(), "client with k8s auth should not be nil")
        token := client.Token()
        g.Expect(token).To(BeEmpty(), "client's token should be empty")
    }
}

/*************************
	Helpers
 *************************/
