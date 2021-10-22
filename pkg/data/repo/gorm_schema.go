package repo

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
)

// GormSchemaResolver extends SchemaResolver to expose more schema related functions
type GormSchemaResolver interface{
	SchemaResolver
	// Schema returns raw schema.Schema.
	Schema() *schema.Schema
}

// GormMetadata implements GormSchemaResolver
type GormMetadata struct {
	model interface{}
	types map[reflect.Type]typeKey
	schema *schema.Schema
}

func newModelMetadata(db *gorm.DB, model interface{}) (GormMetadata, error) {
	if model == nil {
		return GormMetadata{}, ErrorInvalidCrudModel.WithMessage("%T is not a valid model for gorm CRUD repository", model)
	}

	// cache some types
	var sType reflect.Type

	switch t := reflect.TypeOf(model); {
	case t.Kind() == reflect.Struct:
		sType = t
	case t.Kind() == reflect.Ptr:
		for ; t.Kind() == reflect.Ptr; t = t.Elem() {}
		sType = t
	}

	if sType == nil {
		return GormMetadata{}, ErrorInvalidCrudModel.WithMessage("%T is not a valid model for gorm CRUD repository", model)
	}

	pType := reflect.PtrTo(sType)
	types := map[reflect.Type]typeKey{
		pType:                                    typeModelPtr,
		sType:                                    typeModel,
		reflect.PtrTo(reflect.SliceOf(sType)):    typeModelSlicePtr,
		reflect.PtrTo(reflect.SliceOf(pType)):    typeModelPtrSlicePtr,
		reflect.SliceOf(sType):                   typeModelSlice,
		reflect.SliceOf(pType):                   typeModelPtrSlice,
		reflect.TypeOf(map[string]interface{}{}): typeGenericMap,
	}

	// pre-parse schema
	if e := db.Statement.Parse(model); e != nil {
		return GormMetadata{}, ErrorInvalidCrudModel.WithMessage("failed to parse schema of [%T] - %v", model, e)
	}

	return GormMetadata{
		model: reflect.New(sType).Interface(),
		types: types,
		schema: db.Statement.Schema,
	}, nil
}

func (g GormMetadata) ModelType() reflect.Type {
	return reflect.Indirect(reflect.ValueOf(g.model)).Type()
}

func (g GormMetadata) Table() string {
	return g.schema.Table
}

func (g GormMetadata) ColumnName(fieldName string) string {
	if f := g.lookupField(fieldName); f != nil {
		return f.DBName
	}
	return ""
}

func (g GormMetadata) ColumnDataType(fieldName string) string {
	if f := g.lookupField(fieldName); f != nil {
		return string(f.DataType)
	}
	return ""
}

func (g GormMetadata) Schema() *schema.Schema {
	return g.schema
}

// lookupField similar to schema.Schema.LookUpField, but priority to field name
func (g GormMetadata) lookupField(name string) *schema.Field {
	if field, ok := g.schema.FieldsByName[name]; ok {
		return field
	}

	if field, ok := g.schema.FieldsByDBName[name]; ok {
		return field
	}
	return nil
}

