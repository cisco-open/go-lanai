package datacrypto

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

const (
	v1Plain     = `1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["java.util.HashMap",{"key":"value"}]`
	v1Encrypted = `1:965a64ac-42aa-4ec1-b30b-c3894b190691:e:vault:v1:P+CVPjwLBftDBMv1v1DnuRI2Smz7HQ0OTaGrk7yVz0U/tt183H5w5Jc98Xa77IN2FowbbqALUnGSAG5IKFrlUmaKE1rqUzMj4xCpKqBvtxWGUdK5`
	v2Plain     = `{"v":"2","kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":{"key":"value"}}`
	v2Encrypted = ``
)

var (
	supportedV1Variants = []string{
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:{"key":"value"}`,
	}
	invalidV1 = []string{
		`invalid_ver:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["java.util.HashMap",{"key":"value"}]`,
		`1:invalid_uuid:p:["java.util.HashMap",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:invalid_alg:["java.util.HashMap",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["invalid_type",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:"json string"`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:pure_string"`,
	}
	supportedV2Variants = []string{
		`{"v":"2","kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":["value"]}`,
		`{"v":"2","kid":"d034a284-172f-46c3-aead-e7cfb2f78ddc","alg":"p","d":"json string"}`,
	}
)

/*************************
	Test Cases
 *************************/

func TestUnmarshal(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestUnmarshalValidV1Plain(v1Plain), "V1PlainStandard"),
		test.GomegaSubTest(SubTestUnmarshalValidV1Plain(supportedV1Variants...), "V1PlainVariants"),
		test.GomegaSubTest(SubTestUnmarshalInvalidPlain(invalidV1...), "V1PlainInvalid"),
		test.GomegaSubTest(SubTestUnmarshalValidV2Plain(v2Plain), "V2PlainStandard"),
		test.GomegaSubTest(SubTestUnmarshalValidV2Plain(supportedV2Variants...), "V2PlainVariants"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestUnmarshalValidV1Plain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := unmarshalText(text, &v)
			g.Expect(e).To(Succeed(), "unmarshaller should be able to parse non JSON V1 format: %s", text)
			g.Expect(v.Ver).To(BeIdenticalTo(V1), "parsed plain data should be V1: %s", text)
			g.Expect(v.UUID).To(Not(Equal(uuid.Invalid)), "parsed plain data should have valid UUID : %s", text)
			g.Expect(v.Alg).To(BeIdenticalTo(AlgPlain), "parsed plain data should have alg = Plain: %s", text)
			g.Expect(v.Raw).To(BeAssignableToTypeOf(map[string]interface{}{}), "raw data of plain data should be a map[string]interface{}: %s", text)

			d := v.Raw.(map[string]interface{})
			g.Expect(d).To(HaveKeyWithValue("key", "value"), "raw data of plain data should contains correct fields: %s", text)
		}
	}
}

func SubTestUnmarshalValidV2Plain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := unmarshalText(text, &v)
			g.Expect(e).To(Succeed(), "unmarshaller should be able to parse JSON V2 format: %s", text)
			g.Expect(v.Ver).To(BeIdenticalTo(V2), "parsed plain data should be V2: %s", text)
			g.Expect(v.UUID).To(Not(Equal(uuid.Invalid)), "parsed plain data should have valid UUID : %s", text)
			g.Expect(v.Alg).To(BeIdenticalTo(AlgPlain), "parsed plain data should have alg = Plain: %s", text)
			//g.Expect(v.Raw).To(BeAssignableToTypeOf(map[string]interface{}{}), "raw data of plain data should be a map[string]interface{}: %s", text)

			//d := v.Raw.(map[string]interface{})
			//g.Expect(d).To(HaveKeyWithValue("key", "value"), "raw data of plain data should contains correct fields: %s", text)
		}
	}
}

func SubTestUnmarshalInvalidPlain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := unmarshalText(text, &v)
			g.Expect(e).To(Not(Succeed()), "unmarshaller should return error on parsing non JSON invalid V1 format: %s", text)
		}
	}
}
