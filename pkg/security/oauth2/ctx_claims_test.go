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

package oauth2

import (
    "encoding/json"
    "github.com/cisco-open/go-lanai/pkg/utils"
    . "github.com/onsi/gomega"
    "testing"
    "time"
)

const ExtraKey = "additional"

var now = time.Now().Truncate(time.Second)

var valueStruct = ValueStruct{
	String:  "test-nested-string",
	Int:     20,
	BoolPtr: utils.BoolPtr(true),
	Bool:    true,
}

var refBasic = map[string]interface{}{
	ClaimAudience:  StringSetClaim(utils.NewStringSet("res-id")),
	ClaimExpire:    now,
	ClaimJwtId:     "test-id",
	ClaimIssueAt:   now,
	ClaimIssuer:    "test-issuer",
	ClaimNotBefore: now,
	ClaimSubject:   "test-subject",
	ClaimClientId:  "test-client-id",
	ClaimScope:     utils.NewStringSet("read", "write"),
}

var refMore = map[string]interface{}{
	ClaimAudience:  StringSetClaim(utils.NewStringSet("res-id")),
	ClaimExpire:    now,
	ClaimJwtId:     "test-id",
	ClaimIssueAt:   now,
	ClaimIssuer:    "test-issuer",
	ClaimNotBefore: now,
	ClaimSubject:   "test-subject",
	ClaimClientId:  "test-client-id",
	ClaimScope:     utils.NewStringSet("read", "write"),
	"string":       "test-string",
	"int":          10,
	"bool":         true,
	"boolPtr":      utils.BoolPtr(false),
	"struct":       valueStruct,
	"structPtr":    &valueStruct,
}

var refExtra = map[string]interface{}{
	ClaimAudience:  StringSetClaim(utils.NewStringSet("res-id")),
	ClaimExpire:    now,
	ClaimJwtId:     "test-id",
	ClaimIssueAt:   now,
	ClaimIssuer:    "test-issuer",
	ClaimNotBefore: now,
	ClaimSubject:   "test-subject",
	ClaimClientId:  "test-client-id",
	ClaimScope:     utils.NewStringSet("read", "write"),
	"string":       "test-string",
	"int":          10,
	"bool":         true,
	"boolPtr":      utils.BoolPtr(false),
	"struct":       valueStruct,
	"structPtr":    &valueStruct,
	"extra1":       "test-extra-value-1",
	"extra2":       "test-extra-value-2",
}

/*************************
	Test Cases
 *************************/
func TestBasicClaims(t *testing.T) {
	t.Run("GettersTest", GettersTest(newRefBasicClaims(), refBasic))
	t.Run("SettersTest", SettersTest(newRefBasicClaims(), refBasic, assertBasicClaims))
	t.Run("JsonTest", JsonTest(newRefBasicClaims(), &BasicClaims{}, refBasic, assertBasicClaims))
}

func TestFieldEmbeddingClaims(t *testing.T) {
	t.Run("GettersTest", GettersTest(newRefFieldEmbeddingClaims(), refMore))
	t.Run("SettersTest", SettersTest(newRefFieldEmbeddingClaims(), refMore, assertFieldEmbeddingClaims))
	t.Run("JsonTest", JsonTest(newRefFieldEmbeddingClaims(), &fieldEmbeddingClaims{}, refMore, assertFieldEmbeddingClaims))
}

func TestInterfaceEmbeddingClaims(t *testing.T) {
	t.Run("GettersTest",
		GettersTest(newRefInterfaceEmbeddingClaims(), refExtra))
	t.Run("SettersTest",
		SettersTest(newRefInterfaceEmbeddingClaims(), refExtra, assertInterfaceEmbeddingClaims))
	t.Run("JsonTest",
		JsonTest(newRefInterfaceEmbeddingClaims(), &interfaceEmbeddingClaims{
			Claims: &BasicClaims{},
			Extra:  MapClaims{},
		}, refExtra, assertInterfaceEmbeddingClaims))
	t.Run("JsonTestWithDifferentEmbedded",
		JsonTest(newRefInterfaceEmbeddingClaims(), &interfaceEmbeddingClaims{
			Claims: MapClaims{},
			Extra:  MapClaims{},
		}, refExtra, assertAlternativeInterfaceEmbeddingClaims))
}

/*************************
	Sub Tests
 *************************/
func GettersTest(claims Claims, expected map[string]interface{}) func(*testing.T) {
	return func(t *testing.T) {
		assertClaimsUsingGetter(t, expected, claims)
	}
}

func SettersTest(claims Claims, values map[string]interface{}, assertFunc claimsAssertion) func(*testing.T) {
	return func(t *testing.T) {
		// call setters
		for k, v := range values {
			claims.Set(k, v)
		}
		assertFunc(t, values, claims)
	}
}

func JsonTest(claims Claims, empty Claims, expected map[string]interface{}, assertFunc claimsAssertion) func(*testing.T) {
	return func(t *testing.T) {
		// marshal
		data, err := json.Marshal(claims)
		g := NewWithT(t)
		g.Expect(err).NotTo(HaveOccurred(), "JSON marshal should not return error")
		g.Expect(data).NotTo(BeZero(), "JSON marshal should not return empty string")

		t.Logf("JSON: %s", data)

		// unmarshal
		parsed := empty
		err = json.Unmarshal(data, &parsed)
		g.Expect(err).NotTo(HaveOccurred(), "JSON unmarshal should not return error")

		assertFunc(t, expected, parsed)
	}
}

/*************************
	Helpers
 *************************/
type claimsAssertion func(*testing.T, map[string]interface{}, Claims)

func assertClaimsUsingGetter(t *testing.T, ref map[string]interface{}, actual Claims) {
	g := NewWithT(t)
	for k, v := range ref {
		if k == ExtraKey {
			continue
		}
		g.Expect(actual.Get(k)).To(BeEquivalentTo(v), "claim [%s] should be correct", k)
	}
}

func assertBasicClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*BasicClaims)
	g.Expect(actual.Audience).To(Equal(ref[ClaimAudience]))
	g.Expect(actual.ExpiresAt).To(Equal(ref[ClaimExpire]))
	g.Expect(actual.Id).To(Equal(ref[ClaimJwtId]))
	g.Expect(actual.IssuedAt).To(Equal(ref[ClaimIssueAt]))
	g.Expect(actual.Issuer).To(Equal(ref[ClaimIssuer]))
	g.Expect(actual.NotBefore).To(Equal(ref[ClaimNotBefore]))
	g.Expect(actual.Subject).To(Equal(ref[ClaimSubject]))
	g.Expect(actual.Scopes).To(Equal(ref[ClaimScope]))
	g.Expect(actual.ClientId).To(Equal(ref[ClaimClientId]))
}

func assertFieldEmbeddingClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*fieldEmbeddingClaims)
	g.Expect(actual.Audience).To(Equal(ref[ClaimAudience]))
	g.Expect(actual.ExpiresAt).To(Equal(ref[ClaimExpire]))
	g.Expect(actual.Id).To(Equal(ref[ClaimJwtId]))
	g.Expect(actual.IssuedAt).To(Equal(ref[ClaimIssueAt]))
	g.Expect(actual.Issuer).To(Equal(ref[ClaimIssuer]))
	g.Expect(actual.NotBefore).To(Equal(ref[ClaimNotBefore]))
	g.Expect(actual.Subject).To(Equal(ref[ClaimSubject]))
	g.Expect(actual.Scopes).To(Equal(ref[ClaimScope]))
	g.Expect(actual.ClientId).To(Equal(ref[ClaimClientId]))
	g.Expect(actual.String).To(Equal(ref["string"]))
	g.Expect(actual.Int).To(Equal(ref["int"]))
	g.Expect(actual.Bool).To(Equal(ref["bool"]))
	g.Expect(actual.Struct).To(Equal(ref["struct"]))
	g.Expect(actual.BoolPtr).To(Equal(ref["boolPtr"]))
	g.Expect(actual.StructPtr).To(Equal(ref["structPtr"]))
}

func assertInterfaceEmbeddingClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*interfaceEmbeddingClaims)
	g.Expect(actual.String).To(Equal(ref["string"]))
	g.Expect(actual.Int).To(Equal(ref["int"]))
	g.Expect(actual.Bool).To(Equal(ref["bool"]))
	g.Expect(actual.Struct).To(Equal(ref["struct"]))
	g.Expect(actual.BoolPtr).To(Equal(ref["boolPtr"]))
	g.Expect(actual.StructPtr).To(Equal(ref["structPtr"]))

	for k, v := range ref {
		g.Expect(actual.Get(k)).To(Equal(v))
	}
}

func assertAlternativeInterfaceEmbeddingClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*interfaceEmbeddingClaims)
	for k, v := range ref {
		switch val := v.(type) {
		case time.Time:
			g.Expect(actual.Get(k)).To(BeEquivalentTo(val.Unix()))
		case utils.StringSet:
			g.Expect(actual.Get(k)).To(ConsistOf(val.ToSet().Values()...))
		case StringSetClaim:
			switch claimVal := actual.Get(k).(type) {
			case string:
				g.Expect(val).To(HaveKey(claimVal))
			default:
				g.Expect(claimVal).To(ConsistOf(utils.StringSet(val).ToSet().Values()...))
			}
		default:
			g.Expect(actual.Get(k)).To(Equal(v))
		}
	}
}

func newRefBasicClaims() *BasicClaims {
	return &BasicClaims{
		Audience:  refBasic[ClaimAudience].(StringSetClaim),
		ExpiresAt: refBasic[ClaimExpire].(time.Time),
		Id:        refBasic[ClaimJwtId].(string),
		IssuedAt:  refBasic[ClaimIssueAt].(time.Time),
		Issuer:    refBasic[ClaimIssuer].(string),
		NotBefore: refBasic[ClaimNotBefore].(time.Time),
		Subject:   refBasic[ClaimSubject].(string),
		Scopes:    refBasic[ClaimScope].(utils.StringSet),
		ClientId:  refBasic[ClaimClientId].(string),
	}
}

func newRefFieldEmbeddingClaims() *fieldEmbeddingClaims {
	return &fieldEmbeddingClaims{
		BasicClaims: *newRefBasicClaims(),
		String:      refMore["string"].(string),
		Int:         refMore["int"].(int),
		Bool:        refMore["bool"].(bool),
		BoolPtr:     refMore["boolPtr"].(*bool),
		Struct:      refMore["struct"].(ValueStruct),
		StructPtr:   refMore["structPtr"].(*ValueStruct),
	}
}

func newRefInterfaceEmbeddingClaims() *interfaceEmbeddingClaims {
	basic := newRefBasicClaims()
	return &interfaceEmbeddingClaims{
		Claims:    basic,
		String:      refExtra["string"].(string),
		Int:         refExtra["int"].(int),
		Bool:        refExtra["bool"].(bool),
		BoolPtr:     refExtra["boolPtr"].(*bool),
		Struct:      refExtra["struct"].(ValueStruct),
		StructPtr:   refExtra["structPtr"].(*ValueStruct),
		Extra: MapClaims{
			"extra1": refExtra["extra1"],
			"extra2": refExtra["extra2"],
		},
	}
}

/*************************
	composite Type
 *************************/
type ValueStruct struct {
	String  string `json:"string"`
	Int     int    `json:"int"`
	BoolPtr *bool  `json:"boolPtr"`
	Bool    bool   `json:"bool"`
}

// fieldEmbeddingClaims
// Note: having non-claims struct as field is not recommended for deserialization
type fieldEmbeddingClaims struct {
	FieldClaimsMapper
	BasicClaims
	String    string       `claim:"string"`
	Int       int          `claim:"int"`
	Bool      bool         `claim:"bool"`
	BoolPtr   *bool        `claim:"boolPtr"`
	Struct    ValueStruct  `claim:"struct"`
	StructPtr *ValueStruct `claim:"structPtr"`
}

func (c *fieldEmbeddingClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *fieldEmbeddingClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *fieldEmbeddingClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *fieldEmbeddingClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *fieldEmbeddingClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *fieldEmbeddingClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}

// interfaceEmbeddingClaims
type interfaceEmbeddingClaims struct {
	FieldClaimsMapper
	Claims
	String    string       `claim:"string"`
	Int       int          `claim:"int"`
	Bool      bool         `claim:"bool"`
	BoolPtr   *bool        `claim:"boolPtr"`
	Struct    ValueStruct  `claim:"struct"`
	StructPtr *ValueStruct `claim:"structPtr"`
	Extra     Claims
}

func (c *interfaceEmbeddingClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *interfaceEmbeddingClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *interfaceEmbeddingClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *interfaceEmbeddingClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *interfaceEmbeddingClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *interfaceEmbeddingClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}
