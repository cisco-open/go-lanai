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

package opadata

import (
	"context"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/schema"
	"testing"
)

/*************************
	Test
 *************************/

func TestMetadataLoader(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(10*time.Minute),
		dbtest.WithNoopMocks(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestResolveFromModel(), "TestResolveFromModel"),
		test.GomegaSubTest(SubTestFilterAsEmbedded(di), "TestFilterAsEmbedded"),
		test.GomegaSubTest(SubTestFilterAsField(di), "TestFilterAsField"),
		test.GomegaSubTest(SubTestEmbeddedStruct(di), "TestEmbeddedStruct"),
		test.GomegaSubTest(SubTestRelationshipParsing(di), "TestRelationshipParsing"),
		test.GomegaSubTest(SubTestMissingOPATag(di), "TestMissingOPATag"),
		test.GomegaSubTest(SubTestMissingFilterField(di), "TestMissingFilterField"),
		test.GomegaSubTest(SubTestMissingResourceType(di), "TestMissingResourceType"),
		test.GomegaSubTest(SubTestMissingFieldValue(di), "TestMissingFieldValue"),
		test.GomegaSubTest(SubTestInvalidTagKV(di), "TestInvalidTagKV"),
		test.GomegaSubTest(SubTestInvalidTagFormat(di), "TestInvalidTagFormat"),
		test.GomegaSubTest(SubTestOPATagOnPrimaryKey(di), "TestOPATagOnPrimaryKey"),
		test.GomegaSubTest(SubTestOPATagOnCompositePrimaryKey(di), "TestOPATagOnCompositePrimaryKey"),
		test.GomegaSubTest(SubTestFilterGormTagOnEmbedded(di), "TestFilterGormTagOnEmbedded"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestResolveFromModel() test.GomegaSubTestFunc {
	type model struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field:some_field"`
		FilteredModel `opa:"type:res,package:test.res, read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		meta, e := resolveMetadata(&model{})
		g.Expect(e).To(Succeed(), "resolve metadata should not return error")
		assertSchema(meta.Schema, g)
		assertMetadata(meta, g, &ExpectedMetadata{
			ResType:    "res",
			OPAPackage: "test.res",
			Fields:     map[string][]string{"some_field": {"Value"}},
			Policies:   map[string]string{"read": "allow_read", "update": "allow_update", "delete": "-", "create": "-"},
			Mode:       uint(DBOperationFlagUpdate | DBOperationFlagRead),
		})
	}
}

func SubTestFilterAsEmbedded(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field:some_field"`
		FilteredModel `opa:"type:res,package:test.res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		meta, e := loadMetadata(s)
		g.Expect(e).To(Succeed(), "load metadata should not return error")
		assertMetadata(meta, g, &ExpectedMetadata{
			ResType:    "res",
			OPAPackage: "test.res",
			Fields:     map[string][]string{"some_field": {"Value"}},
			Policies:   map[string]string{"read": "allow_read", "update": "allow_update", "delete": "-", "create": "-"},
			Mode:       uint(DBOperationFlagUpdate | DBOperationFlagRead),
		})
	}
}

func SubTestMissingFilterField(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID    uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value string    `opa:"field:some_field"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

func SubTestEmbeddedStruct(di *TestDI) test.GomegaSubTestFunc {
	type Embedded struct {
		Value         string `gorm:"type:text" opa:"field:some_field"`
		FilteredModel `opa:"type:res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	type model struct {
		ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Embedded
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		meta, e := loadMetadata(s)
		g.Expect(e).To(Succeed(), "load metadata should not return error")
		assertMetadata(meta, g, &ExpectedMetadata{
			ResType:  "res",
			Fields:   map[string][]string{"some_field": {"Value"}},
			Policies: map[string]string{"read": "allow_read", "update": "allow_update", "delete": "-", "create": "-"},
			Mode:     uint(DBOperationFlagUpdate | DBOperationFlagRead),
		})
	}
}

func SubTestRelationshipParsing(di *TestDI) test.GomegaSubTestFunc {
	type toMany struct {
		ID    uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		RefID uuid.UUID
		Value string `opa:"input:value"` // "input" and "field" are interchangeable
	}
	type toOne struct {
		ID    uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		RefID uuid.UUID
		Value string `opa:"field:value"`
	}
	type model struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field:value"`
		ToMany        []*toMany `gorm:"foreignKey:RefID;references:ID" opa:"field:to_many"`
		ToOne         *toOne    `gorm:"foreignKey:RefID;references:ID" opa:"input:to_one"`
		BelongToID    uuid.UUID
		BelongTo      *toOne `gorm:"foreignKey:BelongToID;" opa:"field:belong_to"`
		FilteredModel `opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		meta, e := loadMetadata(s)
		g.Expect(e).To(Succeed(), "load metadata should not return error")
		assertMetadata(meta, g, &ExpectedMetadata{
			ResType: "res",
			Fields: map[string][]string{
				"to_many.value":   {"ToMany", "Value"},
				"to_one.value":    {"ToOne", "Value"},
				"belong_to.value": {"BelongTo", "Value"},
				"value":           {"Value"},
			},
			Mode: uint(defaultPolicyMode),
		})
	}
}

func SubTestFilterAsField(di *TestDI) test.GomegaSubTestFunc {
	type model1 struct {
		ID     uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value  string    `opa:"field:some_field"`
		Filter Filter    `opa:"type:res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	type model2 struct {
		ID     uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value  string    `opa:"field:some_field"`
		Filter Filter    `gorm:"-" opa:"type:res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// without gorm tag
		s, e := schema.Parse(&model1{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(HaveOccurred(), "parsing schema should return error when gorm tag is missing")

		// with gorm tag
		s, e = schema.Parse(&model2{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		meta, e := loadMetadata(s)
		g.Expect(e).To(Succeed(), "load metadata should not return error")
		assertMetadata(meta, g, &ExpectedMetadata{
			ResType:  "res",
			Fields:   map[string][]string{"some_field": {"Value"}},
			Policies: map[string]string{"read": "allow_read", "update": "allow_update", "delete": "-", "create": "-"},
			Mode:     uint(DBOperationFlagUpdate | DBOperationFlagRead),
		})
	}
}

func SubTestMissingOPATag(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID    uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value string    `opa:"field:some_field"`
		FilteredModel
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

func SubTestMissingResourceType(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field:some_field"`
		FilteredModel `opa:"read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

func SubTestMissingFieldValue(di *TestDI) test.GomegaSubTestFunc {
	type relation struct {
		ID    uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		RefID uuid.UUID
		Value string `opa:"field:value"`
	}
	type model1 struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field:"`
		FilteredModel `opa:"type:res"`
	}
	type model2 struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Relation      *relation `gorm:"foreignKey:RefID;references:ID" opa:"input:"`
		FilteredModel `opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model1{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema of model 1 should not return error")
		assertSchema(s, g)

		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata of model 1 should return error")

		s, e = schema.Parse(&model2{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema of model 2 should not return error")
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata of model 2 should return error")
	}
}

func SubTestInvalidTagKV(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"invalid:value"`
		FilteredModel `opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

func SubTestInvalidTagFormat(di *TestDI) test.GomegaSubTestFunc {
	type model1 struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field=value"`
		FilteredModel `opa:"type:res"`
	}
	type model2 struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field:value"`
		FilteredModel `opa:"type=res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model1{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema of model 1 should not return error")
		assertSchema(s, g)
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata of model 1 should return error")

		s, e = schema.Parse(&model2{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema of model 2 should not return error")
		assertSchema(s, g)
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata of model 2 should return error")
	}
}

func SubTestOPATagOnPrimaryKey(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();" opa:"field:id"`
		Value         string    `opa:"field:value"`
		FilteredModel `opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

func SubTestOPATagOnCompositePrimaryKey(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		PID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();" opa:"field:pid"`
		UID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();" opa:"field:uid"`
		Value         string
		FilteredModel `opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchema(s, g)

		meta, e := loadMetadata(s)
		g.Expect(e).To(Succeed(), "load metadata should not return error")
		assertMetadata(meta, g, &ExpectedMetadata{
			ResType: "res",
			Fields:  map[string][]string{"pid": {"PID"}, "uid": {"UID"}},
			Mode:    uint(defaultPolicyMode),
		})
	}
}

func SubTestFilterGormTagOnEmbedded(di *TestDI) test.GomegaSubTestFunc {
	type model1 struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value         string    `opa:"field:some_field"`
		FilteredModel `gorm:"-" opa:"type:res,package:test.res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	type nested struct {
		AnotherValue string `opa:"field:another_field"`
		model1
	}
	type Embedded struct {
		Value         string    `opa:"field:some_field"`
		FilteredModel `opa:"type:res,package:test.res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	type model2 struct {
		ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		AnotherValue string `opa:"field:another_field"`
		Embedded `gorm:"-"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// gorm tag directly applied on embedded struct
		s, e := schema.Parse(&model1{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchemaNoClauses(s, g)
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should fail")

		// gorm tag directly applied on embedded struct, but nested
		s, e = schema.Parse(&nested{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchemaNoClauses(s, g)
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should fail")

		// gorm tag indirectly applied on embedded struct
		s, e = schema.Parse(&model2{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		assertSchemaNoClauses(s, g)
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should fail")
	}
}

/*************************
	Helpers
 *************************/

type ExpectedMetadata struct {
	ResType    string
	OPAPackage string
	Fields     map[string][]string
	Policies   map[string]string
	Mode       uint
}

func assertMetadata(meta *Metadata, g *gomega.WithT, expected *ExpectedMetadata) {
	g.Expect(meta).ToNot(BeNil(), "metadata should not be nil")
	g.Expect(meta.ResType).To(Equal(expected.ResType), "metadata should have correct resource type")
	g.Expect(meta.OPAPackage).To(Equal(expected.OPAPackage), "metadata should have correct package")
	// check fields
	g.Expect(meta.Fields).To(HaveLen(len(expected.Fields)), "metadata should have correct fields")
	for k, v := range expected.Fields {
		g.Expect(meta.Fields).To(HaveKey(k), "metadata should have field [%s]", k)
		f := meta.Fields[k]
		g.Expect(f.RelationPath).To(HaveLen(len(v)-1), "field [%s] should have correct relation path", f.Name)
		for i := range v[:len(v)-1] {
			g.Expect(f.RelationPath[i].OPATag.InputField).
				ToNot(BeEmpty(), "field [%s] should have correct relation path with valid OPA tag at index %d", f.Name, i)
			g.Expect(f.RelationPath[i].Field.Name).
				To(Equal(v[i]), "field [%s] should have correct relation path at index %d", f.Name, i)
		}
		g.Expect(f.DBName).ToNot(BeEmpty(), "field [%s] should have column name", f.Name)
		g.Expect(f.OPATag.InputField).ToNot(BeEmpty(), "field [%s] should have valid opa tag", f.Name)
		g.Expect(f.Name).To(Equal(v[len(v)-1]), "field [%s] should have correct input field name", f.Name)
	}
	// check policies
	g.Expect(meta.Policies).To(HaveLen(len(expected.Policies)), "metadata should have correct policies")
	for k, v := range meta.Policies {
		kText, e := k.MarshalText()
		g.Expect(e).To(Succeed(), "marshalling meatadata's policy key should not return error")
		g.Expect(expected.Policies).To(HaveKeyWithValue(string(kText), v), "metadata should have correct policies")
	}
	// check mode
	g.Expect(meta.mode).To(BeEquivalentTo(expected.Mode), "metadata should have correct mode")
}

func assertSchema(s *schema.Schema, g *gomega.WithT) {
	g.Expect(s.CreateClauses).To(HaveLen(1), "schema's create clauses should have exactly one clause")
	g.Expect(s.QueryClauses).To(HaveLen(1), "schema's query clauses should have exactly one clause")
	g.Expect(s.UpdateClauses).To(HaveLen(1), "schema's update clauses should have exactly one clause")
	g.Expect(s.DeleteClauses).To(HaveLen(1), "schema's delete clauses should have exactly one clause")
}

func assertSchemaNoClauses(s *schema.Schema, g *gomega.WithT) {
	g.Expect(s.CreateClauses).To(HaveLen(0), "schema's create clauses should be empty")
	g.Expect(s.QueryClauses).To(HaveLen(0), "schema's query clauses should be empty")
	g.Expect(s.UpdateClauses).To(HaveLen(0), "schema's update clauses should be empty")
	g.Expect(s.DeleteClauses).To(HaveLen(0), "schema's delete clauses should be empty")
}