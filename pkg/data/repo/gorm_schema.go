package repo

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

// GormSchemaResolver extends SchemaResolver to expose more schema related functions
type GormSchemaResolver interface{
	SchemaResolver
	// Schema returns raw schema.Schema.
	Schema() *schema.Schema
}

// GormMetadata implements GormSchemaResolver
type GormMetadata struct {
	gormSchemaResolver
	model interface{}
	types map[reflect.Type]typeKey
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
		gormSchemaResolver: gormSchemaResolver {
			schema: db.Statement.Schema,
		},
		model: reflect.New(sType).Interface(),
		types: types,
	}, nil
}

// GormMetadata implements GormSchemaResolver
type gormSchemaResolver struct {
	schema *schema.Schema
}

func (g gormSchemaResolver) ModelType() reflect.Type {
	return g.schema.ModelType
}

func (g gormSchemaResolver) Table() string {
	return g.schema.Table
}

func (g gormSchemaResolver) ColumnName(fieldName string) string {
	if f := g.lookupField(fieldName); f != nil {
		return f.DBName
	}
	return ""
}

func (g gormSchemaResolver) ColumnDataType(fieldName string) string {
	if f := g.lookupField(fieldName); f != nil {
		return string(f.DataType)
	}
	return ""
}

func (g gormSchemaResolver) RelationshipSchema(fieldName string) SchemaResolver {
	split := strings.Split(fieldName, ".")
	if s := g.followRelationships(split); s != nil {
		return gormSchemaResolver{
			schema: s,
		}
	}
	return nil
}

func (g gormSchemaResolver) Schema() *schema.Schema {
	return g.schema
}

// followRelationships find schema following relationship field path, returns nil if it cannot follow
func (g gormSchemaResolver) followRelationships(fieldPaths []string) *schema.Schema {
	ret := g.schema
	for _, fieldName := range fieldPaths {
		relation, ok := ret.Relationships.Relations[fieldName]
		if !ok || relation == nil || relation.Schema == nil {
			return nil
		}
		ret = relation.FieldSchema
	}
	return ret
}

// lookupField similar to schema.Schema.LookUpField, but priority to field name,
// this function also follow relationships, e.g. "OneToOneFieldName.FieldName"
func (g gormSchemaResolver) lookupField(name string) *schema.Field {
	var s *schema.Schema
	split := strings.Split(name, ".")
	switch len(split) {
	case 0:
		return nil
	case 1:
		s = g.schema
	default:
		if s = g.followRelationships(split[0:len(split) - 1]); s == nil {
			return nil
		}
		name = split[len(split)-1]
	}

	if field, ok := s.FieldsByName[name]; ok {
		return field
	}

	if field, ok := s.FieldsByDBName[name]; ok {
		return field
	}
	return nil
}
