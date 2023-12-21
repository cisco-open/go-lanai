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

package pqcrypt

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
	v2Plain     = `{"v":2,"kid":"e0622fd0-d2ca-11eb-9c82-bd03f2eed750","alg":"p","d":{"key":"value"}}`
	v2Encrypted = `{"v":2,"kid":"965a64ac-42aa-4ec1-b30b-c3894b190691","alg":"e","d":"vault:v1:P+CVPjwLBftDBMv1v1DnuRI2Smz7HQ0OTaGrk7yVz0U/tt183H5w5Jc98Xa77IN2FowbbqALUnGSAG5IKFrlUmaKE1rqUzMj4xCpKqBvtxWGUdK5"}`

	v1PlainJson     = `{"v":1,"kid":"e0622fd0-d2ca-11eb-9c82-bd03f2eed750","alg":"p","d":["java.util.HashMap",{"key":"value"}]}`
	v1EncryptedJson = `{"v":1,"kid":"965a64ac-42aa-4ec1-b30b-c3894b190691","alg":"e","d":"vault:v1:P+CVPjwLBftDBMv1v1DnuRI2Smz7HQ0OTaGrk7yVz0U/tt183H5w5Jc98Xa77IN2FowbbqALUnGSAG5IKFrlUmaKE1rqUzMj4xCpKqBvtxWGUdK5"}`
)

var (
	supportedV1Variants = []string{
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:{"key":"value"}`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:[{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["java.lang.String","json string"]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["java.util.ArrayList",["array value",2]]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:null`,

	}
	invalidV1 = []string{
		`invalid_ver:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["java.util.HashMap",{"key":"value"}]`,
		`1:invalid_uuid:p:["java.util.HashMap",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:invalid_alg:["java.util.HashMap",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:pure_string`,
		// note: following test case doesn't generate error at time of parsing
		// 		 the error is deferred to decrypting phase
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:["invalid_type",{"key":"value"}]`,
		`1:e0622fd0-d2ca-11eb-9c82-bd03f2eed750:p:"json string"`,
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
		test.GomegaSubTest(SubTestParseInvalid(V1, invalidV1...), "V1PlainInvalid"),

		test.GomegaSubTest(SubTestParseValidV2Plain(v2Plain), "V2PlainStandard"),
		test.GomegaSubTest(SubTestParseValidV2Plain(supportedV2Variants...), "V2PlainVariants"),
		test.GomegaSubTest(SubTestParseValidV2Encrypted(v2Encrypted), "V2Encrypted"),
		test.GomegaSubTest(SubTestParseInvalid(V2, invalidV2...), "V2PlainInvalid"),
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

func TestJsonbValuer(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestJsonbValuer(v1Plain, v1PlainJson), "ConvertPlain"),
		test.GomegaSubTest(SubTestJsonbValuer(v1Encrypted, v1EncryptedJson), "ConvertEncrypted"),
	)
}

func TestJsonbScanner(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestScanValidV2(false, v2Plain), "ScanV2Plain"),
		test.GomegaSubTest(SubTestScanValidV2(true, v2Encrypted), "ScanV2Encrypted"),
		test.GomegaSubTest(SubTestScanValidV2(false, supportedV2Variants...), "ScanV2Variants"),
		test.GomegaSubTest(SubTestScanInvalidV2(v1Plain, v1Encrypted), "ScanV1"),
		test.GomegaSubTest(SubTestScanInvalidV2(invalidV1...), "ScanInvalidV1"),
		test.GomegaSubTest(SubTestScanInvalidV2(invalidV2...), "ScanInvalidV2"),
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
			assertPlainRaw(g, v, V1, text)
		}
	}
}

func SubTestParseValidV2Plain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v, e := ParseEncryptedRaw(text)
			g.Expect(e).To(Succeed(), "parsing should be able to parse JSON V2 format: %s", text)
			assertPlainRaw(g, v, V2, text)
		}
	}
}

func SubTestParseInvalid(expectedVer Version, texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			raw, e := ParseEncryptedRaw(text)
			if e != nil {
				g.Expect(errors.Is(e, ErrInvalidFormat)).To(BeTrue(), "parsing should returns ErrInvalidFormat: %s", text)
			} else {
				assertInvalidPlainRaw(g, raw, expectedVer, text)
			}
		}
	}
}

func SubTestJsonUnmarshalValidV2Plain(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := json.Unmarshal([]byte(text), &v)
			g.Expect(e).To(Succeed(), "JSON unmarshaller should be able to parse JSON V2 format: %s", text)
			assertPlainRaw(g, &v.EncryptedRaw, V2, text)
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
			assertEncryptedRaw(g, v, V1, text)
		}
	}
}

func SubTestParseValidV2Encrypted(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v, e := ParseEncryptedRaw(text)
			g.Expect(e).To(Succeed(), "unmarshaller should be able to parse JSON V2 format: %s", text)
			assertEncryptedRaw(g, v, V2, text)
		}
	}
}

func SubTestJsonUnmarshalValidV2Encrypted(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedMap{}
			e := json.Unmarshal([]byte(text), &v)
			g.Expect(e).To(Succeed(), "JSON unmarshaller should be able to parse JSON V2 format: %s", text)
			assertEncryptedRaw(g, &v.EncryptedRaw, V2, text)
		}
	}
}

func SubTestJsonbValuer(text, expected string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		v, e := ParseEncryptedRaw(text)
		g.Expect(e).To(Succeed(), "parsing should be able to parse non JSON V1 format: %s", text)
		jsonb, e := v.Value()
		g.Expect(e).To(Succeed(), "Value() should not returns error: %s", text)
		g.Expect(jsonb).To(BeAssignableToTypeOf(""), "Value() should returns []byte: %s", text)
		jsonStr := jsonb.(string)
		g.Expect(jsonStr).To(MatchJSON(expected), "Value() should returns correct result: %s", text)
		g.Expect(v.Ver).To(Equal(V1), "Value() also correct version: %s", text)
	}
}

func SubTestScanValidV2(encrypted bool, texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedRaw{}
			e := v.Scan(text)
			g.Expect(e).To(Succeed(), "Scanner should be able to parse JSON V2 format: %s", text)
			if encrypted {
				assertEncryptedRaw(g, &v, V2, text)
			} else {
				assertPlainRaw(g, &v, V2, text)
			}
		}
	}
}

func SubTestScanInvalidV2(texts ...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		for _, text := range texts {
			v := EncryptedRaw{}
			e := v.Scan(text)
			g.Expect(e).To(Not(Succeed()), "Scanner should return error on JSON V2 format: %s", text)
		}
	}
}

/*************************
	Helper
 *************************/

func assertEncryptedRaw(g *gomega.WithT, v *EncryptedRaw, expectedVer Version, text string) {
	g.Expect(v.Ver).To(BeIdenticalTo(expectedVer), "parsed encrypted data should have version %s: %s", expectedVer, text)
	g.Expect(v.KeyID).To(Not(Equal(uuid.UUID{})), "parsed encrypted data should have valid KeyID : %s", text)
	g.Expect(v.Alg).To(BeIdenticalTo(AlgVault), "parsed encrypted data should have alg = Vault: %s", text)

	var payload interface{}
	e := json.Unmarshal(v.Raw, &payload)
	g.Expect(e).To(Succeed(), "raw data should be a json")
	g.Expect(payload).To(BeAssignableToTypeOf(""), "raw data of encrypted data should be a string: %s", text)
	d := payload.(string)
	g.Expect(d).To(HavePrefix("vault:v1:"), "raw data of encrypted data should have proper header: %s", text)
}

func assertPlainRaw(g *gomega.WithT, v *EncryptedRaw, expectedVer Version, text string) {
	g.Expect(v.Ver).To(BeIdenticalTo(expectedVer), "parsed plain data should have version %s: %s", expectedVer, text)
	g.Expect(v.KeyID).To(Not(Equal(uuid.UUID{})), "parsed plain data should have valid KeyID : %s", text)
	g.Expect(v.Alg).To(BeIdenticalTo(AlgPlain), "parsed plain data should have alg = Plain: %s", text)

	var payload interface{}
	if expectedVer == V1 {
		data, e := extractV1DecryptedPayload(v.Raw)
		g.Expect(e).To(Succeed(), "V1 plain raw data should be extractable")
		e = json.Unmarshal(data, &payload)
		g.Expect(e).To(Succeed(), "raw data should be a json")
	} else {
		e := json.Unmarshal(v.Raw, &payload)
		g.Expect(e).To(Succeed(), "raw data should be a json")
	}

	switch d := payload.(type) {
	case map[string]interface{}:
		g.Expect(d).To(HaveKeyWithValue("key", "value"), "raw data of plain data should be map contains correct fields: %s", text)
	case []interface{}:
		g.Expect(d).To(ContainElement("array value"), "raw data of plain data should be array with correct elements: %s", text)
	case string:
		g.Expect(d).To(Equal("json string"), "raw data of plain data should be string with correct value: %s", text)
	case nil:
		g.Expect(d).To(BeNil(), "raw data of plain data should be nil: %s", text)
	default:
		g.Expect(true).To(BeFalse(), "unexpeded raw data of plain data")
	}
}

func assertInvalidPlainRaw(g *gomega.WithT, v *EncryptedRaw, expectedVer Version, text string) {
	g.Expect(v.Ver).To(BeIdenticalTo(expectedVer), "parsed plain data should have version %s: %s", expectedVer, text)
	g.Expect(v.KeyID).To(Not(Equal(uuid.UUID{})), "parsed plain data should have valid KeyID : %s", text)
	g.Expect(v.Alg).To(BeIdenticalTo(AlgPlain), "parsed plain data should have alg = Plain: %s", text)

	var payload interface{}
	if expectedVer == V1 {
		_, e := extractV1DecryptedPayload(v.Raw)
		g.Expect(e).To(Not(Succeed()), "V1 plain raw data should not be extractable")
	} else {
		e := json.Unmarshal(v.Raw, &payload)
		g.Expect(e).To(Not(Succeed()), "raw data should not be a json")
	}
}
