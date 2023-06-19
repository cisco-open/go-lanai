package opadata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
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
	errTmplEmbeddedStructNotFound = `PolicyAware not found on targetModel [%s]. Tips: embedding PolicyAware is required for any OPA DB usage`
	errTmplOPATagNotFound         = `'opa' tag is not found on embedded PolicyAware in targetModel [%s]. Tips: the embedded PolicyAware should have 'opa' tag with at least resource type defined`
)

/*******************
	Metadata
 *******************/

type TaggedField struct {
	*schema.Field
	OPATag opaTag
}

type metadata struct {
	ResType string
	Policy  string
	Mode    policyMode
	Fields  map[string]*TaggedField
	Schema  *schema.Schema
}

func newMetadata(s *schema.Schema) (*metadata, error) {
	fields, e := collectFields(s)
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

func collectFields(s *schema.Schema) (ret map[string]*TaggedField, err error) {
	ret = map[string]*TaggedField{}
	for _, f := range s.Fields {
		if tag, ok := f.Tag.Lookup(TagOPA); ok {
			if len(f.DBName) == 0 {
				continue
			}
			tagged := TaggedField{
				Field: f,
			}
			if e := tagged.OPATag.UnmarshalText([]byte(tag)); e != nil {
				return nil, e
			}
			ret[tagged.OPATag.InputField] = &tagged
		}
	}
	return
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

// targetModel collected information about current targetModel
type targetModel struct {
	meta     *metadata
	ptr      reflect.Value
	val      reflect.Value
	valueMap map[string]interface{}
}

func resolveTargetModel(stmt *gorm.Statement, meta *metadata) (m targetModel, err error) {
	m.meta = meta
	v, e := resolveTargetModelValue(stmt)
	if e != nil {
		return m, e
	}
	switch v.Kind() {
	case reflect.Pointer:
		m.ptr = v
		m.val = v.Elem()
		// sanity check
		if m.meta.Schema.ModelType != m.val.Type() {
			return targetModel{}, fmt.Errorf("policy metadata and current model type mismatches")
		}
	case reflect.Map:
		m.valueMap = v.Convert(typeGenericMap).Interface().(map[string]interface{})
	}
	return

}

// resolveTargetModelValue find the pointer of enclosing targetModel struct or a map
func resolveTargetModelValue(stmt *gorm.Statement) (rv reflect.Value, err error) {
	// batch vs single model
	switch stmt.ReflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		if stmt.CurDestIndex >= stmt.ReflectValue.Len() {
			break
		}
		rv = stmt.ReflectValue.Index(stmt.CurDestIndex)
	case reflect.Struct:
		fallthrough
	case reflect.Map:
		rv = stmt.ReflectValue
	}

	switch {
	case !rv.IsValid():
		break
	case rv.Type() == stmt.Schema.ModelType && rv.CanAddr():
		return rv.Addr(), nil
	case rv.Type() == reflect.PointerTo(stmt.Schema.ModelType):
		return rv, nil
	case rv.Type() == typeGenericMap:
		return rv, nil
	}
	return rv, fmt.Errorf("unable to extract current model value")
}

func resolveTargetModelAsMap(stmt *gorm.Statement) (rv reflect.Value, err error) {
	return
}
