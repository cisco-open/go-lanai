package utils

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Reverse will reverse the order of the given slice
func Reverse[T any](input []T) {
	for i, j := 0, len(input)-1; i < j; i, j = i+1, j-1 {
		input[i], input[j] = input[j], input[i]
	}
}

// RemoveStable will remove an element from the slice and keep its order. This
// operation can be potentially costly depending on how large the slice is since
// it needs to shift all elements that appear after index i over by 1.
//
// This function will automatically type itself using type inference.
//
// If the given index is not within the bounds of the slice, then the function will
// panic
func RemoveStable[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		panic("invalid slice index")
	}
	return append(slice[:index], slice[index+1:]...)
}

// Remove will not keep the ordering of the slice. It has a very fast operation.
// This function will automatically type itself using type inference
//
//		intSlice := []int{1, 2, 3, 4}
//		intSlice = Remove(intSlice, 1)
//	 	result: {1, 4, 3}
func Remove[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		panic("invalid slice index")
	}
	slice[index] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

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
