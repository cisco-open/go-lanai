package utils

import (
	"encoding/json"
	"reflect"
	"strings"
)

// ConvertSlice attempt to convert []interface{} to []elementType using the first element's type.
// if given slice is empty, or any elements is not the same type of first one, same slice is returned
func ConvertSlice(slice []interface{}) interface{} {
	if len(slice) == 0 {
		return slice
	}

	var success bool
	vSlice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(slice[0])), len(slice), len(slice))
	for i, v := range slice {
		rv := reflect.ValueOf(v)
		ev := vSlice.Index(i)
		if !rv.Type().ConvertibleTo(ev.Type()) {
			success = false
			break
		}
		ev.Set(rv.Convert(ev.Type()))
	}
	if success {
		return vSlice.Interface()
	}
	return slice
}

// CommaSeparatedSlice alias of []string that can deserialize from comma delimited string
type CommaSeparatedSlice []string

// fmt.Stringer
func (s CommaSeparatedSlice) String() string {
	return strings.Join(s, ", ")
}

// MarshalText encoding.TextMarshaler
func (s CommaSeparatedSlice) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText encoding.TextUnmarshaler
func (s *CommaSeparatedSlice) UnmarshalText(data []byte) error {
	if string(data) == "" {
		return nil
	}
	var result []string
	split := strings.Split(string(data), ",")
	for _, s := range split {
		s = strings.TrimSpace(s)
		result = append(result, s)
	}
	*s = result
	return nil
}

// UnmarshalJSON json.Unmarshaler
func (s *CommaSeparatedSlice) UnmarshalJSON(data []byte) error {
	// first try regular array
	var slice []string
	if e := json.Unmarshal(data, &slice); e == nil {
		*s = slice
		return nil
	}

	// try comma separated format
	var str string
	if e := json.Unmarshal(data, &str); e != nil {
		return e
	}
	return s.UnmarshalText([]byte(str))
}
