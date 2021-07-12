package datacrypto

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

/*************************
	Test Cases
 *************************/

func TestPlainTextEncryptor(t *testing.T) {
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
	strValue := "this is a string"
	arrValue := []interface{}{"value1", 2.0}
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainTextEncryptor(V1, mapValue), "PlainTextMapV1"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(V2, mapValue), "PlainTextMapV2"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(V1, strValue), "PlainTextStringV1"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(V2, strValue), "PlainTextStringV2"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(V1, arrValue), "PlainTextSliceV1"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(V2, arrValue), "PlainTextSliceV2"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(V1, nil), "PlainTextNilV1"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(V2, nil), "PlainTextNilV2"),
	)
}

func TestPlainTextFailedEncrypt(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainTextFailedEncryption(Version(-1), AlgPlain), "InvalidVersion"),
		test.GomegaSubTest(SubTestPlainTextFailedEncryption(V1, AlgVault), "V1UnsupportedAlg"),
		test.GomegaSubTest(SubTestPlainTextFailedEncryption(V2, AlgVault), "V2UnsupportedAlg"),
	)
}

func TestPlainTextFailedDecrypt(t *testing.T) {
	m := map[string]interface{}{}
	s := ""
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainTextFailedDecryption(Version(-1), AlgPlain), "InvalidVersion"),
		test.GomegaSubTest(SubTestPlainTextFailedDecryption(V1, AlgVault), "V1UnsupportedAlg"),
		test.GomegaSubTest(SubTestPlainTextFailedDecryption(V2, AlgVault), "V2UnsupportedAlg"),
		test.GomegaSubTest(SubTestPlainTextTypeMismatch(m), "AssignmentNonPointer"),
		test.GomegaSubTest(SubTestPlainTextTypeMismatch(&s), "AssignmentTypeMismatch"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestPlainTextEncryptor(ver Version, v interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		encryptor = plainTextEncryptor{}
		kid := uuid.New()
		raw := EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   AlgPlain,
		}

		// encrypt
		e := encryptor.Encrypt(ctx, v, &raw)
		g.Expect(e).To(Succeed(), "Encrypt shouldn't return error")
		g.Expect(raw.Ver).To(BeIdenticalTo(V2), "encrypted data should be V2")
		if v != nil {
			g.Expect(raw.Raw).To(Equal(v), "encrypted raw should be correct")
		} else {
			g.Expect(raw.Raw).To(BeNil(), "encrypted raw should be correct")
		}

		// serialize
		bytes, e := json.Marshal(raw)
		g.Expect(e).To(Succeed(), "JSON marshal of raw data shouldn't return error")

		// deserialize
		parsed := EncryptedRaw{}
		e = json.Unmarshal(bytes, &parsed)
		g.Expect(e).To(Succeed(), "JSON unmarshal of raw data shouldn't return error")
		g.Expect(parsed.Ver).To(BeIdenticalTo(V2), "unmarshalled data should be V2")
		g.Expect(parsed.KeyID).To(Equal(kid), "unmarshalled KeyID should be correct")
		g.Expect(parsed.Alg).To(BeIdenticalTo(AlgPlain), "unmarshalled Alg should be correct")
		if v != nil {
			g.Expect(parsed.Raw).To(BeEquivalentTo(v), "unmarshalled Raw should be correct")
		} else {
			g.Expect(parsed.Raw).To(BeNil(), "unmarshalled Raw should be correct")
		}

		// decrypt
		decrypted := interface{}(nil)
		e = encryptor.Decrypt(ctx, &parsed, &decrypted)
		g.Expect(e).To(Succeed(), "decrypted of raw data shouldn't return error")
		if v != nil {
			g.Expect(decrypted).To(BeEquivalentTo(v), "decrypted value should be correct")
		} else {
			g.Expect(decrypted).To(BeNil(), "decrypted value should be correct")
		}
	}
}

func SubTestPlainTextFailedEncryption(ver Version, alg Algorithm) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		encryptor = plainTextEncryptor{}

		// encrypt with nil values
		e := encryptor.Encrypt(ctx, nil, nil)
		g.Expect(e).To(Not(Succeed()), "Encrypt should return error")

		kid := uuid.New()
		raw := EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   alg,
		}

		// encrypt
		any := map[string]string{}
		e = encryptor.Encrypt(ctx, any, &raw)
		g.Expect(e).To(Not(Succeed()), "Encrypt should return error")
	}
}

func SubTestPlainTextFailedDecryption(ver Version, alg Algorithm) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		encryptor = plainTextEncryptor{}
		kid := uuid.New()
		raw := EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   alg,
			Raw:   map[string]interface{}{},
		}

		// decrypt
		decrypted := interface{}(nil)
		e := encryptor.Decrypt(ctx, &raw, &decrypted)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")
	}
}

func SubTestPlainTextTypeMismatch(v interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		encryptor = plainTextEncryptor{}
		kid := uuid.New()
		raw := EncryptedRaw{
			Ver:   V2,
			KeyID: kid,
			Alg:   AlgPlain,
			Raw:   map[string]interface{}{},
		}

		// decrypt
		e := encryptor.Decrypt(ctx, &raw, v)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")
	}
}
