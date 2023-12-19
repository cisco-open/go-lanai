package vault_test

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "gopkg.in/dnaeon/go-vcr.v3/recorder"
    "testing"
)

/*************************
	Tests Setup Helpers
 *************************/

func RecordedVaultProvider() fx.Annotated {
    return fx.Annotated{
        Group: "vault",
        Target: VaultWithRecorder,
    }
}

func VaultWithRecorder(recorder *recorder.Recorder) vault.Options {
    return func(cfg *vault.ClientConfig) error {
        recorder.SetRealTransport(cfg.HttpClient.Transport)
        cfg.HttpClient.Transport = recorder
        return nil
    }
}

func NewTestClient(g *gomega.WithT, props vault.ConnectionProperties, recorder *recorder.Recorder) (*vault.Client,) {
    client, e := vault.New(vault.WithProperties(props), VaultWithRecorder(recorder))
    g.Expect(e).To(Succeed(), "create vault client should not fail")
    return client
}

func SetupTestConvertV1HttpRecords(src, dest string) test.SetupFunc {
    return func(ctx context.Context, t *testing.T) (context.Context, error) {
        e := ittest.ConvertCassetteFileV1toV2(src, dest)
        return ctx, e
    }
}