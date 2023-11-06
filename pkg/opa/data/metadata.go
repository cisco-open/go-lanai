package opadata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

var (
	metadataCache = &sync.Map{}
	schemaCache   = &sync.Map{}
)

var (
	typePolicyFilter  = reflect.TypeOf(PolicyFilter{})
	typeGenericMap    = reflect.TypeOf(map[string]interface{}{})
	policyMarkerTypes = utils.NewSet(
		typePolicyFilter, reflect.PointerTo(typePolicyFilter),
	)
)

const (
	errTmplEmbeddedStructNotFound = `PolicyAware not found on policyTarget [%s]. Tips: embedding PolicyAware is required for any OPA DB usage`
	errTmplOPATagNotFound         = `'opa' tag is not found on Embedded PolicyAware in policyTarget [%s]. Tips: the Embedded PolicyAware should have 'opa' tag with at least resource type defined`
)

/*******************
	Metadata
 *******************/

type TaggedField struct {
	schema.Field
	OPATag       OPATag
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
	OPATag OPATag
}

type TaggedRelationPath []*TaggedRelationship

func (path TaggedRelationPath) InputField() string {
	names := make([]string, len(path))
	for i := range path {
		names[i] = path[i].OPATag.InputField
	}
	return strings.Join(names, ".")
}

// Metadata contains all static/declarative information of a model struct.
type Metadata struct {
	OPATag
	Fields map[string]*TaggedField
	Schema *schema.Schema
}

func newMetadata(s *schema.Schema) (*Metadata, error) {
	fields, e := collectAllFields(s)
	if e != nil {
		return nil, e
	}
	tag, e := parseTag(s)
	if e != nil {
		return nil, e
	}
	return &Metadata{
		OPATag: *tag,
		Fields: fields,
		Schema: s,
	}, nil
}

func resolveMetadata(model interface{}) (*Metadata, error) {
	s, e := schema.Parse(model, schemaCache, schema.NamingStrategy{})
	if e != nil {
		return nil, e
	}
	return loadMetadata(s)
}

func loadMetadata(s *schema.Schema) (*Metadata, error) {
	key := s.ModelType
	v, ok := metadataCache.Load(key)
	if ok {
		return v.(*Metadata), nil
	}
	newV, e := newMetadata(s)
	if e != nil {
		return nil, e
	}
	v, _ = metadataCache.LoadOrStore(key, newV)
	return v.(*Metadata), nil
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

func parseTag(s *schema.Schema) (*OPATag, error) {
	tags, ok := findTag(s.ModelType)
	if !ok {
		return nil, fmt.Errorf(errTmplEmbeddedStructNotFound, s.Name)
	}
	tag, ok := tags.Lookup(TagOPA)
	if !ok {
		return nil, fmt.Errorf(errTmplOPATagNotFound, s.Name)
	}
	var parsed OPATag
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
// result is undefined if given type and Embedded type are not Struct
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
