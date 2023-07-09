package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"sync"
	"testing"
)

const (
	Value       = `correct`
	ValueAlt    = `also correct`
	ValueExtra1 = `correct but not very useful`
	ValueExtra2 = `correct but even less useful`
)

var (
	schemaCache = &sync.Map{}
)

/*************************
	Test Setup
 *************************/

func SkipPolicyFilteringScopeDecorator(db *gorm.DB) *gorm.DB {
	return db.Scopes(SkipPolicyFiltering())
}

func SetupTestCreateModels(di *dbtest.DI) test.SetupFunc {
	// We use special DB scope to prepare data, to by-pass policy filtering
	return dbtest.PrepareData(di,
		dbtest.SetupUsingSQLQueries(createTableSql),
		dbtest.SetupTruncateTables(Model{}.TableName()),
	)
}

func SetupTestResultHolder() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		holder := TestResultHolder{Result: make([]policyTarget, 0, 4)}
		return context.WithValue(ctx, CKTestResultHolder{}, &holder), nil
	}
}

/*************************
	Test
 *************************/

type TestDI struct {
	fx.In
	dbtest.DI
}

func TestGormDestResolver(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(10*time.Minute),
		dbtest.WithNoopMocks(),
		apptest.WithFxOptions(
			fx.Decorate(SkipPolicyFilteringScopeDecorator),
		),
		apptest.WithDI(di),
		//test.SubTestSetup(SetupTestCreateModels(&di.DI)),
		test.SubTestSetup(SetupTestResultHolder()),
		test.GomegaSubTest(SubTestModelCreate(di), "TestModelCreate"),
		test.GomegaSubTest(SubTestModelGet(di), "TestModelGet"),
		test.GomegaSubTest(SubTestModelUpdate(di), "TestModelUpdate"),
		test.GomegaSubTest(SubTestModelDelete(di), "TestModelDelete"),
		test.GomegaSubTest(SubTestModelSave(di), "TestModelSave"),
	)
}

func TestMetadataLoader(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(10*time.Minute),
		dbtest.WithNoopMocks(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestValidOPATag(di), "TestValidOPATag"),
		test.GomegaSubTest(SubTestEmbeddedStruct(di), "TestEmbeddedStruct"),
		test.GomegaSubTest(SubTestRelationshipParsing(di), "TestRelationshipParsing"),
		test.GomegaSubTest(SubTestMissingOPATag(di), "TestMissingOPATag"),
		test.GomegaSubTest(SubTestMissingFilterField(di), "TestMissingFilterField"),
		test.GomegaSubTest(SubTestMissingResourceType(di), "TestMissingResourceType"),
		test.GomegaSubTest(SubTestMissingFieldValue(di), "TestMissingFieldValue"),
		test.GomegaSubTest(SubTestInvalidTagKV(di), "TestInvalidTagKV"),
		test.GomegaSubTest(SubTestInvalidTagFormat(di), "TestInvalidTagFormat"),
		test.GomegaSubTest(SubTestOPATagOnPrimaryKey(di), "TestOPATagOnPrimaryKey"),
	)
}

/*************************
	Sub Tests
 *************************/

/*** Dest Resolver SubTests ***/

func SubTestModelCreate(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		// by model
		model := Model{
			Value: Value,
		}
		rs = di.DB.WithContext(ctx).Create(&model)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, Value)

		// by map
		modelMap := map[string]interface{}{
			"Value": ValueAlt,
		}
		rs = di.DB.WithContext(ctx).Model(&Model{}).Create(modelMap)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, ValueAlt)

		// batch (2 batches)
		models := []*Model{{Value: Value}, {Value: ValueAlt}, {Value: ValueExtra1}, {Value: ValueExtra2}}
		rs = di.DB.WithContext(ctx).Model(&Model{}).CreateInBatches(models, 2)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, Value, ValueAlt, ValueExtra1, ValueExtra2)
	}
}

func SubTestModelGet(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		var models []*Model
		// regular find
		rs = di.DB.WithContext(ctx).Find(&models, "Value = ?", Value)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g)

		// find with model
		rs = di.DB.WithContext(ctx).Model(&Model{}).Find(&models, "Value = ?", Value)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g)

		// regular find in batches
		rs = di.DB.WithContext(ctx).FindInBatches(&models, 2, func(tx *gorm.DB, batch int) error {
			return nil
		})
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g)
	}
}

func SubTestModelUpdate(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		// by model
		model := &Model{
			Value: ValueAlt,
		}
		rs = di.DB.WithContext(ctx).Model(&Model{ID: uuid.New()}).Updates(&model)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, ValueAlt)

		// By Map
		modelMap := &map[string]interface{}{
			"Value": Value,
		}
		rs = di.DB.WithContext(ctx).Model(&Model{ID: uuid.New()}).Updates(modelMap)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, Value)

		// by column
		rs = di.DB.WithContext(ctx).Model(&Model{ID: uuid.New()}).Update("model", ValueExtra1)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, ValueExtra1)

		// UpdateColumn
		rs = di.DB.WithContext(ctx).Model(&Model{ID: uuid.New()}).UpdateColumn("model", ValueExtra2)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, ValueExtra2)

		// UpdateColumns by model
		rs = di.DB.WithContext(ctx).Model(&Model{ID: uuid.New()}).UpdateColumns(model)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, ValueAlt)

		// UpdateColumns by map
		rs = di.DB.WithContext(ctx).Model(&Model{ID: uuid.New()}).UpdateColumns(modelMap)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, Value)
	}
}

func SubTestModelDelete(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		// by ID
		rs = di.DB.WithContext(ctx).Delete(&Model{}, uuid.New())
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, "")

		// By Model
		model := &Model{
			ID:    uuid.New(),
			Value: Value,
		}
		rs = di.DB.WithContext(ctx).Delete(&model)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, Value)
	}
}

func SubTestModelSave(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		// Update
		model := &Model{
			ID:    uuid.New(),
			Value: ValueAlt,
		}
		rs = di.DB.WithContext(ctx).Save(model)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, ValueAlt)

		// New
		model.ID = uuid.UUID{}
		model.Value = Value
		rs = di.DB.WithContext(ctx).Save(&model)
		g.Expect(rs.Error).To(Succeed(), "DB operation should succeed")
		assertResults(ctx, t, g, Value)
	}
}

/*** Metadata SubTests ***/

func SubTestValidOPATag(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"field:some_field"`
		PolicyFilter `gorm:"-" opa:"type:res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
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

func SubTestEmbeddedStruct(di *TestDI) test.GomegaSubTestFunc {
	type Embedded struct {
		Value        string `gorm:"type:text" opa:"field:some_field"`
		PolicyFilter `gorm:"-" opa:"type:res,read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	type model struct {
		ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Embedded
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
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
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"field:value"`
		ToMany       []*toMany `gorm:"foreignKey:RefID;references:ID" opa:"field:to_many"`
		ToOne        *toOne    `gorm:"foreignKey:RefID;references:ID" opa:"input:to_one"`
		BelongToID   uuid.UUID
		BelongTo     *toOne `gorm:"foreignKey:BelongToID;" opa:"field:belong_to"`
		PolicyFilter `gorm:"-" opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
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

func SubTestMissingOPATag(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"field:some_field"`
		PolicyFilter `gorm:"-"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

func SubTestMissingResourceType(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"field:some_field"`
		PolicyFilter `gorm:"-" opa:"read:allow_read, update:allow_update,delete:-,create:-,"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
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
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"field:"`
		PolicyFilter `gorm:"-" opa:"type:res"`
	}
	type model2 struct {
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Relation     *relation `gorm:"foreignKey:RefID;references:ID" opa:"input:"`
		PolicyFilter `gorm:"-" opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model1{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema of model 1 should not return error")
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
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"invalid:value"`
		PolicyFilter `gorm:"-" opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

func SubTestInvalidTagFormat(di *TestDI) test.GomegaSubTestFunc {
	type model1 struct {
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"field=value"`
		PolicyFilter `gorm:"-" opa:"type:res"`
	}
	type model2 struct {
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		Value        string    `opa:"field:value"`
		PolicyFilter `gorm:"-" opa:"type=res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model1{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema of model 1 should not return error")
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata of model 1 should return error")

		s, e = schema.Parse(&model2{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema of model 2 should not return error")
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata of model 2 should return error")
	}
}

func SubTestOPATagOnPrimaryKey(di *TestDI) test.GomegaSubTestFunc {
	type model struct {
		ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();" opa:"field:id"`
		Value        string    `opa:"field:value"`
		PolicyFilter `gorm:"-" opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		s, e := schema.Parse(&model{}, schemaCache, di.DB.NamingStrategy)
		g.Expect(e).To(Succeed(), "parsing schema should not return error")
		_, e = loadMetadata(s)
		g.Expect(e).To(HaveOccurred(), "load metadata should return error")
	}
}

/*************************
	Helpers
 *************************/

func assertResults(ctx context.Context, t *testing.T, g *gomega.WithT, expected ...string) {
	resolved := ConsumeTestResults(ctx)
	g.Expect(resolved).To(HaveLen(len(expected)), "resolved values should have correct length")
	for i := range resolved {
		assertResolvedValue(resolved[i], t, g, expected[i])
	}
}

func assertResolvedValue(resolved interface{}, t *testing.T, g *gomega.WithT, expectedValue string) {
	g.Expect(resolved).ToNot(BeNil(), "resolved model should not be nil")
	g.Expect(resolved).To(BeAssignableToTypeOf(policyTarget{}), "resolved model should be correct type")
	value := resolved.(policyTarget)
	g.Expect(value.meta).ToNot(BeNil(), "resolved model should have metadata")
	if value.modelValue.IsValid() {
		assertResolvedModelValue(&value, t, g, expectedValue)
	} else {
		assertResolvedValueMap(&value, t, g, expectedValue)
	}
}

func assertResolvedModelValue(model *policyTarget, _ *testing.T, g *gomega.WithT, expected string) {
	g.Expect(model.modelValue.IsValid()).To(BeTrue(), ".modelValue should be valid")
	g.Expect(model.modelPtr.IsValid()).To(BeTrue(), ".modelPtr should be valid")
	g.Expect(model.valueMap).To(HaveLen(0), ".valueMap should be empty")
	g.Expect(model.modelValue.Type()).To(Equal(model.meta.Schema.ModelType), ".modelValue.Type() should be model's type")
	g.Expect(reflect.Indirect(model.modelPtr)).To(Equal(model.modelValue), ".modelPtr should be the pointer of .modelValue")
	v := model.modelValue.FieldByName("Value")
	g.Expect(v.IsValid()).To(BeTrue(), ".modelValue should have field 'Value'")
	g.Expect(v.Interface()).To(BeEquivalentTo(expected), ".modelValue should have correct field model")
}

func assertResolvedValueMap(model *policyTarget, _ *testing.T, g *gomega.WithT, expected string) {
	g.Expect(model.modelValue.IsValid()).To(BeFalse(), ".modelValue should be invalid")
	g.Expect(model.modelPtr.IsValid()).To(BeFalse(), ".modelPtr should be invalid")
	g.Expect(model.valueMap).ToNot(HaveLen(0), ".valueMap should not be empty")
	v, ok := model.valueMap["Value"]
	if !ok {
		v = model.valueMap["model"]
	}
	g.Expect(v).To(BeEquivalentTo(expected), ".valueMap should have correct entry")
}

type ExpectedMetadata struct {
	ResType  string
	Fields   map[string][]string
	Policies map[string]string
	Mode     uint
}

func assertMetadata(meta *metadata, g *gomega.WithT, expected *ExpectedMetadata) {
	g.Expect(meta).ToNot(BeNil(), "metadata should not be nil")
	g.Expect(meta.ResType).To(Equal(expected.ResType), "metadata should have correct resource type")
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

/*************************
	Models
 *************************/

const (
	createTableSql = `
CREATE TABLE IF NOT EXISTS public.test_opa_utils_model
(
    id            UUID        NOT NULL DEFAULT gen_random_uuid(),
    "model"       STRING      NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    FAMILY        "primary"(id, model)
);
`
)

type Model struct {
	ID           uuid.UUID                `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Value        string                   `opa:"field:model"`
	Extractor    TestModelTargetExtractor `gorm:"-"`
	PolicyFilter `gorm:"-" opa:"type:test,read:allow_read, update:allow_update,delete:-,"`
}

func (Model) TableName() string {
	return "test_opa_utils_model"
}
