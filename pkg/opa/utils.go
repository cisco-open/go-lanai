package opa

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// marshalMergedJSON merge extra into v, v have to be struct or map.
// "processors" are invoked after merge and before marshalling
func marshalMergedJSON(obj interface{}, extra map[string]interface{}, processors ...func(m map[string]interface{})) ([]byte, error) {
	data, e := json.Marshal(obj)
	if len(extra) == 0 && len(processors) == 0 || e != nil {
		return data, e
	}
	// merge extra
	var m map[string]interface{}
	if e := json.Unmarshal(data, &m); e != nil {
		return nil, fmt.Errorf("unable to merge JSON: %v", e)
	}
	for k, v := range extra {
		m[k] = v
	}
	for _, fn := range processors {
		fn(m)
	}
	return json.Marshal(m)
}

// minimizeMap recursively remove any zero valued entries
func minimizeMap(m map[string]interface{}) {
	minimize(reflect.ValueOf(m))
}

func minimize(rv reflect.Value) (minimized reflect.Value, isZero bool) {
	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	isZero = !rv.IsValid() || rv.IsZero()
	switch rv.Kind() {
	case reflect.Map:
		rv = minimizeMapValue(rv)
		isZero = isZero || rv.Len() == 0
	case reflect.Slice:
		rv = minimizeSliceValue(rv)
		isZero = isZero || rv.Len() == 0
	}
	return rv, isZero
}

func minimizeMapValue(mapV reflect.Value) reflect.Value {
	for _, k := range mapV.MapKeys() {
		v := mapV.MapIndex(k)
		v, zero := minimize(v)
		if zero {
			mapV.SetMapIndex(k, reflect.Value{})
		} else {
			mapV.SetMapIndex(k, v)
		}
	}
	return mapV
}

func minimizeSliceValue(sliceV reflect.Value) reflect.Value {
	newV := reflect.MakeSlice(sliceV.Type().Elem(), 0, sliceV.Len())
	for i:=0; i < sliceV.Len(); i ++ {
		v := sliceV.Index(i)
		if v, zero := minimize(v); !zero {
			newV = reflect.Append(newV, v)
		}
	}
	return newV
}

