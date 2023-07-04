package opadata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

var cache = &sync.Map{}

var (
	typePolicyAware   = reflect.TypeOf(PolicyAware{})
	typePolicyFilter  = reflect.TypeOf(PolicyFilter{})
	typeGenericMap    = reflect.TypeOf(map[string]interface{}{})
	policyMarkerTypes = utils.NewSet(
		typePolicyAware, typePolicyFilter,
		reflect.PointerTo(typePolicyAware),
		reflect.PointerTo(typePolicyFilter),
	)
)

const (
	errTmplEmbeddedStructNotFound = `PolicyAware not found on policyTarget [%s]. Tips: embedding PolicyAware is required for any OPA DB usage`
	errTmplOPATagNotFound         = `'opa' tag is not found on embedded PolicyAware in policyTarget [%s]. Tips: the embedded PolicyAware should have 'opa' tag with at least resource type defined`
)

/*******************
	Metadata
 *******************/

type TaggedField struct {
	schema.Field
	OPATag       opaTag
	RelationPath TaggedRelationPath
}

func (f TaggedField) InputField() string {
	if len(f.RelationPath) == 0 {
		return f.OPATag.InputField
	}
	return f.RelationPath.InputField() + "." + f.OPATag.InputField
}

type TaggedRelationship struct {
	schema.Relationship
	OPATag       opaTag
}

type TaggedRelationPath []*TaggedRelationship

func (path TaggedRelationPath) InputField() string {
	names := make([]string, len(path))
	for i := range path {
		names[i] = path[i].OPATag.InputField
	}
	return strings.Join(names, ".")
}

type metadata struct {
	ResType string
	Policy  string
	Mode    policyMode
	Fields  map[string]*TaggedField
	Schema  *schema.Schema
}

func newMetadata(s *schema.Schema) (*metadata, error) {
	fields, e := collectAllFields(s)
	if e != nil {
		return nil, e
	}
	tag, e := parseTag(s)
	if e != nil {
		return nil, e
	}
	return &metadata{
		ResType: tag.ResType,
		Policy:  tag.Policy,
		Mode:    tag.Mode,
		Fields:  fields,
		Schema:  s,
	}, nil
}

func loadMetadata(s *schema.Schema) (*metadata, error) {
	key := s.ModelType
	v, ok := cache.Load(key)
	if ok {
		return v.(*metadata), nil
	}
	newV, e := newMetadata(s)
	if e != nil {
		return nil, e
	}
	v, _ = cache.LoadOrStore(key, newV)
	return v.(*metadata), nil
}

func collectAllFields(s *schema.Schema) (ret map[string]*TaggedField, err error) {
	ret = map[string]*TaggedField{}
	if err = collectFields(s, ret); err != nil {
		return
	}
	for _, r := range s.Relationships.Relations {
		if err = collectRelationship(r, nil, utils.NewSet(), ret); err != nil {
			return
		}
	}
	return
}

func collectFields(s *schema.Schema, dest map[string]*TaggedField) error {
	for _, f := range s.Fields {
		if tag, ok := f.Tag.Lookup(TagOPA); ok {
			if len(f.DBName) == 0 {
				continue
			}
			if f.PrimaryKey {
				return ErrUnsupportedUsage.WithMessage(`"%s" tag cannot be used on primary key`, TagOPA)
			}
			tagged := TaggedField{
				Field: *f,
			}

			switch e := tagged.OPATag.UnmarshalText([]byte(tag)); {
			case e != nil:
				return ErrUnsupportedUsage.WithMessage(`invalid "%s" tag on %s.%s: %v`, TagOPA, s.Name, f.Name, e)
			case len(tagged.OPATag.InputField) == 0:
				return ErrUnsupportedUsage.WithMessage(`invalid "%s" tag on %s.%s: "%s" or "%s" is required`, TagOPA, s.Name, f.Name, TagKeyInputField, TagKeyInputFieldAlt)
			}
			dest[tagged.OPATag.InputField] = &tagged
		}
	}
	return nil
}

func collectRelationship(r *schema.Relationship, path TaggedRelationPath, visited utils.Set, dest map[string]*TaggedField) error {
	tag, ok := r.Field.Tag.Lookup(TagOPA)
	if !ok || visited.Has(r.FieldSchema) {
		return nil
	}
	visited.Add(r.FieldSchema)

	// parse OPA tag of given relation
	taggedR := TaggedRelationship{
		Relationship: *r,
	}
	switch e := taggedR.OPATag.UnmarshalText([]byte(tag)); {
	case e != nil:
		return ErrUnsupportedUsage.WithMessage(`invalid "%s" tag on %s.%s: %v`, TagOPA, r.Schema.Name, r.Field.Name, e)
	case len(taggedR.OPATag.InputField) == 0:
		return ErrUnsupportedUsage.WithMessage(`invalid "%s" tag on %s.%s: "%s" or "%s" is required`, TagOPA, r.Schema.Name, r.Field.Name, TagKeyInputField, TagKeyInputFieldAlt)
	}
	path = append(path, &taggedR)

	// collect fields of relationship's fields
	fields := map[string]*TaggedField{}
	if e := collectFields(r.FieldSchema, fields); e != nil {
		return e
	}
	for _, tagged := range fields {
		tagged.RelationPath = make([]*TaggedRelationship, len(path))
		copy(tagged.RelationPath, path)
		dest[tagged.InputField()] = tagged
	}
	// recursively collect fields of relationship
	for _, r := range r.FieldSchema.Relationships.Relations {
		if e := collectRelationship(r, path, visited, dest); e != nil {
			return e
		}
	}
	return nil
}

func parseTag(s *schema.Schema) (*opaTag, error) {
	tags, ok := findTag(s.ModelType)
	if !ok {
		return nil, fmt.Errorf(errTmplEmbeddedStructNotFound, s.Name)
	}
	tag, ok := tags.Lookup(TagOPA)
	if !ok {
		return nil, fmt.Errorf(errTmplOPATagNotFound, s.Name)
	}
	var parsed opaTag
	if e := parsed.UnmarshalText([]byte(tag)); e != nil {
		return nil, e
	}
	switch {
	case len(parsed.ResType) == 0:
		return nil, fmt.Errorf(errTmplOPATagNotFound, s.Name)
	}
	return &parsed, nil

}

// findTag recursively find tag of marker types
// result is undefined if given type and embedded type are not Struct
func findTag(typ reflect.Type) (reflect.StructTag, bool) {
	count := typ.NumField()
	for i := 0; i < count; i++ {
		f := typ.Field(i)
		if policyMarkerTypes.Has(f.Type) {
			return f.Tag, true
		}
		if f.Anonymous {
			if tag, ok := findTag(f.Type); ok {
				return tag, ok
			}
		}
	}
	return "", false
}

/*********************
	Model Resolver
 *********************/

// policyTarget collected information about current policyTarget
type policyTarget struct {
	meta       *metadata
	modelPtr   reflect.Value
	modelValue reflect.Value
	model      interface{}
	valueMap   map[string]interface{}
}

// toResourceValues convert to opa.ResourceValues
// might return nil without error if there is no recognized changes
func (m policyTarget) toResourceValues() (*opa.ResourceValues, error) {
	input := map[string]interface{}{}
	switch {
	case m.modelValue.IsValid():
		// create by model struct
		for k, tagged := range m.meta.Fields {
			rv := m.modelValue.FieldByIndex(tagged.StructField.Index)
			if rv.IsValid() && !rv.IsZero() {
				input[k] = rv.Interface()
			}
		}
	case m.valueMap != nil:
		// create by model map
		for k, tagged := range m.meta.Fields {
			v, _ := m.valueMap[tagged.Name]
			if v == nil {
				v, _ = m.valueMap[tagged.DBName]
			}
			if v != nil && !reflect.ValueOf(v).IsZero() {
				input[k] = v
			}
		}
	default:
		return nil, ErrUnsupportedUsage.WithMessage(`Cannot resolve values for model create/update`)
	}

	if len(input) == 0 {
		return nil, nil
	}
	return &opa.ResourceValues{
		ExtraData: input,
	}, nil
}

// resolvePolicyTargets resolve to be created/updated/read/deleted model values.
// depending on the operation and GORM usage, values may be extracted from Dest or ReflectValue and the extracted values
// could be struct or map
func resolvePolicyTargets(stmt *gorm.Statement, meta *metadata, op DBOperationFlag) ([]policyTarget, error) {
	resolved := make([]policyTarget, 0, 5)
	fn := func(v reflect.Value) error {
		model := policyTarget{
			meta:  meta,
			model: v.Interface(),
		}
		switch {
		case v.Type() == reflect.PointerTo(stmt.Schema.ModelType):
			model.modelPtr = v
			model.modelValue = v.Elem()
		case v.Type() == typeGenericMap:
			model.valueMap = v.Convert(typeGenericMap).Interface().(map[string]interface{})
		default:
			return fmt.Errorf("unsupported dest model [%T]", v.Interface())
		}
		resolved = append(resolved, model)
		return nil
	}

	var e error
	switch op {
	case DBOperationFlagUpdate:
		// for update, Statement.Dest should be used instead of Statement.ReflectValue.
		// See callbacks.SetupUpdateReflectValue() (update.go)
		e = walkDest(stmt, fn)
	default:
		e = walkReflectValue(stmt, fn)
	}

	if e != nil {
		return nil, fmt.Errorf("unable to extract current model model: %v", e)
	}
	return resolved, nil
}

// walkDest is similar to callbacks.callMethod. It walkthrough statement's ReflectValue
// and call given function with the pointer of the model.
func walkDest(stmt *gorm.Statement, fn func(value reflect.Value) error) (err error) {
	rv := reflect.ValueOf(stmt.Dest)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return walkValues(rv, fn)
}

// walkReflectValue is similar to callbacks.callMethod. It walkthrough statement's ReflectValue
// and call given function with the pointer of the model.
func walkReflectValue(stmt *gorm.Statement, fn func(value reflect.Value) error) (err error) {
	return walkValues(stmt.ReflectValue, fn)
}

// walkValues recursively walk give model, support slice, array, struct and map
func walkValues(rv reflect.Value, fn func(value reflect.Value) error) error {
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			for elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}
			if e := walkValues(elem, fn); e != nil {
				return e
			}
		}
	case reflect.Struct:
		if !rv.CanAddr() {
			return gorm.ErrInvalidValue
		}
		return fn(rv.Addr())
	case reflect.Map:
		return fn(rv)
	}
	return nil
}
