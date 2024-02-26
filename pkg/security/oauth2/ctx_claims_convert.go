// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "reflect"
    "time"
)

var (
	stringType    = reflect.TypeOf("")
	timeType      = reflect.TypeOf(time.Time{})
	int64Type     = reflect.TypeOf(int64(0))
	float64Type   = reflect.TypeOf(float64(0))
	float32Type   = reflect.TypeOf(float32(0))
	sSliceType    = reflect.TypeOf([]string{})
	iSliceType    = reflect.TypeOf([]interface{}{})
	sSetType      = reflect.TypeOf(utils.NewStringSet())
	sSetClaimType = reflect.TypeOf(StringSetClaim(utils.NewStringSet()))
	iSetType      = reflect.TypeOf(utils.NewSet())
	mapType       = reflect.TypeOf(map[string]interface{}{})
	anyType       = reflect.TypeOf(interface{}(0))
)

// some conversions
func claimMarshalConvert(v reflect.Value) (reflect.Value, error) {
	t := v.Type()
	switch {
	case timeType.AssignableTo(t):
		return timeToTimestamp(v)
	case float64Type.AssignableTo(t):
		fallthrough
	case float32Type.AssignableTo(t):
		return v.Convert(int64Type), nil
	default:
		return v, nil
	}
}

func claimUnmarshalConvert(v reflect.Value, fieldType reflect.Type) (reflect.Value, error) {
	switch {
	// special target types
	case timeType.AssignableTo(fieldType):
		return timestampToTime(v)
	case sSetClaimType.AssignableTo(fieldType):
		return toStringSetClaim(v)
	case sSetType.AssignableTo(fieldType):
		return toStringSet(v)
	case iSetType.AssignableTo(fieldType):
		return toSet(v)
	case fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() != reflect.Struct:
		return toAddr(v)
	case v.Type().AssignableTo(mapType) && isStructOrStructPtr(fieldType):
		return mapToStruct(v, fieldType)
	case sSliceType.AssignableTo(fieldType):
		return toStringSlice(v)

	// special source types
	case v.Type().AssignableTo(float32Type):
		fallthrough
	case v.Type().AssignableTo(float64Type):
		v = v.Convert(int64Type)
	}

	switch {
	// convertable and assignable
	case v.Type().AssignableTo(fieldType):
		return v, nil
	case v.Type().ConvertibleTo(fieldType):
		return v.Convert(fieldType), nil
	default:
		return v, nil
	}
}

func isStructOrStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Struct || t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
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

func toStringSet(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(sSliceType):
		slice := v.Convert(sSliceType).Interface().([]string)
		return reflect.ValueOf(utils.NewStringSet(slice...)), nil
	case v.Type().ConvertibleTo(iSliceType):
		slice := v.Convert(iSliceType).Interface().([]interface{})
		set := utils.NewStringSetFromSet(utils.NewSet(slice...))
		return reflect.ValueOf(set), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to string set", v.Interface())
	}
}

func toStringSetClaim(v reflect.Value) (reflect.Value, error) {
	var set utils.StringSet
	switch {
	case v.Type().ConvertibleTo(stringType):
		str := v.Convert(stringType).Interface().(string)
		set = utils.NewStringSet(str)
	case v.Type().ConvertibleTo(sSliceType):
		slice := v.Convert(sSliceType).Interface().([]string)
		set = utils.NewStringSet(slice...)
	case v.Type().ConvertibleTo(iSliceType):
		slice := v.Convert(iSliceType).Interface().([]interface{})
		set = utils.NewStringSetFromSet(utils.NewSet(slice...))
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to string set", v.Interface())
	}
	return reflect.ValueOf(StringSetClaim(set)), nil
}

func toSet(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(sSliceType):
		slice := v.Convert(sSliceType).Interface().([]string)
		return reflect.ValueOf(utils.NewStringSet(slice...).ToSet()), nil
	case v.Type().ConvertibleTo(iSliceType):
		slice := v.Convert(iSliceType).Interface().([]interface{})
		return reflect.ValueOf(utils.NewSet(slice...)), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to set", v.Interface())
	}
}

func toAddr(v reflect.Value) (reflect.Value, error) {
	if v.CanAddr() {
		return v.Addr(), nil
	}

	switch v.Kind() {
	case reflect.Bool:
		return reflect.ValueOf(utils.BoolPtr(v.Bool())), nil
	case reflect.Int:
		return reflect.ValueOf(utils.IntPtr(int(v.Int()))), nil
	case reflect.Uint:
		return reflect.ValueOf(utils.UIntPtr(uint(v.Uint()))), nil
	case reflect.Float64:
		return reflect.ValueOf(utils.Float64Ptr(v.Float())), nil
	default:
		return reflect.Value{}, fmt.Errorf("value [%v, %T] cannot be addressed", v.Interface(), v.Interface())
	}
}

func toStringSlice(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(sSliceType):
		return v.Convert(sSliceType), nil
	case v.Type().ConvertibleTo(iSliceType):
		srcSlice := v.Convert(iSliceType)
		slice := reflect.MakeSlice(sSliceType, srcSlice.Len(), srcSlice.Len())
		for i := 0; i < srcSlice.Len(); i++ {
			elem := srcSlice.Index(i).Elem()
			if !elem.Type().ConvertibleTo(stringType) {
				return reflect.Value{}, fmt.Errorf("type %T cannot be converted to []string, source contains non-string type %T", v.Interface(), elem.Interface())
			}
			slice.Index(i).Set(elem.Convert(stringType))
		}
		return slice, nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to []string", v.Interface())
	}
}

func mapToStruct(v reflect.Value, ft reflect.Type) (reflect.Value, error) {
	isPtr := false
	if ft.Kind() == reflect.Ptr {
		isPtr = true
		ft = ft.Elem()
	}

	if ft.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("map can only convert to struct or pointer of struct. got [%T]", ft.String())
	}

	// first instantiate
	nv := reflect.New(ft)

	// try convert
	// instead of reflection, we use JSON to do the convert. This is much slower but safer
	m := v.Interface()
	data, e := json.Marshal(m)
	if e != nil {
		return reflect.Value{}, fmt.Errorf("map cannot be serialized to json: %v", e)
	}

	if e := json.Unmarshal(data, nv.Interface()); e != nil {
		return reflect.Value{}, fmt.Errorf("json cannot be converted: %v", e)
	}

	if !isPtr {
		nv = nv.Elem()
	}
	return nv, nil
}
