package pqcrypt

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	vaultinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"strings"
	"testing"
)

var (
	testKid          = "d3803a9e-f2f2-4960-bdb1-aeec92d88ca4"
	incorrectTestKid = "3100e6b7-eb62-4676-9bf4-391aba1f2fae"
	newTestKid       = "480d3866-40f5-4a3f-ab9a-c52249fca519"
)

/*************************
	Test Cases
 *************************/

type transitDI struct {
	fx.In
	Client *vault.Client `optional:"true"`
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
	props := KeyProperties{
		Type:                 defaultKeyType,
		Exportable:           true,
		AllowPlaintextBackup: true,
	}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(vaultinit.Module),
		apptest.WithDI(&di),
		test.SubTestSetup(SubTestSetupCreateKey(&di)),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, mapValue), "VaultMap"),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, strValue), "VaultString"),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, arrValue), "VaultSlice"),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, nil), "VaultNil"),
		test.GomegaSubTest(SubTestVaultCreateKey(&di, &props, newTestKid), "VaultCreateKeySuccess"),
		test.GomegaSubTest(SubTestVaultCreateKey(&di, &props, ""), "VaultCreateKeyFail"),
	)
}

func TestVaultEncryptorWithMockedTransitEngine(t *testing.T) {
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
	strValue := "this is a string"
	arrValue := []interface{}{"value1", 2.0}

	di := transitDI{}
	props := KeyProperties{
		Type:                 defaultKeyType,
		Exportable:           true,
		AllowPlaintextBackup: true,
	}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, mapValue), "VaultMap"),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, strValue), "VaultString"),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, arrValue), "VaultSlice"),
		test.GomegaSubTest(SubTestVaultEncryptor(&di, &props, testKid, nil), "VaultNil"),
		test.GomegaSubTest(SubTestVaultCreateKey(&di, &props, newTestKid), "VaultCreateKeySuccess"),
		test.GomegaSubTest(SubTestVaultCreateKey(&di, &props, ""), "VaultCreateKeyFail"),
	)
}

func TestVaultFailedEncrypt(t *testing.T) {
	enc := newMockedVaultEncryptor().(*vaultEncryptor)
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestVaultEncryptWithoutKid(enc), "EncryptWithoutKeyID"),
		test.GomegaSubTest(SubTestVaultEncryptWithBadKid(enc), "EncryptWithBadKeyID"),
	)
}

func TestVaultFailedDecrypt(t *testing.T) {
	enc := newMockedVaultEncryptor().(*vaultEncryptor)
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestVaultFailedDecryption(enc, Version(-1), AlgVault, ErrUnsupportedVersion), "InvalidVersion"),
		test.GomegaSubTest(SubTestVaultFailedDecryption(enc, V1, AlgPlain, ErrUnsupportedAlgorithm), "V1UnsupportedAlg"),
		test.GomegaSubTest(SubTestVaultFailedDecryption(enc, V2, AlgPlain, ErrUnsupportedAlgorithm), "V2UnsupportedAlg"),
		test.GomegaSubTest(SubTestVaultDecryptWithoutKid(enc), "DecryptWithoutKeyID"),
		test.GomegaSubTest(SubTestVaultDecryptWithBadKid(enc), "DecryptWithBadKeyID"),
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

func SubTestVaultEncryptor(di *transitDI, props *KeyProperties, uuidStr string, v interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var enc Encryptor
		if di.Client == nil {
			enc = newMockedVaultEncryptor()
		} else {
			enc = newVaultEncryptor(di.Client, props)
		}
		kid := uuidStr

		// encrypt
		raw, e := enc.Encrypt(ctx, kid, v)
		g.Expect(e).To(Succeed(), "Encrypt shouldn't return error")
		g.Expect(raw.Ver).To(BeIdenticalTo(V2), "encrypted data should be V2")
		g.Expect(raw.Alg).To(BeIdenticalTo(AlgVault), "encrypted data should have correct alg")
		g.Expect(raw.KeyID).To(BeIdenticalTo(kid), "encrypted data should have correct KeyID")
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
			// Note nil value always get decoded, no need to test incorrect KeyID
			g.Expect(decrypted).To(BeNil(), "decrypted value should be correct")
			return
		}

		// decrypt with incorrect key
		incorrectRaw := EncryptedRaw{
			Ver:   parsed.Ver,
			KeyID: incorrectTestKid,
			Alg:   parsed.Alg,
			Raw:   parsed.Raw,
		}
		any := interface{}(nil)
		e = enc.Decrypt(ctx, &incorrectRaw, &any)
		g.Expect(e).To(Not(Succeed()), "decrypt with incorrect kid should return error")
	}
}

func SubTestVaultFailedDecryption(enc *vaultEncryptor, ver Version, alg Algorithm, expectedErr error) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// decrypt with nil value
		e := enc.Decrypt(ctx, nil, nil)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")

		kid := uuid.New().String()
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
		// encrypt
		any := map[string]string{}
		_, e := enc.Encrypt(ctx, "", any)
		g.Expect(e).To(Not(Succeed()), "Encrypt without KeyID should return error")
	}
}

func SubTestVaultEncryptWithBadKid(enc *vaultEncryptor) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// encrypt
		any := map[string]string{}
		_, e := enc.Encrypt(ctx, incorrectTestKid, any)
		g.Expect(e).To(Not(Succeed()), "Encrypt without KeyID should return error")
	}
}

func SubTestVaultDecryptWithoutKid(enc *vaultEncryptor) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		raw := EncryptedRaw{
			Ver: V2,
			Alg: AlgVault,
			Raw: fmt.Sprintf("%s:%s", testKid, "{}"),
		}

		// encrypt
		decrypted := interface{}(nil)
		e := enc.Decrypt(ctx, &raw, &decrypted)
		g.Expect(e).To(Not(Succeed()), "Encrypt without KeyID should return error")
	}
}

func SubTestVaultDecryptWithBadKid(enc *vaultEncryptor) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		raw := EncryptedRaw{
			Ver: V2,
			Alg: AlgVault,
			Raw: fmt.Sprintf("%s:%s", incorrectTestKid, "{}"),
		}

		// encrypt
		decrypted := interface{}(nil)
		e := enc.Decrypt(ctx, &raw, &decrypted)
		g.Expect(e).To(Not(Succeed()), "Encrypt without KeyID should return error")
	}
}

func SubTestVaultCreateKey(di *transitDI, props *KeyProperties, uuidStr string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var enc Encryptor
		if di.Client == nil {
			enc = newMockedVaultEncryptor()
		} else {
			enc = newVaultEncryptor(di.Client, props)
		}
		keyOps := enc.KeyOperations()
		g.Expect(keyOps).To(Not(BeNil()), "KeyOperations shouldn't return nil")

		expectErr := uuidStr == ""

		if e := keyOps.Create(ctx, uuidStr); expectErr {
			g.Expect(e).To(Not(Succeed()), "Key creation should fail on invalid kid")
		} else {
			g.Expect(e).To(Succeed(), "Key creation should succeed on valid kid")
		}
	}
}

func newTestEngine(di *transitDI) vault.TransitEngine {
	return vault.NewTransitEngine(di.Client, func(opt *vault.KeyOption) {
		opt.Exportable = true
		opt.AllowPlaintextBackup = true
	})
}

/*************************
	Mocked
 *************************/

type mockedTransitEngine struct{}

func newMockedVaultEncryptor() Encryptor {
	return &vaultEncryptor{transit: mockedTransitEngine{}}
}

func (t mockedTransitEngine) PrepareKey(_ context.Context, _ string) error {
	return nil
}

func (t mockedTransitEngine) Encrypt(_ context.Context, kid string, plaintext []byte) ([]byte, error) {
	switch kid {
	case "":
		return nil, fmt.Errorf("invalid KeyID")
	case incorrectTestKid:
		return nil, fmt.Errorf("failed to encrypt")
	}
	cipher := fmt.Sprintf("%s:%s", kid, string(plaintext))
	return []byte(cipher), nil
}

func (t mockedTransitEngine) Decrypt(_ context.Context, kid string, cipher []byte) ([]byte, error) {
	split := strings.SplitN(string(cipher), ":", 2)
	if len(split) < 2 {
		return nil, fmt.Errorf("bad data")
	}
	if kid != split[0] {
		return nil, fmt.Errorf("wrong key")
	}
	return []byte(split[1]), nil
}
