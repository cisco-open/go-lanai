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
    "encoding/json"
    "github.com/cisco-open/go-lanai/test"
    "github.com/google/uuid"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "testing"
)

/*************************
	Test Cases
 *************************/

func TestPlainTextEncryptor(t *testing.T) {
	enc := plainTextEncryptor{}
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
	strValue := "this is a string"
	arrValue := []interface{}{"value1", 2.0}
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainTextEncryptor(enc, mapValue), "PlainTextMap"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(enc, strValue), "PlainTextString"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(enc, arrValue), "PlainTextSlice"),
		test.GomegaSubTest(SubTestPlainTextEncryptor(enc, nil), "PlainTextNil"),
	)
}

func TestPlainTextV1Decryption(t *testing.T) {
	enc := plainTextEncryptor{}
	const kid = `e0622fd0-d2ca-11eb-9c82-bd03f2eed750`
	const mapData     = `{"v":1,"kid":"e0622fd0-d2ca-11eb-9c82-bd03f2eed750","alg":"p","d":["java.util.HashMap",{"key1":"value1","key2":2}]}`
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
	const strData     = `{"v":1,"kid":"e0622fd0-d2ca-11eb-9c82-bd03f2eed750","alg":"p","d":["java.lang.String","this is a string"]}`
	strValue := "this is a string"
	const arrData     = `{"v":1,"kid":"e0622fd0-d2ca-11eb-9c82-bd03f2eed750","alg":"p","d":["java.util.ArrayList",["value1",2]]}`
	arrValue := []interface{}{"value1", 2.0}
	const jsonNullData     = `{"v":1,"kid":"e0622fd0-d2ca-11eb-9c82-bd03f2eed750","alg":"p","d":null}`
	const nilData     = `{"v":1,"kid":"e0622fd0-d2ca-11eb-9c82-bd03f2eed750","alg":"p"}`
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainTextV1Decryption(enc, mapData, kid, mapValue), "PlainTextMap"),
		test.GomegaSubTest(SubTestPlainTextV1Decryption(enc, strData, kid, strValue), "PlainTextString"),
		test.GomegaSubTest(SubTestPlainTextV1Decryption(enc, arrData, kid, arrValue), "PlainTextSlice"),
		test.GomegaSubTest(SubTestPlainTextV1Decryption(enc, jsonNullData, kid, nil), "PlainTextJsonNull"),
		test.GomegaSubTest(SubTestPlainTextV1Decryption(enc, nilData, kid, nil), "PlainTextNil"),
	)
}

func TestPlainTextFailedEncrypt(t *testing.T) {
	enc := plainTextEncryptor{}
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainTextFailedEncryption(enc), "InvalidKeyID"),
	)
}

func TestPlainTextFailedDecrypt(t *testing.T) {
	enc := plainTextEncryptor{}
	m := map[string]interface{}{}
	s := ""

	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainTextFailedDecryption(enc, Version(-1), AlgPlain, ErrUnsupportedVersion), "InvalidVersion"),
		test.GomegaSubTest(SubTestPlainTextFailedDecryption(enc, V1, AlgVault, ErrUnsupportedAlgorithm), "V1UnsupportedAlg"),
		test.GomegaSubTest(SubTestPlainTextFailedDecryption(enc, V2, AlgVault, ErrUnsupportedAlgorithm), "V2UnsupportedAlg"),
		test.GomegaSubTest(SubTestPlainTextTypeMismatch(enc, m), "AssignmentNonPointer"),
		test.GomegaSubTest(SubTestPlainTextTypeMismatch(enc, &s), "AssignmentTypeMismatch"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestPlainTextEncryptor(enc Encryptor, v interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		kid := uuid.New().String()

		// encrypt
		raw, e := enc.Encrypt(ctx, kid, v)
		g.Expect(e).To(Succeed(), "Encrypt shouldn't return error")
		g.Expect(raw.Ver).To(BeIdenticalTo(V2), "encrypted data should be V2")
		g.Expect(raw.Alg).To(BeIdenticalTo(AlgPlain), "encrypted data should have correct alg")
		g.Expect(raw.KeyID).To(BeIdenticalTo(kid), "encrypted data should have correct KeyID")
		expected, _ := json.Marshal(v)
		g.Expect(raw.Raw).To(MatchJSON(expected), "encrypted raw should be correct")

		// serialize
		bytes, e := json.Marshal(raw)
		g.Expect(e).To(Succeed(), "JSON marshal of raw data shouldn't return error")

		testPlainTextDecryption(g, enc, bytes, V2, kid, v)
	}
}

func SubTestPlainTextV1Decryption(enc Encryptor, text string, expectedKid string, expectedVal interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		testPlainTextDecryption(g, enc, []byte(text), V1, expectedKid, expectedVal)
	}
}

func SubTestPlainTextFailedEncryption(enc Encryptor) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// encrypt with nil values
		_, e := enc.Encrypt(ctx, "", nil)
		g.Expect(e).To(Not(Succeed()), "Encrypt should return error")
	}
}

func SubTestPlainTextFailedDecryption(enc Encryptor, ver Version, alg Algorithm, expectedErr error) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// decrypt with nil value
		e := enc.Decrypt(ctx, nil, nil)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")

		kid := uuid.New().String()
		raw := EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   alg,
			Raw:   json.RawMessage(`{}`),
		}

		// decrypt
		decrypted := interface{}(nil)
		e = enc.Decrypt(ctx, &raw, &decrypted)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")
		g.Expect(e).To(BeIdenticalTo(expectedErr), "Encrypt should return correct error")
	}
}

func SubTestPlainTextTypeMismatch(enc Encryptor, v interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		encryptor = plainTextEncryptor{}
		kid := uuid.New().String()
		raw := EncryptedRaw{
			Ver:   V2,
			KeyID: kid,
			Alg:   AlgPlain,
			Raw:   json.RawMessage(`{}`),
		}

		// decrypt
		e := enc.Decrypt(ctx, &raw, v)
		g.Expect(e).To(Not(Succeed()), "Decrypt of raw data should return error")
	}
}

/* Helpers */

func testPlainTextDecryption(g *gomega.WithT, enc Encryptor, bytes []byte, expectedVer Version, expectedKid string, expectedVal interface{}) {
	// deserialize
	parsed := EncryptedRaw{}
	e := json.Unmarshal(bytes, &parsed)
	g.Expect(e).To(Succeed(), "JSON unmarshal of raw data shouldn't return error")
	g.Expect(parsed.Ver).To(BeIdenticalTo(expectedVer), "unmarshalled data should be V2")
	g.Expect(parsed.KeyID).To(Equal(expectedKid), "unmarshalled KeyID should be correct")
	g.Expect(parsed.Alg).To(BeIdenticalTo(AlgPlain), "unmarshalled Alg should be correct")

	// decrypt
	decrypted := interface{}(nil)
	e = enc.Decrypt(context.Background(), &parsed, &decrypted)
	g.Expect(e).To(Succeed(), "decrypted of raw data shouldn't return error")
	if expectedVal != nil {
		g.Expect(decrypted).To(BeEquivalentTo(expectedVal), "decrypted value should be correct")
	} else {
		g.Expect(decrypted).To(BeNil(), "decrypted value should be correct")
	}
}
