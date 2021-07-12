package datacrypto

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

const (
	v1Plain     = `1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["java.util.HashMap",{"key":"value"}]`
	v1Encrypted = `1:965a64ac-42aa-4ec1-b30b-c3894b190691:e:vault:v1:P+CVPjwLBftDBMv1v1DnuRI2Smz7HQ0OTaGrk7yVz0U/tt183H5w5Jc98Xa77IN2FowbbqALUnGSAG5IKFrlUmaKE1rqUzMj4xCpKqBvtxWGUdK5`
	v2Plain     = `{"v":2,"kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":{"key":"value"}}`
	v2Encrypted = `{"v":2,"kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"e","d":"vault:v1:P+CVPjwLBftDBMv1v1DnuRI2Smz7HQ0OTaGrk7yVz0U/tt183H5w5Jc98Xa77IN2FowbbqALUnGSAG5IKFrlUmaKE1rqUzMj4xCpKqBvtxWGUdK5"}`
)

var (
	supportedV1Variants = []string{
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:{"key":"value"}`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:[{"key":"value"}]`,
	}
	invalidV1 = []string{
		`invalid_ver:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["java.util.HashMap",{"key":"value"}]`,
		`1:invalid_uuid:p:["java.util.HashMap",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:invalid_alg:["java.util.HashMap",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["invalid_type",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:"json string"`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:pure_string`,
	}
	supportedV2Variants = []string{
		`{"v":2,"kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":["array value"]}`,
		`{"v":2,"kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":"json string"}`,
		`{"v":2,"kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":null}`,
	}
	invalidV2 = []string{
		`{"v":4,"kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":{"key":"value"}}`,
		`[{"v":2,"kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":{"key":"value"}}]`,
		`"json string"`,
		`json string`,
	}
)

/*************************
	Test Cases
 *************************/

func TestParseEncryptedData(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestParseValidV1Plain(v1Plain), "V1PlainStandard"),
		test.GomegaSubTest(SubTestParseValidV1Plain(supportedV1Variants...), "V1PlainVariants"),
		test.GomegaSubTest(SubTestParseValidV1Encrypted(v1Encrypted), "V1Encrypted"),
		test.GomegaSubTest(SubTestParseInvalid(invalidV1...), "V1PlainInvalid"),

		test.GomegaSubTest(SubTestParseValidV2Plain(v2Plain), "V2PlainStandard"),
		test.GomegaSubTest(SubTestParseValidV2Plain(supportedV2Variants...), "V2PlainVariants"),
		test.GomegaSubTest(SubTestParseValidV2Encrypted(v2Encrypted), "V2Encrypted"),
		test.GomegaSubTest(SubTestParseInvalid(invalidV2...), "V2PlainInvalid"),
	)
}

func TestJsonUnmarshal(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestJsonUnmarshalValidV2Plain(v2Plain), "V2JsonPlainStandard"),
		test.GomegaSubTest(SubTestJsonUnmarshalValidV2Plain(supportedV2Variants...), "V2JsonPlainVariants"),
		test.GomegaSubTest(SubTestJsonUnmarshalValidV2Encrypted(v2Encrypted), "V2JsonEncrypted"),
		test.GomegaSubTest(SubTestJsonUnmarshalInvalidV2(invalidV2...), "V2JsonPlainInvalid"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestParseValidV1Plain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v, e := ParseEncryptedRaw(text)
			g.Expect(e).To(Succeed(), "parsing should be able to parse non JSON V1 format: %s", text)
			assertPlainMap(g, v, V1, text)
		}
	}
}

func SubTestParseValidV2Plain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v, e := ParseEncryptedRaw(text)
			g.Expect(e).To(Succeed(), "parsing should be able to parse JSON V2 format: %s", text)
			assertPlainMap(g, v, V2, text)
		}
	}
}

func SubTestParseInvalid(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			_, e := ParseEncryptedRaw(text)
			g.Expect(e).To(Not(Succeed()), "parsing should return error on parsing non JSON invalid V1 format: %s", text)
			g.Expect(errors.Is(e, ErrInvalidFormat)).To(BeTrue(), "parsing should returns ErrInvalidFormat: %s", text)
		}
	}
}

func SubTestJsonUnmarshalValidV2Plain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := json.Unmarshal([]byte(text), &v)
			g.Expect(e).To(Succeed(), "JSON unmarshaller should be able to parse JSON V2 format: %s", text)
			assertPlainMap(g, &v.EncryptedRaw, V2, text)
		}
	}
}

func SubTestJsonUnmarshalInvalidV2(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := json.Unmarshal([]byte(text), &v)
			g.Expect(e).To(Not(Succeed()), "JSON unmarshaller should return error on JSON V2 format: %s", text)
		}
	}
}

func SubTestParseValidV1Encrypted(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v, e := ParseEncryptedRaw(text)
			g.Expect(e).To(Succeed(), "unmarshaller should be able to parse non JSON V1 format: %s", text)
			assertEncryptedMap(g, v, V1, text)
		}
	}
}

func SubTestParseValidV2Encrypted(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v, e := ParseEncryptedRaw(text)
			g.Expect(e).To(Succeed(), "unmarshaller should be able to parse JSON V2 format: %s", text)
			assertEncryptedMap(g, v, V2, text)
		}
	}
}

func SubTestJsonUnmarshalValidV2Encrypted(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := json.Unmarshal([]byte(text), &v)
			g.Expect(e).To(Succeed(), "JSON unmarshaller should be able to parse JSON V2 format: %s", text)
			assertEncryptedMap(g, &v.EncryptedRaw, V2, text)
		}
	}
}

/*************************
	Helper
 *************************/

func assertEncryptedMap(g *gomega.WithT, v *EncryptedRaw, expectedVer Version, text string) {
	g.Expect(v.Ver).To(BeIdenticalTo(expectedVer), "parsed encrypted data should have version %s: %s", expectedVer, text)
	g.Expect(v.KeyID).To(Not(Equal(uuid.Invalid)), "parsed encrypted data should have valid KeyID : %s", text)
	g.Expect(v.Alg).To(BeIdenticalTo(AlgVault), "parsed encrypted data should have alg = Vault: %s", text)
	g.Expect(v.Raw).To(BeAssignableToTypeOf(""), "raw data of encrypted data should be a string: %s", text)

	d := v.Raw.(string)
	g.Expect(d).To(HavePrefix("vault:v1:"), "raw data of encrypted data should have proper header: %s", text)
}

func assertPlainMap(g *gomega.WithT, v *EncryptedRaw, expectedVer Version, text string) {
	g.Expect(v.Ver).To(BeIdenticalTo(expectedVer), "parsed plain data should have version %s: %s", expectedVer, text)
	g.Expect(v.KeyID).To(Not(Equal(uuid.Invalid)), "parsed plain data should have valid KeyID : %s", text)
	g.Expect(v.Alg).To(BeIdenticalTo(AlgPlain), "parsed plain data should have alg = Plain: %s", text)

	switch d := v.Raw.(type) {
	case map[string]interface{}:
		g.Expect(d).To(HaveKeyWithValue("key", "value"), "raw data of plain data should contains correct fields: %s", text)
	case []interface{}:
		g.Expect(d).To(ContainElement("array value"), "raw data of plain data should contains correct fields: %s", text)
	case string:
		g.Expect(d).To(Equal("json string"), "raw data of plain data should contains correct fields: %s", text)

	}
}