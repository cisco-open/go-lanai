package opadata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"gorm.io/gorm/schema"
	"reflect"
	"sync"
)

var cache = &sync.Map{}

var (
	typePolicyAware   = reflect.TypeOf(PolicyAware{})
	typePolicyFilter  = reflect.TypeOf(PolicyFilter{})
	policyMarkerTypes = utils.NewSet(
		typePolicyAware, typePolicyFilter,
		reflect.PointerTo(typePolicyAware),
		reflect.PointerTo(typePolicyFilter),
	)

	typeTenancy       = reflect.TypeOf(Tenancy{})
	typeTenancyPtr    = reflect.TypeOf(&Tenancy{})
	mapKeysTenantID   = utils.NewStringSet(fieldTenantID, colTenantID)
	mapKeysTenantPath = utils.NewStringSet(fieldTenantPath, colTenantPath)
)

const (
	errTmplEmbeddedStructNotFound = `PolicyAware not found on modelInfo [%s]. Tips: embedding PolicyAware is required for any OPA DB usage`
	errTmplOPATagNotFound         = `'opa' tag is not found on embedded PolicyAware in modelInfo [%s]. Tips: the embedded PolicyAware should have 'opa' tag with at least resource type defined`
)

type TaggedField struct {
	*schema.Field
	OPATag opaTag
}

type metadata struct {
	ResType string
	Policy  string
	Fields  map[string]*TaggedField
	Schema  *schema.Schema
}

func newMetadata(s *schema.Schema) (*metadata, error) {
	fields, e := collectFields(s)
	if e != nil {
		return nil, e
	}
	tag, e := parsePolicy(s)
	if e != nil {
		return nil, e
	}
	return &metadata{
		ResType: tag.ResType,
		Policy:  tag.Policy,
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

func parsePolicy(s *schema.Schema) (*opaTag, error) {
	tags, ok := findPolicyTag(s.ModelType)
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

// findPolicyTag recursively find tag of marker types
// result is undefined if given type and embedded type are not Struct
func findPolicyTag(typ reflect.Type) (reflect.StructTag, bool) {
	count := typ.NumField()
	for i := 0; i < count; i++ {
		f := typ.Field(i)
		if policyMarkerTypes.Has(f.Type) {
			return f.Tag, true
		}
		if f.Anonymous {
			if tag, ok := findPolicyTag(f.Type); ok {
				return tag, ok
			}
		}
	}
	return "", false
}