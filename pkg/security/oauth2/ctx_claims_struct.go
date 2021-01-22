package oauth2

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

/***************************
	Struct Claims Helpers
 ***************************/
var (
	mapperType = reflect.TypeOf(StructClaimsMapper{})
	timeType   = reflect.TypeOf(time.Time{})
	int64Type    = reflect.TypeOf(int64(0))
)

// StructClaimsMapper is a helper type that can be embedded into struct based claims
type StructClaimsMapper struct {
	fieldIndex map[string][]int
}

func (m *StructClaimsMapper) Get(owner interface{}, claim string) interface{} {
	v := m.findFieldValue(owner, claim)
	if v.IsValid() && !v.IsZero() {
		return v.Interface()
	} else {
		return nil
	}
}

func (m *StructClaimsMapper) Has(owner interface{}, claim string) bool {
	v := m.findFieldValue(owner, claim)
	return v.IsValid() && !v.IsZero()
}

func (m *StructClaimsMapper) Set(owner interface{}, claim string, value interface{}) {
	v := m.findFieldValue(owner, claim)
	if v.IsValid() {
		if e := m.set(v, value); e != nil {
			panic(e)
		}
	}
}

func (c *StructClaimsMapper) DoMarshalJSON(owner interface{}) ([]byte, error) {
	v, e := c.toMap(owner)
	if e != nil {
		return nil, e
	}
	return json.Marshal(v)
}

func (c *StructClaimsMapper) DoUnmarshalJSON(owner interface{}, bytes []byte) error {
	m := map[string]interface{}{}
	if e := json.Unmarshal(bytes, &m); e != nil {
		return e
	}

	if e := c.fromMap(owner, m); e != nil {
		return e
	}
	return nil
}

func (m *StructClaimsMapper) findFieldValue(owner interface{}, claim string) (ret reflect.Value) {
	m.prepare(owner)
	index, ok := m.fieldIndex[claim]
	if !ok {
		return
	}

	v := reflect.ValueOf(owner)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.FieldByIndex(index)
}

func (m *StructClaimsMapper) set(fv reflect.Value, setTo interface{}) error {
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

func (m *StructClaimsMapper) toMap(owner interface{}) (map[string]interface{}, error) {
	m.prepare(owner)
	ov := reflect.ValueOf(owner)
	if ov.Kind() == reflect.Ptr {
		ov = ov.Elem()
	}

	ret := map[string]interface{}{}
	for k, index := range m.fieldIndex {
		fv := ov.FieldByIndex(index)
		if fv.IsValid() && !fv.IsZero() {
			v, e := marshalConvert(fv)
			if e != nil {
				return nil, e
			}
			ret[k] = v.Interface()
		}
	}
	return ret, nil
}

func (m *StructClaimsMapper) fromMap(owner interface{}, src map[string]interface{}) error {
	m.prepare(owner)
	ov := reflect.ValueOf(owner)
	if ov.Kind() == reflect.Ptr {
		ov = ov.Elem()
	}

	for k, setTo := range src {
		index, ok := m.fieldIndex[k]
		if !ok {
			continue
		}

		fv := ov.FieldByIndex(index)
		if fv.IsValid() && fv.CanSet() {
			// some types requires special conversion
			v, e := unmarshalConvert(reflect.ValueOf(setTo), fv.Type())
			if e != nil {
				return e
			}
			fv.Set(v)
		}
	}
	return nil
}

func (m *StructClaimsMapper) prepare(owner interface{}) {
	if m.fieldIndex != nil {
		return
	}

	m.fieldIndex = map[string][]int{}
	m.populateFieldMap(reflect.TypeOf(owner), []int{})
}

// populateFieldMap recursively map fields of given struct type with its claim value, take embedded filed into consideration
func (m *StructClaimsMapper) populateFieldMap(structType reflect.Type, index []int) {
	t := structType
	if t.Kind() == reflect.Interface {
		t = t.Elem()
	}

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
			m.fieldIndex[claim] = append(index, field.Index...)
		}
	}
}

// some conversions
func marshalConvert(v reflect.Value) (reflect.Value, error) {
	t := v.Type()
	switch {
	case  timeType.AssignableTo(t):
		return timeToTimestamp(v)
	default:
		return v, nil
	}
}

func unmarshalConvert(v reflect.Value, fieldType reflect.Type) (reflect.Value, error) {
	switch {
	case  timeType.AssignableTo(fieldType):
		return timestampToTime(v)
	default:
		return v, nil
	}
}

func timestampToTime(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(int64Type):
		timestamp := v.Convert(int64Type).Interface().(int64)
		return reflect.ValueOf(time.Unix(timestamp, 0)), nil
	case v.Type().ConvertibleTo(timeType):
		return v.Convert(timeType), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to time.Time", v.Interface())
	}
}

func timeToTimestamp(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(timeType):
		time := v.Convert(timeType).Interface().(time.Time)
		return reflect.ValueOf(time.Unix()), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to timestamp", v.Interface())
	}
}
