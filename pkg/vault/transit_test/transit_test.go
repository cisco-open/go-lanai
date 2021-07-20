package transit_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	vaultinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"encoding/json"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

var (
	testKid1 = "d3803a9e-f2f2-4960-bdb1-aeec92d88ca4"
	testKid2 = "3100e6b7-eb62-4676-9bf4-391aba1f2fae"
	testKidIncorrect = "e668ce1d-e2fe-42d2-a1e2-9b553555378f"
	plaintextData = map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
)

/*************************
	Test Cases
 *************************/

type transitDI struct {
	fx.In
	Client *vault.Client
}

func TestTransitEngineWithRealVault(t *testing.T) {
	t.Skipf("skipped because this test requires real vault server")
	di := transitDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(vaultinit.Module),
		apptest.WithDI(&di),
		test.SubTestSetup(SubTestSetupCreateKey(&di)),
		test.GomegaSubTest(SubTestEncryption(&di, testKid1), "EncryptionWithFirstKey"),
		test.GomegaSubTest(SubTestEncryption(&di, testKid2), "EncryptionWithSecondKey"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSetupCreateKey(di *transitDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		engine := newTestEngine(di)
		for _, kid := range []string{testKid1, testKid2, testKidIncorrect} {
			e := engine.PrepareKey(ctx, kid)
			g := gomega.NewWithT(t)
			g.Expect(e).To(Succeed(), "PrepareKey shouldn't return error")
		}
		return ctx, nil
	}
}

func SubTestEncryption(di *transitDI, kid string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// encrypt
		plaintext, e := json.Marshal(plaintextData)
		g.Expect(e).To(Succeed(), "json marshal should succeed")

		engine := newTestEngine(di)
		v, e := engine.Encrypt(ctx, kid, plaintext)
		g.Expect(e).To(Succeed(), "Encrypt should succeed")
		g.Expect(v).To(Not(BeEmpty()), "encrypted data shouldn't be empty")
		str := string(v)
		g.Expect(str).To(HavePrefix("vault:v1:"), "encrypted data should have correct prefix")

		// decrypt with correct key
		decrypted, e := engine.Decrypt(ctx, kid, v)
		g.Expect(e).To(Succeed(), "Decrypt should succeed")
		g.Expect(decrypted).To(Not(BeEmpty()), "decrypted data shouldn't be empty")
		decryptedData := map[string]interface{}{}
		e = json.Unmarshal(decrypted, &decryptedData)
		g.Expect(e).To(Succeed(), "json unmarshal should succeed")
		g.Expect(decryptedData).To(Equal(plaintextData), "decrypted data should be same as original")

		// decrypt with incorrect key
		_, e = engine.Decrypt(ctx, testKidIncorrect, v)
		g.Expect(e).To(Not(Succeed()), "Decrypt with incorrect key should fail")
	}
}

func newTestEngine(di *transitDI) vault.TransitEngine {
	return vault.NewTransitEngine(di.Client, func(opt *vault.KeyOption) {
		opt.Exportable = true
		opt.AllowPlaintextBackup = true
	})
}
