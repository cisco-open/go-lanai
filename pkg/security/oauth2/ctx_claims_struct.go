package oauth2

import (
	"encoding/json"
	"fmt"
	"reflect"
)

/***************************
	Struct Claims Helpers
 ***************************/
var (
	claimsType    = reflect.TypeOf(Claims(MapClaims{}))
	mapClaimsType = reflect.TypeOf(MapClaims{})
	mapperType    = reflect.TypeOf(FieldClaimsMapper{})
)

type accumulator func(i interface{}, claims Claims) (accumulated interface{}, shouldContinue bool)

type claimsMapper interface {
	toMap(owner interface{}) (map[string]interface{}, error)
	fromMap(owner interface{}, src map[string]interface{}) error
}

// FieldClaimsMapper is a helper type that can be embedded into struct based claims
// FieldClaimsMapper implements claimsMapper
type FieldClaimsMapper struct {
	fields     map[string][]int // Index of fields holding claim. Includes embedded structs
	interfaces [][]int 			// Index of directly embedded Cliams interfaces
}

func (m *FieldClaimsMapper) Get(owner interface{}, claim string) interface{} {
	v := m.findFieldValue(owner, claim)
	if v.IsValid() && !v.IsZero() {
		return v.Interface()
	}
	// try with all embedded Claims interface
	return m.aggregateEmbeddedClaims(owner, nil, func(i interface{}, claims Claims) (interface{}, bool) {
		if claims.Has(claim) {
			// found it, don't continue
			return claims.Get(claim), false
		}
		return nil, true
	})
}

func (m *FieldClaimsMapper) Has(owner interface{}, claim string) bool {
	v := m.findFieldValue(owner, claim)
	if !v.IsValid() || v.IsZero() {
		// try with all embedded Claims interface
		return m.aggregateEmbeddedClaims(owner, false, func(i interface{}, claims Claims) (interface{}, bool) {
			has := claims.Has(claim)
			return has, !has
		}).(bool)
	}
	return true
}

func (m *FieldClaimsMapper) Set(owner interface{}, claim string, value interface{}) {
	v := m.findFieldValue(owner, claim)
	if v.IsValid() {
		if e := m.set(v, value); e != nil {
			panic(e)
		}
	}

	// try with all embedded Claims interface
	m.aggregateEmbeddedClaims(owner, nil, func(i interface{}, claims Claims) (interface{}, bool) {
		claims.Set(claim, value)
		return nil, true
	})
}

func (m *FieldClaimsMapper) DoMarshalJSON(owner interface{}) ([]byte, error) {
	v, e := m.toMap(owner)
	if e != nil {
		return nil, e
	}
	return json.Marshal(v)
}

func (m *FieldClaimsMapper) DoUnmarshalJSON(owner interface{}, bytes []byte) error {
	values := map[string]interface{}{}
	if e := json.Unmarshal(bytes, &values); e != nil {
		return e
	}

	if e := m.fromMap(owner, values); e != nil {
		return e
	}
	return nil
}

func (m *FieldClaimsMapper) findFieldValue(owner interface{}, claim string) (ret reflect.Value) {
	m.prepare(owner)
	index, ok := m.fields[claim]
	if !ok {
		return
	}

	v := reflect.ValueOf(owner)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.FieldByIndex(index)
}

func (m *FieldClaimsMapper) set(fv reflect.Value, setTo interface{}) error {
	if fv.Kind() == reflect.Interface {
		fv = fv.Elem()
	}

	if fv.Kind() == reflect.Ptr {
		fv = fv.Elem()
	}

	if !fv.CanSet() {
		return fmt.Errorf("field [%v] is not settable", fv.Type())
	}

	v := reflect.ValueOf(setTo)
	t := v.Type()
	ft := fv.Type()

	switch {
	case t.AssignableTo(ft):
		fv.Set(v)
	case v.Type().ConvertibleTo(ft):
		fv.Set(v.Convert(ft))
	default:
		return fmt.Errorf("value with type [%v] cannot be set to field [%v]", t, ft)
	}

	return nil
}

func (m *FieldClaimsMapper) toMap(owner interface{}) (map[string]interface{}, error) {
	m.prepare(owner)

	// try to aggregate values from internal Claims interfaces first
	var err error
	ret := m.aggregateEmbeddedClaims(owner, map[string]interface{}{}, func(i interface{}, claims Claims) (interface{}, bool) {
		var values map[string]interface{}
		if sc, ok := claims.(claimsMapper); ok {
			values, err = sc.toMap(sc)
			if err != nil {
				return nil, false
			}
		} else if mc, ok := claims.(MapClaims); ok {
			values, err = mc.toMap()
			if err != nil {
				return nil, false
			}
		}

		aggregated := i.(map[string]interface{})
		if values != nil {
			for k, v := range values {
				aggregated[k] = v
			}
		}
		return aggregated, true
	}).(map[string]interface{})

	if err != nil {
		return nil, err
	}

	// collect claims from known fields
	ov := reflect.ValueOf(owner)
	if ov.Kind() == reflect.Ptr {
		ov = ov.Elem()
	}
	for k, index := range m.fields {
		fv := ov.FieldByIndex(index)
		if fv.IsValid() && !fv.IsZero() {
			v, e := claimMarshalConvert(fv)
			if e != nil {
				return nil, e
			}
			ret[k] = v.Interface()
		}
	}
	return ret, nil
}

func (m *FieldClaimsMapper) fromMap(owner interface{}, src map[string]interface{}) error {
	m.prepare(owner)
	ov := reflect.ValueOf(owner)
	if ov.Kind() == reflect.Ptr {
		ov = ov.Elem()
	}

	for k, index := range m.fields {
		value, ok := src[k]
		if !ok {
			continue
		}

		fv := ov.FieldByIndex(index)
		if fv.IsValid() && fv.CanSet() {
			// some types requires special conversion
			v, e := claimUnmarshalConvert(reflect.ValueOf(value), fv.Type())
			if e != nil {
				return e
			}
			fv.Set(v)
		}
	}

	// try set internal Claims interfaces
	err,_ := m.aggregateEmbeddedClaims(owner, nil, func(i interface{}, claims Claims) (interface{}, bool) {
		if sc, ok := claims.(claimsMapper); ok {
			e := sc.fromMap(sc, src)
			if e != nil {
				return e, false
			}
		} else if mc, ok := claims.(MapClaims); ok {
			e := mc.fromMap(src)
			if e != nil {
				return e, false
			}
		}
		return nil, true
	}).(error)

	return err
}

func (m *FieldClaimsMapper) aggregateEmbeddedClaims(owner interface{}, initial interface{}, accumulator accumulator) interface{} {
	m.prepare(owner)
	if len(m.interfaces) == 0 {
		return initial
	}

	v := reflect.ValueOf(owner)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	i := initial
	next := true
	for _,index := range m.interfaces {
		fv := v.FieldByIndex(index)
		if !fv.IsValid() || fv.IsZero() {
			continue
		}
		if claims, ok := fv.Interface().(Claims); ok {
			i, next = accumulator(i, claims)
			if !next {
				break
			}
		}
	}
	return i
}

func (m *FieldClaimsMapper) prepare(owner interface{}) {
	if m.fields != nil {
		return
	}

	t := reflect.TypeOf(owner)
	m.fields = map[string][]int{}
	m.populateFieldMap(t, []int{})
	m.interfaces = [][]int{}
	m.populateInterfaceList(t)
}

// populateFieldMap recursively map fields of given struct type with its claim value, take embedded filed into consideration
func (m *FieldClaimsMapper) populateFieldMap(structType reflect.Type, index []int) {
	t := structType
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct || mapperType.AssignableTo(t){
		return
	}

	total := t.NumField()
	for i := 0; i < total; i++ {
		field := t.Field(i)

		if field.Anonymous {
			m.populateFieldMap(field.Type, append(index, field.Index...))
			continue
		}

		if claim, ok := field.Tag.Lookup(ClaimTag); ok {
			m.fields[claim] = append(index, field.Index...)
		}
	}
}

// populateInterfaceList find all fields with Claims interface as a type
func (m *FieldClaimsMapper) populateInterfaceList(structType reflect.Type) {
	t := structType
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct || mapperType.AssignableTo(t){
		return
	}

	total := t.NumField()
	for i := 0; i < total; i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Interface && claimsType.AssignableTo(field.Type) || field.Type.AssignableTo(mapClaimsType) {
			m.interfaces = append(m.interfaces, field.Index)
		}
	}
}
