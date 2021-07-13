package datacrypto

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	vaultinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

var (
	testKid    = "d3803a9e-f2f2-4960-bdb1-aeec92d88ca4"
	incorrectTestKid = "3100e6b7-eb62-4676-9bf4-391aba1f2fae"
)

/*************************
	Test Cases
 *************************/

type transitDI struct {
	fx.In
	Client *vault.Client
}

func TestVaultEncryptorWithRealVault(t *testing.T) {
	t.Skipf("skipped because this test requires real vault server")
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
	strValue := "this is a string"
	arrValue := []interface{}{"value1", 2.0}

	di := transitDI{}
	newVeDI := veDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(vaultinit.Module),
		apptest.WithDI(&di, &newVeDI),
		test.SubTestSetup(SubTestSetupCreateKey(&di)),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V1, testKid, mapValue), "VaultMapV1"),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V2, testKid, mapValue), "VaultMapV2"),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V1, testKid, strValue), "VaultStringV1"),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V2, testKid, strValue), "VaultStringV2"),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V1, testKid, arrValue), "VaultSliceV1"),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V2, testKid, arrValue), "VaultSliceV2"),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V1, testKid, nil), "VaultNilV1"),
		test.GomegaSubTest(SubTestVaultEncryptor(&newVeDI, V2, testKid, nil), "VaultNilV2"),
	)
}

func TestVaultFailedEncrypt(t *testing.T) {
	enc := newVaultEncryptor(veDI{}).(*vaultEncryptor)
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestVaultFailedEncryption(enc, Version(-1), AlgVault, ErrUnsupportedVersion), "InvalidVersion"),
		test.GomegaSubTest(SubTestVaultFailedEncryption(enc, V1, AlgPlain, ErrUnsupportedAlgorithm), "V1UnsupportedAlg"),
		test.GomegaSubTest(SubTestVaultFailedEncryption(enc, V2, AlgPlain, ErrUnsupportedAlgorithm), "V2UnsupportedAlg"),
		test.GomegaSubTest(SubTestVaultEncryptWithoutKid(enc), "EncryptWithoutKeyID"),
	)
}

func TestVaultFailedDecrypt(t *testing.T) {
	enc := newVaultEncryptor(veDI{}).(*vaultEncryptor)
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestVaultFailedDecryption(enc, Version(-1), AlgVault, ErrUnsupportedVersion), "InvalidVersion"),
		test.GomegaSubTest(SubTestVaultFailedDecryption(enc, V1, AlgPlain, ErrUnsupportedAlgorithm), "V1UnsupportedAlg"),
		test.GomegaSubTest(SubTestVaultFailedDecryption(enc, V2, AlgPlain, ErrUnsupportedAlgorithm), "V2UnsupportedAlg"),
		test.GomegaSubTest(SubTestVaultDecryptWithoutKid(enc), "DecryptWithoutKeyID"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSetupCreateKey(di *transitDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		engine := newTestEngine(di)
		for _, kid := range []string{testKid, incorrectTestKid} {
			e := engine.PrepareKey(ctx, kid)
			g := gomega.NewWithT(t)
			g.Expect(e).To(Succeed(), "PrepareKey shouldn't return error")
		}
		return ctx, nil
	}
}

func SubTestVaultEncryptor(di *veDI, ver Version, uuidStr string, v interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		enc := newVaultEncryptor(*di)
		kid := uuid.Must(uuid.Parse(uuidStr))
		raw := EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   AlgVault,
		}

		// encrypt
		e := enc.Encrypt(ctx, v, &raw)
		g.Expect(e).To(Succeed(), "Encrypt shouldn't return error")
		g.Expect(raw.Ver).To(BeIdenticalTo(V2), "encrypted data should be V2")
		g.Expect(raw.Raw).To(BeAssignableToTypeOf(""), "encrypted raw should be a string")

		// serialize
		bytes, e := json.Marshal(raw)
		g.Expect(e).To(Succeed(), "JSON marshal of raw data shouldn't return error")

		// deserialize
		parsed := EncryptedRaw{}
		e = json.Unmarshal(bytes, &parsed)
		g.Expect(e).To(Succeed(), "JSON unmarshal of raw data shouldn't return error")
		g.Expect(parsed.Ver).To(BeIdenticalTo(V2), "unmarshalled data should be V2")
		g.Expect(parsed.KeyID).To(Equal(kid), "unmarshalled KeyID should be correct")
		g.Expect(parsed.Alg).To(BeIdenticalTo(AlgVault), "unmarshalled Alg should be correct")
		g.Expect(parsed.Raw).To(BeAssignableToTypeOf(""), "unmarshalled raw should be a string")

		// decrypt with correct key
		decrypted := interface{}(nil)
		e = enc.Decrypt(ctx, &parsed, &decrypted)
		g.Expect(e).To(Succeed(), "decrypted of raw data shouldn't return error")
		if v != nil {
			g.Expect(decrypted).To(BeEquivalentTo(v), "decrypted value should be correct")
		} else {
			g.Expect(decrypted).To(BeNil(), "decrypted value should be correct")
		}

		// decrypt with incorrect key
		incorrectRaw := EncryptedRaw{
			Ver:   ver,
			KeyID: uuid.Must(uuid.Parse(incorrectTestKid)),
			Alg:   AlgVault,
		}
		any := interface{}(nil)
		e = enc.Decrypt(ctx, &incorrectRaw, &any)
		g.Expect(e).To(Not(Succeed()), "decrypt with incorrect kid should return error")
	}
}

func SubTestVaultFailedEncryption(enc *vaultEncryptor, ver Version, alg Algorithm, expectedErr error) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// encrypt with nil values
		e := enc.Encrypt(ctx, nil, nil)
		g.Expect(e).To(Not(Succeed()), "Encrypt should return error")

		kid := uuid.New()
		raw := EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   alg,
		}

		// encrypt
		any := map[string]string{}
		e = enc.Encrypt(ctx, any, &raw)
		g.Expect(e).To(Not(Succeed()), "Encrypt should return error")
		g.Expect(e).To(BeIdenticalTo(expectedErr), "Encrypt should return correct error")
	}
}

func SubTestVaultFailedDecryption(enc *vaultEncryptor, ver Version, alg Algorithm, expectedErr error) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// decrypt with nil value
		e := enc.Decrypt(ctx, nil, nil)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")

		kid := uuid.New()
		raw := EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   alg,
			Raw:   map[string]interface{}{},
		}

		// decrypt
		decrypted := interface{}(nil)
		e = enc.Decrypt(ctx, &raw, &decrypted)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")
		g.Expect(e).To(BeIdenticalTo(expectedErr), "Encrypt should return correct error")
	}
}

func SubTestVaultEncryptWithoutKid(enc *vaultEncryptor) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		raw := EncryptedRaw{
			Ver:   V2,
			Alg:   AlgVault,
		}

		// encrypt
		any := map[string]string{}
		e := enc.Encrypt(ctx, any, &raw)
		g.Expect(e).To(Not(Succeed()), "Encrypt without KeyID should return error")
	}
}

func SubTestVaultDecryptWithoutKid(enc *vaultEncryptor) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		raw := EncryptedRaw{
			Ver:   V2,
			Alg:   AlgVault,
		}

		// encrypt
		decrypted := interface{}(nil)
		e := enc.Decrypt(ctx, &raw, &decrypted)
		g.Expect(e).To(Not(Succeed()), "Encrypt without KeyID should return error")
	}
}

func newTestEngine(di *transitDI) vault.TransitEngine {
	return vault.NewTransitEngine(di.Client, func(opt *vault.KeyOption) {
		opt.Exportable = true
		opt.AllowPlaintextBackup = true
	})
}
